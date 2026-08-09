# MobileLab Architecture

MobileLab is a local-first mobile development environment. Its architectural north star is to make a complex mobile scenario reproducible from one portable definition, without requiring a real backend, cloud account, Docker, or a framework-specific SDK.

## v0.1 scope

The first release is a modular monolith distributed as one Go binary. It contains four cooperating capabilities:

1. an HTTP API and authentication sandbox;
2. a platform-neutral scenario engine;
3. a device engine accessed only through adapter interfaces;
4. a CLI and a small embedded dashboard as inbound adapters.

Framework SDKs remain optional. A Flutter, React Native, native Android, native iOS, or Capacitor application can use the HTTP sandbox without modifying application code beyond choosing the local endpoint.

## Dependency rule

Dependencies point inward:

```text
CLI / HTTP / YAML / SQLite / adb / simctl
                  |
          infrastructure adapters
                  |
          application use cases
                  |
          domain models and ports
```

The domain never imports CLI, HTTP, YAML, SQLite, `adb`, `simctl`, dashboard, or framework packages. Application services coordinate domain ports. Infrastructure implements those ports. Composition happens only in `cmd/mobilelab` and `internal/app`.

## Repository layout

```text
cmd/mobilelab/              executable composition root
internal/domain/            scenario, device, request and capability models
internal/app/               use cases and lifecycle orchestration
internal/config/            YAML DTOs, loading, defaults and validation
internal/sandbox/           HTTP mock/auth server and runtime fault controls
internal/scenario/          YAML parser and scenario runner adapters
internal/device/            fake, Android and iOS DeviceAdapter implementations
internal/detect/            project/tool detectors
internal/reporting/         terminal and JSON reporters
internal/storage/           repository implementations
internal/dashboard/         embedded local dashboard and typed events
pkg/mobilelab/              deliberately small public extension API
schemas/                    public configuration/scenario schemas
examples/                   runnable samples, introduced incrementally
docs/                       user and platform documentation
```

Packages are grouped by product capability rather than by technical utility. Small cross-cutting helpers stay private to their consumer until there is a demonstrated shared abstraction.

## Core ports

- `DeviceAdapter`: detection, capability reporting, app lifecycle, deep links, location, and supported network conditions.
- `RequestRepository`: append/query sanitized request records.
- `ScenarioRunRepository`: persist reproducible run results.
- `EventPublisher`: publish typed state and request events without knowing WebSocket details.
- `Clock`: deterministic latency/timestamp behavior in tests.
- `ProcessRunner`: isolate calls to official local tools such as `adb` and `xcrun simctl`.

Unsupported device operations return a typed capability error. They are never reported as successful.

## Runtime design

`mobilelab start` loads and strictly validates `mobilelab.yaml`, resolves fixtures beneath the configured project directory, starts one loopback-only HTTP server, and records its PID and control metadata in a local state file. Sandbox routes and internal control/dashboard routes share a listener initially to keep the process and port model simple. `stop` uses an authenticated local control endpoint whose random token is stored with owner-only permissions; it does not kill arbitrary processes. Graceful shutdown handles both the control request and OS signals.

Mutable fault state (global latency, forced HTTP error, auth expiry) is held behind a concurrency-safe application port. CLI mutation commands reach the running process through loopback control endpoints. Configuration files remain declarative and are not rewritten for temporary runtime faults.

The server binds to `127.0.0.1` by default. Binding to a non-loopback address requires explicit configuration and produces a warning. Request capture removes authorization, cookies, tokens, passwords, and configurable sensitive fields before logging or persistence.

## Configuration and scenario boundaries

YAML structures are transport DTOs. Parsers convert them into validated domain types before execution. Unknown YAML fields fail validation. Paths are normalized and checked to remain within the MobileLab workspace before fixtures are read.

Scenario execution is platform-neutral. A scenario selects an adapter at the composition boundary; steps operate on `DeviceAdapter` and sandbox-control ports. Assertions consume observed request/response records. The initial `FakeDeviceAdapter` makes all engine tests independent of Android and iOS installations.

