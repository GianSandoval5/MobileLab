# MobileLab Architecture

MobileLab is a local-first mobile development environment. Its architectural north star is to make a complex mobile scenario reproducible from one portable definition, without requiring a real backend, cloud account, Docker, or a framework-specific SDK.

## v0.1 scope

The first release is a modular monolith distributed as one Go binary. It contains four cooperating capabilities:

1. an HTTP API and authentication sandbox;
2. a platform-neutral scenario engine;
3. a device engine accessed only through adapter interfaces;
4. a CLI and a small embedded dashboard as inbound adapters.

Framework SDKs remain optional. A Flutter, React Native, native Android, native iOS, or Capacitor application can use the HTTP sandbox without modifying application code beyond choosing the local endpoint.

From 0.3 onward, optional SDKs share a versioned HTTP event contract owned by the Core. Framework packages translate their native lifecycle and assertion APIs into that contract; they do not become device adapters and the Core never branches on application framework during scenario execution. Accepted events are validated, sanitized before persistence, and published through the existing typed event bus.

For 0.4, Android, iOS, and Capacitor use the same protocol v1 endpoint and domain event model as Flutter and React Native. Supported framework identifiers are centralized in the domain instead of being repeated across transports. Native packages use platform HTTP primitives, expose injectable transports for deterministic tests, and add lifecycle helpers at their own boundary. The Kotlin client stays Android-compatible without requiring Android classes in its core artifact; `MobileLabKit` conditionally layers UIKit observation over a Foundation client; the Capacitor package accepts the App plugin through a narrow structural interface so installing the SDK does not add a second Capacitor runtime.

For 0.5, recording is an application service behind a typed domain port. Core owns at most one active recording and accepts already-sanitized HTTP exchanges, runtime environment mutations, and successful deep links from their existing adapters. Stopping a recording returns an immutable ordered event stream; a separate generator converts it to the public scenario model and writes YAML atomically. Generated HTTP synchronization steps preserve event order across later environment mutations, while final request/response assertions retain the readable result report. Replay remains an alias over the existing scenario runner, so recorded scenarios do not introduce a second execution engine.

For 0.6, a scenario suite is a domain aggregate over one or more ordinary scenario results. Directory discovery and strict preflight parsing stay in the CLI adapter; execution remains sequential through the same scenario runner and preserves per-scenario environment isolation. Reporting adapters consume the completed suite without reaching into the runtime: terminal and JSON remain available, JUnit XML maps checks to CI test cases, and a standalone HTML adapter renders escaped user-controlled content with no external assets or JavaScript. Provider examples own artifact retention because retention is a CI policy, not Core persistence.

For 0.7, extensions cross an explicit out-of-process boundary instead of being linked into Core. A project-local catalog reads strict `mobilelab.plugin/v1` manifests without executing them, resolves each executable within its declared plugin directory, and fingerprints its current contents. Only `mobilelab plugin run` starts a plugin: one bounded JSON request is sent on standard input and one correlated, bounded JSON response is accepted on standard output. The host supplies a minimal environment, a deadline, and no shell. The public Go package models this wire protocol for authors, while the protocol remains language-neutral. Plugins are trusted local programs rather than a security sandbox, and neither startup nor scenario execution loads them implicitly.

For 1.0, the YAML transport boundary is explicitly versioned. Configuration and scenario schema v1 are published as embedded JSON Schemas; parsers accept legacy unversioned documents during 1.x, reject future versions, bound input size, and require exactly one YAML document. A migration application service preflights configuration and every recursive scenario before performing comment/permission-preserving atomic replacements. Endpoint resolution is a separate application policy over configuration plus device facts, so the CLI can report host, standard Android Emulator, iOS Simulator, and physical-device reachability without leaking platform conditionals into scenarios.

