package dev.mobilelab

import com.sun.net.httpserver.HttpServer
import java.net.InetSocketAddress
import java.util.concurrent.Executor
import kotlin.test.Test
import kotlin.test.assertContains
import kotlin.test.assertEquals

class MobileLabTest {
    @Test
    fun `client emits Android protocol events`() {
        val events = mutableListOf<MobileLabEvent>()
        val client = MobileLabClient(
            endpoint = "http://10.0.2.2:4566",
            sessionId = "run-1",
            transport = MobileLabTransport(events::add),
        )

        client.marker("checkout.loaded", mapOf("screen" to "checkout"))
        client.assertThat("cart.total", true)

        assertEquals(2, events.size)
        assertEquals(MOBILELAB_PROTOCOL_VERSION, events[0].protocolVersion)
        assertEquals("android", events[0].framework)
        assertEquals("run-1", events[0].sessionId)
        assertEquals(true, events[1].passed)
    }

    @Test
    fun `lifecycle reporter deduplicates state`() {
        val events = mutableListOf<MobileLabEvent>()
        val directExecutor = Executor { operation -> operation.run() }
        val client = MobileLabClient(
            endpoint = "http://127.0.0.1:4566",
            transport = MobileLabTransport(events::add),
        )
        val reporter = MobileLabLifecycleReporter(client, directExecutor)

        reporter.onBackground()
        reporter.onBackground()
        reporter.onForeground()

        assertEquals(listOf("background", "foreground"), events.map(MobileLabEvent::name))
    }

    @Test
    fun `HTTP transport posts JSON to SDK bridge`() {
        var body = ""
        var sdkHeader = ""
        val server = HttpServer.create(InetSocketAddress("127.0.0.1", 0), 0)
        server.createContext("/__mobilelab/sdk/events") { exchange ->
            body = exchange.requestBody.bufferedReader().use { it.readText() }
            sdkHeader = exchange.requestHeaders.getFirst("X-MobileLab-SDK")
            exchange.sendResponseHeaders(202, -1)
            exchange.close()
        }
        server.start()
        try {
            val client = MobileLabClient("http://127.0.0.1:${server.address.port}/nested/path")
            client.marker("transport.ready", mapOf("quoted" to "a\"b", "items" to listOf(1, true)))
        } finally {
            server.stop(0)
        }

        assertEquals("mobilelab-android/0.4.0", sdkHeader)
        assertContains(body, "\"framework\":\"android\"")
        assertContains(body, "\"name\":\"transport.ready\"")
        assertContains(body, "\"quoted\":\"a\\\"b\"")
    }
}
