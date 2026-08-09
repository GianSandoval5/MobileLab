import MobileLabKit
import SwiftUI

@main
struct ExampleApp: App {
    private let client: MobileLabClient
    private let lifecycle: MobileLabUIKitLifecycleReporter

    init() {
        let client = try! MobileLabClient(endpoint: URL(string: "http://127.0.0.1:4566")!)
        self.client = client
        lifecycle = MobileLabUIKitLifecycleReporter(client: client)
        lifecycle.attach()
    }

    var body: some Scene {
        WindowGroup {
            Button("Report marker") {
                Task { try? await client.marker("example.clicked") }
            }
        }
    }
}
