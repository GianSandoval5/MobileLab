import Foundation
#if canImport(FoundationNetworking)
import FoundationNetworking
#endif

public let mobileLabProtocolVersion = 1
public let mobileLabKitVersion = "1.1.0"

public enum MobileLabEventKind: String, Codable, Sendable {
    case lifecycle
    case marker
    case assertion
}

public enum MobileLabValue: Codable, Sendable, Equatable {
    case string(String)
    case bool(Bool)
    case integer(Int64)
    case number(Double)
    case object([String: MobileLabValue])
    case array([MobileLabValue])
    case null

    public init(from decoder: Decoder) throws {
        let value = try decoder.singleValueContainer()
        if value.decodeNil() { self = .null }
        else if let decoded = try? value.decode(Bool.self) { self = .bool(decoded) }
        else if let decoded = try? value.decode(Int64.self) { self = .integer(decoded) }
        else if let decoded = try? value.decode(Double.self) { self = .number(decoded) }
        else if let decoded = try? value.decode(String.self) { self = .string(decoded) }
        else if let decoded = try? value.decode([String: MobileLabValue].self) { self = .object(decoded) }
        else if let decoded = try? value.decode([MobileLabValue].self) { self = .array(decoded) }
        else { throw DecodingError.dataCorruptedError(in: value, debugDescription: "Unsupported MobileLab value") }
    }

    public func encode(to encoder: Encoder) throws {
        var value = encoder.singleValueContainer()
        switch self {
        case let .string(item): try value.encode(item)
        case let .bool(item): try value.encode(item)
        case let .integer(item): try value.encode(item)
        case let .number(item): try value.encode(item)
        case let .object(item): try value.encode(item)
        case let .array(item): try value.encode(item)
        case .null: try value.encodeNil()
        }
    }
}

extension MobileLabValue: ExpressibleByStringLiteral {
    public init(stringLiteral value: String) { self = .string(value) }
}

extension MobileLabValue: ExpressibleByBooleanLiteral {
    public init(booleanLiteral value: Bool) { self = .bool(value) }
}

extension MobileLabValue: ExpressibleByIntegerLiteral {
    public init(integerLiteral value: Int64) { self = .integer(value) }
}

extension MobileLabValue: ExpressibleByFloatLiteral {
    public init(floatLiteral value: Double) { self = .number(value) }
}

extension MobileLabValue: ExpressibleByArrayLiteral {
    public init(arrayLiteral elements: MobileLabValue...) { self = .array(elements) }
}

extension MobileLabValue: ExpressibleByDictionaryLiteral {
    public init(dictionaryLiteral elements: (String, MobileLabValue)...) {
        self = .object(Dictionary(uniqueKeysWithValues: elements))
    }
}

extension MobileLabValue: ExpressibleByNilLiteral {
    public init(nilLiteral: ()) { self = .null }
}

public struct MobileLabEvent: Codable, Sendable, Equatable {
    public let protocolVersion: Int
    public let framework: String
    public let kind: MobileLabEventKind
    public let name: String
    public let passed: Bool?
    public let sessionId: String?
    public let attributes: [String: MobileLabValue]?

    public init(
        kind: MobileLabEventKind,
        name: String,
        passed: Bool? = nil,
        sessionId: String? = nil,
        attributes: [String: MobileLabValue] = [:]
    ) {
        protocolVersion = mobileLabProtocolVersion
        framework = "ios"
        self.kind = kind
        self.name = name
        self.passed = passed
        self.sessionId = sessionId
        self.attributes = attributes.isEmpty ? nil : attributes
    }
}

public protocol MobileLabTransport: Sendable {
    func send(_ event: MobileLabEvent) async throws
}

public final class MobileLabHTTPTransport: MobileLabTransport, @unchecked Sendable {
    private let eventsURL: URL
    private let timeout: TimeInterval
    private let session: URLSession

    public init(endpoint: URL, timeout: TimeInterval = 3, session: URLSession = .shared) throws {
        guard let scheme = endpoint.scheme?.lowercased(), scheme == "http" || scheme == "https" else {
            throw MobileLabError.invalidEndpoint
        }
        guard timeout > 0 else { throw MobileLabError.invalidTimeout }
        var components = URLComponents(url: endpoint, resolvingAgainstBaseURL: false)
        components?.path = "/__mobilelab/sdk/events"
        components?.query = nil
        components?.fragment = nil
        guard let eventsURL = components?.url else { throw MobileLabError.invalidEndpoint }
        self.eventsURL = eventsURL
        self.timeout = timeout
        self.session = session
    }

    public func send(_ event: MobileLabEvent) async throws {
        var request = URLRequest(url: eventsURL, timeoutInterval: timeout)
        request.httpMethod = "POST"
        request.setValue("application/json", forHTTPHeaderField: "Content-Type")
        request.setValue("MobileLabKit/\(mobileLabKitVersion)", forHTTPHeaderField: "X-MobileLab-SDK")
        request.httpBody = try JSONEncoder().encode(event)
        let (_, response) = try await session.data(for: request)
        guard let http = response as? HTTPURLResponse else { throw MobileLabError.invalidResponse }
        guard (200...299).contains(http.statusCode) else { throw MobileLabError.httpStatus(http.statusCode) }
    }
}

public struct MobileLabClient: Sendable {
    private let sessionId: String?
    private let transport: any MobileLabTransport

    public init(endpoint: URL, sessionId: String? = nil, timeout: TimeInterval = 3) throws {
        self.init(sessionId: sessionId, transport: try MobileLabHTTPTransport(endpoint: endpoint, timeout: timeout))
    }

    public init(sessionId: String? = nil, transport: any MobileLabTransport) {
        self.sessionId = sessionId
        self.transport = transport
    }

    public func lifecycle(_ name: String, attributes: [String: MobileLabValue] = [:]) async throws {
        try await send(kind: .lifecycle, name: name, attributes: attributes)
    }

    public func marker(_ name: String, attributes: [String: MobileLabValue] = [:]) async throws {
        try await send(kind: .marker, name: name, attributes: attributes)
    }

    public func assertThat(_ name: String, passed: Bool, attributes: [String: MobileLabValue] = [:]) async throws {
        try await send(kind: .assertion, name: name, passed: passed, attributes: attributes)
    }

    private func send(
        kind: MobileLabEventKind,
        name: String,
        passed: Bool? = nil,
        attributes: [String: MobileLabValue]
    ) async throws {
        try await transport.send(
            MobileLabEvent(kind: kind, name: name, passed: passed, sessionId: sessionId, attributes: attributes)
        )
    }
}

public actor MobileLabLifecycleReporter {
    private let client: MobileLabClient
    private var lastLifecycle: String?

    public init(client: MobileLabClient) {
        self.client = client
    }

    public func reportReady() async throws {
        try await client.lifecycle("ready")
    }

    public func onForeground() async throws {
        try await reportOnce("foreground")
    }

    public func onBackground() async throws {
        try await reportOnce("background")
    }

    private func reportOnce(_ lifecycle: String) async throws {
        guard lastLifecycle != lifecycle else { return }
        lastLifecycle = lifecycle
        try await client.lifecycle(lifecycle)
    }
}

public enum MobileLabError: Error, Equatable {
    case invalidEndpoint
    case invalidTimeout
    case invalidResponse
    case httpStatus(Int)
}