## Persistence

SQLite implements the `RequestRepository` and `ScenarioRunRepository` ports in `mobilelab/mobilelab.db`. Schema migrations are transactional and versioned, WAL mode and a bounded busy timeout support concurrent dashboard/control reads, and results are returned in chronological order from a bounded recent window. The in-memory repository remains available for deterministic tests. Secrets are redacted before crossing the repository boundary. PID/control metadata is intentionally a small owner-only state file, not domain persistence.

## Device adapters

- Android invokes only verified `adb` commands. Discovery, app launch/stop, explicit `pm clear`, deep links, and emulator-only location are capability-gated. Emulator startup remains unavailable until an AVD-name port is introduced.
- iOS invokes `xcrun simctl` only on macOS and reports unavailable capabilities elsewhere. Detected shutdown simulators expose boot; booted simulators expose launch/stop, explicit uninstall-as-clear, deep links, and location. Portable runtime checks preserve cross-platform builds.
- SDK/framework adapters are optional outbound integrations and never branch through the core by framework name.

## Realtime and dashboard

Application events use versioned envelopes (`type`, `version`, `timestamp`, `payload`). Producers know only `EventPublisher`; a non-blocking in-process bus fans sanitized events out to the WebSocket adapter. New dashboard connections first receive a persisted snapshot, then live request, fault-state, and scenario events. The dashboard is embedded into the Go binary so Node.js is not required at runtime, and both its page and event endpoint are loopback-only even if the mock API is explicitly exposed to a network.

## Dependencies

The dependency footprint is intentionally small:

- Go standard library for HTTP, CLI parsing, process management, embedding, JSON, and concurrency;
- `gopkg.in/yaml.v3` for strict YAML decoding;
- `github.com/getkin/kin-openapi` v0.133 for OpenAPI 3 loading, reference resolution, and validation (a maintained release compatible with the Go 1.23 baseline selected for this first importer slice);
- `modernc.org/sqlite` v1.38.2 for SQLite without a mandatory C toolchain (Go 1.23 baseline);
- `github.com/coder/websocket` v1.8.15 for the same-origin dashboard event transport (Go 1.23 baseline).

Every new dependency must remove meaningful implementation risk and be actively maintained. The repository pins direct versions in `go.mod` and CI checks all supported operating systems.

## Architectural decisions

1. **Go modular monolith:** one deployable keeps startup, installation, shutdown, and local debugging simple while package boundaries preserve future extension points.
2. **Ports over platform conditionals:** Android/iOS differences live in adapters and capability values, never scattered framework checks.
3. **Standard-library CLI initially:** command handlers remain thin and testable without committing the public UX to a CLI framework.
4. **One loopback server initially:** API, control, and embedded dashboard share lifecycle and eliminate partial startup failures. Typed interfaces allow later port separation.
5. **Strict input at boundaries:** unknown fields, invalid methods/statuses/ports, unsafe fixture paths, and malformed scenarios fail before the environment starts.
6. **Temporary faults are runtime state:** latency/error/auth commands are reversible and do not silently mutate project configuration.
7. **Honest capabilities:** partial or unavailable device features are explicit domain results, not no-op success.

## Incremental delivery plan

1. Bootstrap the module; implement strict configuration, project detection, and safe `init` generation.
2. Add the API sandbox with routes, fixtures, variables, latency, forced errors, basic auth, request capture, and redaction.
3. Add lifecycle/control commands (`start`, `status`, `stop`, `api`) with graceful shutdown.
4. Add scenario domain/parser/runner, fake device adapter, assertions, and JSON reports.
5. Add real Android and conditional iOS slices, doctor/capabilities, dashboard events, and OpenAPI import.
6. Add SQLite repositories and live dashboard events, then continue with examples, richer assertions, and release packaging.

Each slice must finish with formatting, focused tests, the full test suite, and a cross-platform-oriented build before the next slice starts.
