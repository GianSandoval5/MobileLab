# mobilelab-android

Optional Kotlin hooks for native Android applications using MobileLab protocol v1. The API Sandbox works without this package; add it only for lifecycle events, markers, and scenario assertions.

The artifact is a Kotlin/JVM library with no AndroidX or serialization runtime dependency. Its networking uses `HttpURLConnection`, which is available on Android. Call direct client methods away from the main thread; `MobileLabLifecycleReporter` dispatches lifecycle calls on its own executor.

The v0.5 GitHub Release includes `mobilelab-android-0.5.0.jar`. Place it under the application's `app/libs` directory and add `implementation(files("libs/mobilelab-android-0.5.0.jar"))`; the JAR targets JVM bytecode 8 and expects the application's normal Kotlin standard library.

```kotlin
val mobileLab = MobileLabClient(endpoint = "http://10.0.2.2:4566")
val lifecycle = MobileLabLifecycleReporter(mobileLab)

override fun onStart(owner: LifecycleOwner) = lifecycle.onForeground()
override fun onStop(owner: LifecycleOwner) = lifecycle.onBackground()

// From a coroutine IO context or worker thread:
mobileLab.marker("checkout.loaded", mapOf("screen" to "checkout"))
mobileLab.assertThat("cart.total", cart.total == expectedTotal)
```

Use `10.0.2.2` for the host from the standard Android Emulator. A physical device needs an address reachable from the device, or `adb reverse tcp:4566 tcp:4566` with a loopback endpoint.

Build and test with `./gradlew test`; the release JAR is produced by `./gradlew jar`.