For 1.1, optional business data is a separate adapter boundary declared by `mobilelab/data.yaml` schema v1 and persisted in `mobilelab/data.db`. A generic JSON-document store exposes collection and singleton CRUD without allowing raw SQL. It composes beneath the existing sandbox request pipeline, so faults, authentication, redaction, capture, and realtime request events have one implementation. Static fixture endpoints retain their 1.x behavior. The business database never shares tables or migrations with Core's operational history, and automatic startup seeding is non-destructive.

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
internal/datastore/         optional business-data YAML, SQLite store, and REST adapter
internal/sandbox/           HTTP mock/auth server and runtime fault controls
internal/scenario/          YAML parser and scenario runner adapters
internal/device/            fake, Android and iOS DeviceAdapter implementations
internal/detect/            project/tool detectors
internal/reporting/         terminal, JSON, JUnit XML and standalone HTML reporters
internal/plugins/           strict manifest catalog and bounded process host
internal/migration/         preflighted public YAML schema upgrades
internal/endpoint/          honest host/device URL resolution policy
internal/storage/           repository implementations
internal/dashboard/         embedded local dashboard and typed events
sdk/                        optional framework clients; never imported by Core
examples/sdk-events/        shared sandbox and framework-neutral event scenarios
examples/ci/                executable headless suite and provider pipeline definitions
examples/plugins/           executable project-local plugin example
pkg/plugin/                 public, versioned plugin authoring protocol
schemas/                    public configuration/data/scenario schemas
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

Suite inputs are discovered recursively in lexical order and every YAML document is parsed before the first scenario executes, avoiding partial runs caused by a late malformed file. Failures do not stop later scenarios, so CI receives one complete report and a non-zero process exit. Report output directories are created explicitly at the CLI boundary.

## Plugin boundary

Plugin discovery is local to `<project>/mobilelab/plugins`; there is no global search path, network registry, automatic installation, or implicit execution. Manifest names and actions use a constrained identifier grammar, versions are SemVer, unknown YAML fields fail, executable paths are base names, and resolved symlinks must remain inside the plugin directory. Inspection hashes the resolved regular executable.

Invocation uses a short-lived child process with the plugin directory as its working directory. The host passes only path/temp variables needed for process startup plus `MOBILELAB_PLUGIN_PROTOCOL` and `MOBILELAB_PLUGIN_NAME`; arbitrary parent variables are excluded. Requests and responses are strict JSON documents capped at 1 MiB and correlated with cryptographically random request IDs. The CLI default deadline is 30 seconds and its public upper bound is five minutes. A plugin cannot add in-process routes, mutate MobileLab memory, or weaken the single-binary installation path.

These controls reduce accidental coupling and secret inheritance, but do not create an OS sandbox. Once explicitly invoked, a plugin has the same user-level filesystem and network authority as any local executable. Distribution trust, signatures, installation, dependency resolution, and automatic updates remain outside the v0.7 contract.

## Persistence

SQLite implements the `RequestRepository` and `ScenarioRunRepository` ports in `mobilelab/mobilelab.db`. Schema migrations are transactional and versioned, WAL mode and a bounded busy timeout support concurrent dashboard/control reads, and results are returned in chronological order from a bounded recent window. The in-memory repository remains available for deterministic tests. Secrets are redacted before crossing the repository boundary. PID/control metadata is intentionally a small owner-only state file, not domain persistence.

The optional business store uses a second SQLite database, `mobilelab/data.db`, with generic resource/document rows and a public declarative YAML boundary. Its JSON values are always passed as bound SQL parameters. Seed files are size-bounded and confined beneath `mobilelab/`; request documents are size-bounded JSON objects. Keeping the databases separate prevents business resets from deleting request history and allows older projects without `data.yaml` to run unchanged.

Schema v2 adds `AppEventRepository`. The public SDK ingestion endpoint accepts only protocol v1 DTOs, converts them to domain events, replaces sensitive nested attributes before storage, and then publishes `app.event`. The authenticated control adapter exposes recent app events to the scenario runner; SDK clients never receive the control token.

Schema v3 adds sanitized response headers and response bodies to request history. Recording sessions themselves are deliberately process-local in v0.5: the active Core owns ordering and finalization, while the generated YAML is the durable portable artifact. An interrupted Core cannot claim that an incomplete recording was finalized.

## Device adapters

