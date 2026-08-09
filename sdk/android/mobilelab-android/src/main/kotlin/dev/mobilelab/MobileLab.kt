package dev.mobilelab

import java.io.IOException
import java.net.HttpURLConnection
import java.net.URI
import java.util.concurrent.Executor
import java.util.concurrent.Executors

public const val MOBILELAB_PROTOCOL_VERSION: Int = 1

public enum class MobileLabEventKind(public val wireName: String) {
    LIFECYCLE("lifecycle"),
    MARKER("marker"),
    ASSERTION("assertion"),
}

public data class MobileLabEvent(
    val protocolVersion: Int = MOBILELAB_PROTOCOL_VERSION,
    val framework: String = "android",
    val kind: MobileLabEventKind,
    val name: String,
    val passed: Boolean? = null,
    val sessionId: String? = null,
    val attributes: Map<String, Any?> = emptyMap(),
)

public fun interface MobileLabTransport {
    @Throws(IOException::class)
    public fun send(event: MobileLabEvent)
}

public class MobileLabHttpTransport(
    endpoint: String,
    private val timeoutMillis: Int = 3_000,
) : MobileLabTransport {
    private val eventsUri: URI = URI(endpoint).resolve("/__mobilelab/sdk/events")

    init {
        require(eventsUri.scheme == "http" || eventsUri.scheme == "https") {
            "MobileLab endpoint must use http or https"
        }
        require(timeoutMillis > 0) { "timeoutMillis must be greater than zero" }
    }

    @Throws(IOException::class)
    override fun send(event: MobileLabEvent) {
        val connection = eventsUri.toURL().openConnection() as HttpURLConnection
        try {
            connection.requestMethod = "POST"
            connection.connectTimeout = timeoutMillis
            connection.readTimeout = timeoutMillis
            connection.doOutput = true
            connection.setRequestProperty("Content-Type", "application/json")
            connection.setRequestProperty("X-MobileLab-SDK", "mobilelab-android/1.0.0")
            connection.outputStream.use { output ->
                output.write(Json.encode(event.toWireValue()).toByteArray(Charsets.UTF_8))
            }
            if (connection.responseCode !in 200..299) {
                throw MobileLabException("SDK bridge returned HTTP ${connection.responseCode}")
            }
        } catch (error: MobileLabException) {
            throw error
        } catch (error: IOException) {
            throw MobileLabException("SDK bridge is unreachable at $eventsUri", error)
        } finally {
            connection.disconnect()
        }
    }
}

public class MobileLabClient(
    endpoint: String,
    private val sessionId: String? = null,
    private val transport: MobileLabTransport = MobileLabHttpTransport(endpoint),
) {
    @Throws(IOException::class)
    public fun lifecycle(name: String, attributes: Map<String, Any?> = emptyMap()): Unit =
        send(MobileLabEventKind.LIFECYCLE, name, null, attributes)

    @Throws(IOException::class)
    public fun marker(name: String, attributes: Map<String, Any?> = emptyMap()): Unit =
        send(MobileLabEventKind.MARKER, name, null, attributes)

    @Throws(IOException::class)
    public fun assertThat(name: String, passed: Boolean, attributes: Map<String, Any?> = emptyMap()): Unit =
        send(MobileLabEventKind.ASSERTION, name, passed, attributes)

    @Throws(IOException::class)
    private fun send(kind: MobileLabEventKind, name: String, passed: Boolean?, attributes: Map<String, Any?>) {
        transport.send(
            MobileLabEvent(
                kind = kind,
                name = name,
                passed = passed,
                sessionId = sessionId,
                attributes = attributes.toMap(),
            ),
        )
    }
}

public class MobileLabLifecycleReporter(
    private val client: MobileLabClient,
    private val executor: Executor = Executors.newSingleThreadExecutor { operation ->
        Thread(operation, "mobilelab-lifecycle").apply { isDaemon = true }
    },
    private val onError: (Throwable) -> Unit = {},
) {
    private var lastLifecycle: String? = null

    public fun reportReady(): Unit = report { client.lifecycle("ready") }

    @Synchronized
    public fun onForeground() {
        reportOnce("foreground")
    }

    @Synchronized
    public fun onBackground() {
        reportOnce("background")
    }

    private fun reportOnce(lifecycle: String) {
        if (lastLifecycle == lifecycle) return
        lastLifecycle = lifecycle
        report { client.lifecycle(lifecycle) }
    }

    private fun report(operation: () -> Unit) {
        executor.execute {
            try {
                operation()
            } catch (error: Throwable) {
                onError(error)
            }
        }
    }
}

public class MobileLabException(message: String, cause: Throwable? = null) : IOException(message, cause)

private fun MobileLabEvent.toWireValue(): Map<String, Any?> = buildMap {
    put("protocolVersion", protocolVersion)
    put("framework", framework)
    put("kind", kind.wireName)
    put("name", name)
    passed?.let { put("passed", it) }
    sessionId?.takeIf { it.isNotEmpty() }?.let { put("sessionId", it) }
    attributes.takeIf { it.isNotEmpty() }?.let { put("attributes", it) }
}

private object Json {
    fun encode(value: Any?): String = when (value) {
        null -> "null"
        is String -> encodeString(value)
        is Boolean -> value.toString()
        is Byte, is Short, is Int, is Long -> value.toString()
        is Float -> encodeFloating(value.toDouble())
        is Double -> encodeFloating(value)
        is Map<*, *> -> value.entries.joinToString(prefix = "{", postfix = "}") { (key, item) ->
            require(key is String) { "MobileLab attribute keys must be strings" }
            "${encodeString(key)}:${encode(item)}"
        }
        is Iterable<*> -> value.joinToString(prefix = "[", postfix = "]") { encode(it) }
        is Array<*> -> value.joinToString(prefix = "[", postfix = "]") { encode(it) }
        else -> throw IllegalArgumentException("Unsupported MobileLab attribute type: ${value::class.java.name}")
    }

    private fun encodeFloating(value: Double): String {
        require(value.isFinite()) { "MobileLab numeric attributes must be finite" }
        return value.toString()
    }

    private fun encodeString(value: String): String = buildString(value.length + 2) {
        append('"')
        value.forEach { character ->
            when (character) {
                '"' -> append("\\\"")
                '\\' -> append("\\\\")
                '\b' -> append("\\b")
                '\u000C' -> append("\\f")
                '\n' -> append("\\n")
                '\r' -> append("\\r")
                '\t' -> append("\\t")
                else -> if (character < ' ') {
                    append("\\u")
                    append(character.code.toString(16).padStart(4, '0'))
                } else {
                    append(character)
                }
            }
        }
        append('"')
    }
}
