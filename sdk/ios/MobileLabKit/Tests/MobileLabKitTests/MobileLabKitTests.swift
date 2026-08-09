import Foundation
import Testing
@testable import MobileLabKit

private actor RecordingTransport: MobileLabTransport {
    private(set) var events: [MobileLabEvent] = []

    func send(_ event: MobileLabEvent) async throws {
        events.append(event)
    }
}

@Test func clientEmitsIOSProtocolEvents() async throws {
    let transport = RecordingTransport()
    let client = MobileLabClient(sessionId: "run-1", transport: transport)

    try await client.marker("checkout.loaded", attributes: ["screen": "checkout"])
    try await client.assertThat("cart.total", passed: true)

    let events = await transport.events
    #expect(events.count == 2)
    #expect(events[0].protocolVersion == mobileLabProtocolVersion)
    #expect(events[0].framework == "ios")
    #expect(events[0].sessionId == "run-1")
    #expect(events[1].passed == true)
}

@Test func lifecycleReporterDeduplicatesState() async throws {
    let transport = RecordingTransport()
    let reporter = MobileLabLifecycleReporter(client: MobileLabClient(transport: transport))

    try await reporter.onBackground()
    try await reporter.onBackground()
    try await reporter.onForeground()

    let events = await transport.events
    #expect(events.map(\.name) == ["background", "foreground"])
}

@Test func eventEncodesProtocolWireShape() throws {
    let event = MobileLabEvent(
        kind: .marker,
        name: "transport.ready",
        attributes: ["nested": ["enabled": true], "items": [1, "two"]]
    )
    let object = try #require(JSONSerialization.jsonObject(with: JSONEncoder().encode(event)) as? [String: Any])

    #expect(object["protocolVersion"] as? Int == 1)
    #expect(object["framework"] as? String == "ios")
    #expect(object["name"] as? String == "transport.ready")
}