- Android invokes only verified `adb` and Android Emulator commands. Discovery combines attached targets with configured AVDs; inactive AVDs use stable `avd:<name>` IDs and a non-blocking process port starts the explicitly selected AVD. Running AVDs are de-duplicated when the emulator console reports their names. App launch/stop, explicit `pm clear`, deep links, and emulator-only location remain capability-gated.
- iOS invokes `xcrun simctl` only on macOS and reports unavailable capabilities elsewhere. Detected shutdown simulators expose boot; booted simulators expose launch/stop, explicit uninstall-as-clear, deep links, and location. Portable runtime checks preserve cross-platform builds.
- Device discovery doubles as a runtime capability probe. Android enriches attached targets from a strict property allowlist and promotes emulator-console features only after the console responds. iOS derives metadata and lifecycle capabilities from the current `simctl` inventory. Human-readable and JSON inspection use the same domain snapshot.
- Android network shaping uses only documented emulator-console delay/speed profiles and is labeled partial because it covers Ethernet/cellular rather than every modern transport. Restoring `online` removes those two constraints. Reliable offline behavior and iOS Simulator conditioning stay explicitly unavailable.
- Local push is a device port backed only by `simctl push` when that command is runtime-probed. Configuration fixtures become size-limited APNs payloads in owner-private temporary files that are deleted immediately after invocation. Android stays unavailable without a generic official delivery mechanism.
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
8. **One SDK contract, thin platform clients:** framework packages own lifecycle wiring and transport ergonomics while validation, persistence, redaction, and scenario semantics remain in Core.
9. **Injectable SDK transports:** each package separates event construction from network I/O so its contract can be tested without a running simulator, emulator, or Core process.
10. **Coordinated SDK release versions:** framework packages follow the MobileLab release tag even when a release only broadens adapters; GitHub assets remain traceable to one protocol-compatible source revision.
11. **Reporters are pure output adapters:** reporting consumes immutable domain results and performs no device, HTTP, SQLite, or process operations; standard-library XML/templates keep CI formats portable and user content escaped.
12. **Plugins are explicit out-of-process extensions:** strict project-local manifests and a versioned JSON protocol preserve framework/language neutrality; no plugin code runs during discovery, startup, or ordinary scenarios.
13. **No registry before a trust model:** v0.7 exposes inspection and SHA-256 fingerprints but does not download executables or imply authenticity. First-party cloud integrations wait for demonstrated demand.
14. **Version public documents, not domain objects:** `schema_version` belongs to YAML DTOs; the scenario domain remains transport-independent while v1 schemas and compatibility checks live at adapters.
15. **Preflight migrations before mutation:** every targeted YAML file must parse under the old compatibility rules before any file is replaced; replacements preserve permissions and use same-directory atomic rename.
16. **Resolve endpoints from device facts:** Android Emulator host aliases and iOS Simulator loopback are explicit policy results. Physical devices and wildcard bind addresses fail with remediation instead of returning a URL that cannot work.
17. **Separate operational and business persistence:** `mobilelab.db` is Core-owned history; optional `data.db` is user-controlled synthetic API state. They have independent schemas and lifecycles.
18. **Declarative CRUD instead of SQL passthrough:** data resources expose bounded JSON documents over predictable REST routes. Mobile clients never receive filesystem or raw SQL access.
19. **Non-destructive automatic seeding:** startup seeds only empty resources; explicit `db seed` upserts and explicit `db reset` is the sole destructive restore operation.

## Incremental delivery plan

1. Bootstrap the module; implement strict configuration, project detection, and safe `init` generation.
2. Add the API sandbox with routes, fixtures, variables, latency, forced errors, basic auth, request capture, and redaction.
3. Add lifecycle/control commands (`start`, `status`, `stop`, `api`) with graceful shutdown.
4. Add scenario domain/parser/runner, fake device adapter, assertions, and JSON reports.
5. Add real Android and conditional iOS slices, doctor/capabilities, dashboard events, and OpenAPI import.
6. Add SQLite repositories and live dashboard events, then continue with examples, richer assertions, and release packaging.
7. Add the versioned plugin protocol, project-local catalog, explicit bounded execution, author SDK, and executable conformance example.
8. Stabilize v1 YAML/plugin/SDK contracts, publish schemas and compatibility guarantees, add safe project migration, and harden cross-platform distribution.

Each slice must finish with formatting, focused tests, the full test suite, and a cross-platform-oriented build before the next slice starts.
