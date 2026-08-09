# MobileLabKit

Optional Swift hooks for native iOS applications using MobileLab protocol v1. The API Sandbox works without this package; add it only for lifecycle events, markers, and scenario assertions.

Extract `mobilelab-kit-0.7.0.tar.gz` from the v0.7 GitHub Release and add that directory as a local Swift Package, then configure a URL reachable by the Simulator or device:

```swift
let client = try MobileLabClient(endpoint: URL(string: "http://127.0.0.1:4566")!)
let lifecycle = MobileLabUIKitLifecycleReporter(client: client)
lifecycle.attach()

try await client.marker("checkout.loaded", attributes: ["screen": "checkout"])
try await client.assertThat("cart.total", passed: cart.total == expectedTotal)
```

`MobileLabUIKitLifecycleReporter` observes foreground/background notifications and suppresses duplicate states. The transport uses `URLSession`, has no runtime dependency, and can be replaced through `MobileLabTransport` in tests.

Simulator loopback reaches the Mac directly. A physical device needs a host address reachable on the same network and an explicit non-loopback MobileLab binding; review the security warning before exposing the local sandbox. Clear-text HTTP may also require a development-only App Transport Security exception for non-loopback hosts.

Run `swift test` from this directory.
