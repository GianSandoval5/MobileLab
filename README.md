<div align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/mobilelab-wordmark-dark.svg">
    <source media="(prefers-color-scheme: light)" srcset="docs/assets/mobilelab-wordmark-light.svg">
    <img alt="MobileLab" src="docs/assets/mobilelab-wordmark-light.svg" width="680">
  </picture>

  <p><strong>Build the mobile failure before it reaches production.</strong></p>
  <p>A local-first API sandbox, device adapter, and portable scenario runner for mobile apps.<br>No real backend. No cloud account. No Docker. No framework lock-in.</p>

  <p>
    <a href="https://github.com/GianSandoval5/MobileLab/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/v/release/GianSandoval5/MobileLab?style=flat-square&color=0969da"></a>
    <a href="https://github.com/GianSandoval5/MobileLab/actions/workflows/ci.yml"><img alt="Core CI" src="https://img.shields.io/github/actions/workflow/status/GianSandoval5/MobileLab/ci.yml?branch=master&style=flat-square&label=core%20CI"></a>
    <a href="https://github.com/GianSandoval5/MobileLab/actions/workflows/sdk.yml"><img alt="SDK CI" src="https://img.shields.io/github/actions/workflow/status/GianSandoval5/MobileLab/sdk.yml?branch=master&style=flat-square&label=SDKs"></a>
    <a href="LICENSE"><img alt="MIT license" src="https://img.shields.io/github/license/GianSandoval5/MobileLab?style=flat-square"></a>
    <img alt="Go 1.23 or newer" src="https://img.shields.io/badge/Go-1.23%2B-00ADD8?style=flat-square&logo=go&logoColor=white">
  </p>

  <p>
    <a href="#quick-start">Quick start</a> ·
    <a href="#why-mobilelab">Why MobileLab?</a> ·
    <a href="#scenarios">Scenarios</a> ·
    <a href="#optional-framework-sdks">SDKs</a> ·
    <a href="#runnable-example-applications">Examples</a> ·
    <a href="#architecture">Architecture</a> ·
    <a href="docs/configuration.md">Docs</a>
  </p>
</div>

---

MobileLab lets Flutter, React Native, Android, iOS, and Capacitor applications exercise repeatable backend failures and device scenarios from one framework-neutral CLI. Version **1.0.0** provides stable configuration and scenario contracts, safe migrations, project-local plugins, CI reporting, record/replay, and optional framework integrations.

> [!NOTE]
> MobileLab is built for local development and CI. Start with the [current limitations](#current-limitations), review the [compatibility policy](docs/compatibility.md), or see exactly how the project maps to [`PRODUCT_SPEC.md`](docs/spec-coverage.md).

## Why MobileLab?

Mobile applications fail at boundaries: expired sessions, slow responses, malformed backends, deep links, device location, and unavailable networks. MobileLab brings those boundaries into one local CLI and one framework-neutral scenario format:

> Write the mobile scenario once. Run it everywhere.

The Go core is a single process, binds to `127.0.0.1` by default, and does not require Node.js, Flutter, or Docker. Optional framework SDKs add advanced app events; basic HTTP mocks do not need them.

## Installation

Download the binary for your operating system from [GitHub Releases](https://github.com/GianSandoval5/MobileLab/releases), verify it against `SHA256SUMS`, rename it to `mobilelab` (`mobilelab.exe` on Windows), and place it on your `PATH`.

To build from source, use Go 1.23 or newer:

```sh
git clone https://github.com/GianSandoval5/MobileLab.git
cd MobileLab
make build
./bin/mobilelab --version
```

## Quick start

The paths below are examples—replace them with the real locations on your computer. First, enter the directory where you cloned or saved MobileLab and build the CLI:

```sh
# This is the directory where you saved the MobileLab repository.
cd /path/where/you/saved/MobileLab

make build
```

Next, enter the root directory of the mobile application you want to test. This is your Flutter, React Native, Android, iOS, Ionic, or Capacitor project—not the MobileLab repository:

```sh
# This is the directory of your own mobile application.
cd /path/to/your/mobile/project

# Run the binary from the directory where you saved MobileLab.
/path/where/you/saved/MobileLab/bin/mobilelab init
/path/where/you/saved/MobileLab/bin/mobilelab doctor
/path/where/you/saved/MobileLab/bin/mobilelab start
```

For example, if MobileLab is at `/Users/you/projects/MobileLab` and your application is at `/Users/you/projects/MyApp`, those commands become:

```sh
cd /Users/you/projects/MyApp
/Users/you/projects/MobileLab/bin/mobilelab init
/Users/you/projects/MobileLab/bin/mobilelab doctor
/Users/you/projects/MobileLab/bin/mobilelab start
```

Run `init` once per mobile project. It inspects the current directory for real project evidence (`pubspec.yaml`, React Native dependencies, Gradle/manifest files, Xcode projects, Ionic and Capacitor configuration), then creates:

```text
mobilelab.yaml
mobilelab/fixtures/
mobilelab/mocks/
mobilelab/scenarios/
```

After initialization, `doctor` validates the generated Core configuration and checks which local platform/framework tools are available, such as ADB, the Android Emulator, Xcode, Flutter, Node.js, Java, and Swift. It reports readiness and actionable warnings without changing your application.

Run `start` from that same mobile-project directory. It reads the generated `mobilelab.yaml`, starts the local sandbox and dashboard (normally at `http://127.0.0.1:4566`), and stays in the foreground. Press Ctrl+C to stop it.

If you install the binary on your `PATH`, you no longer need its full filesystem location:

```sh
cd /path/to/your/mobile/project
mobilelab init
mobilelab doctor
mobilelab start
```

With the generated environment running:

```sh
curl http://127.0.0.1:4566/api/profile
mobilelab api latency 3000
mobilelab api error 500
mobilelab api reset
mobilelab status
mobilelab scenario history
mobilelab stop
```

Print the URL appropriate for the host or a detected emulator/simulator:

```sh
mobilelab endpoint
mobilelab endpoint --platform android --device emulator-5554
mobilelab endpoint --platform ios --device <simulator-udid> --json
```

For the standard Android Emulator, a loopback Core becomes `http://10.0.2.2:<port>`. iOS Simulator shares the Mac network namespace. Physical devices require either an explicitly reachable trusted-network host or a deliberately configured tunnel such as `adb reverse`; MobileLab refuses to present a wildcard bind address as a reachable URL.

`start` remains in the foreground and shuts down cleanly on Ctrl+C. From another terminal, `stop` uses a token-authenticated loopback control endpoint—not arbitrary process termination. Use `start --headless` to disable the embedded dashboard.

## Runnable example applications

Two complete source-only shop applications exercise the same MobileLab fixtures and API contract:

- [Flutter example](examples/apps/flutter/README.md), using Dart, Riverpod, and a platform-aware API URL.
- [React Native example](examples/apps/react-native/README.md), using TypeScript, Zustand, React Navigation, and AsyncStorage.

Both include `mobilelab.yaml`, fixtures, scenarios, native Android/iOS projects, and verification commands. Generated dependencies and builds are excluded, keeping the combined examples near 2 MB instead of several gigabytes. See the [example applications guide](examples/apps/README.md) for URLs and startup instructions.

## API Sandbox

Endpoints are declared in strict YAML. Unknown fields and invalid methods, ports, paths, statuses, or duplicate route templates fail before startup.

The `endpoints` list is MobileLab's equivalent of an API collection. Every HTTP URL has a path: opening only `http://127.0.0.1:4566` requests `GET /`, while the dashboard lives at `/dashboard`. Declare a `GET /` mock if the base URL should return content; otherwise a `no mock configured` response is expected and does not mean that Core stopped.

```yaml
schema_version: 1
variables:
  userId: "123"

push:
  payment-success:
    title: Payment completed
    body: Your payment was processed
    data:
      transactionId: ABC123

endpoints:
  - path: /api/users/{id}
    method: GET
    delay: 200
    response:
      status: 200
      headers:
        X-Source: MobileLab
      fixture: profile.json
```

Fixtures are loaded only from `mobilelab/fixtures`, cannot escape via `..` or symlinks, must remain valid JSON after `{{variable}}` substitution, and can be replaced by inline `body` values. Global runtime faults are temporary and do not rewrite configuration.

Routes such as `/api/users/{id}` match concrete paths such as `/api/users/456`. MobileLab records path, query parameters, headers, and JSON request bodies, but configuration v1 returns the response selected by method/path; it does not yet branch on query/body/header values or inject a captured `{id}` into the response. Configuration variables are static values. OpenAPI 3 documents can generate mocks with `mobilelab api import`; direct Postman collection import is not part of 1.0.

### Local auth sandbox

When `auth.enabled` is true, MobileLab provides development-only login and refresh endpoints:

```sh
curl -X POST http://127.0.0.1:4566/__mobilelab/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"mobilelab","password":"mobilelab"}'

mobilelab auth expire
mobilelab auth reset
```

Mark an endpoint with `protected: true` to require the emitted HS256 access token. Signing keys are random and process-local. These fake credentials and tokens must never be used as production identity.

Every sandbox request is captured in SQLite at `mobilelab/mobilelab.db`. Authorization, cookies, API keys, passwords, tokens, secrets, and sensitive nested JSON fields are redacted before reaching persistence. Transactional migrations are applied automatically; generated state/database files are ignored by the `mobilelab/.gitignore` created during initialization.

## OpenAPI 3 import

Import into an existing environment or during initialization:

```sh
mobilelab api import openapi.yaml
mobilelab init --openapi openapi.yaml
```

The importer uses `kin-openapi` for parsing, reference resolution, and validation. It adds or updates route mocks, creates deterministic schema examples, and creates missing `openapi-*.yaml` scenarios. This first slice accepts self-contained OpenAPI 3 documents; remote/external references are intentionally disabled.

## Scenarios

MobileLab 1.x stabilizes configuration and scenario schema v1. Newly initialized/generated YAML contains `schema_version: 1`; legacy pre-1.0 files remain readable and can be upgraded safely:

```sh
mobilelab migrate --check
mobilelab migrate
mobilelab schema config > mobilelab-config-v1.schema.json
mobilelab schema scenario > mobilelab-scenario-v1.schema.json
```

Migration preflights every document before writing and atomically replaces only legacy files. See the [configuration reference](docs/configuration.md), [scenario reference](docs/scenarios.md), and [compatibility policy](docs/compatibility.md).

Scenario YAML is independent of Flutter, React Native, and other application frameworks:

```yaml
schema_version: 1
name: Payment with expired session
backend:
  latency: 2000
  error: 500
auth:
  token: expired
steps:
  - launch_app
  - open_deeplink: myapp://payments
expect:
  - request:
      method: POST
      path: /api/payments
  - response:
      status: 500
  - app_event:
      framework: flutter
      kind: assertion
      name: checkout.ready
      passed: true
```

Run it while the environment is active:

```sh
mobilelab run mobilelab/scenarios/payment.yaml \
  --platform android \
  --app-id com.example.app

mobilelab run mobilelab/scenarios/payment.yaml \
  --platform fake \
  --app-id com.example.app \
  --report json \
  --output report.json

mobilelab scenario list
mobilelab scenario run payment --platform android --app-id com.example.app
mobilelab scenario history --json
```

`run` also accepts a directory, recursively preflights every `.yaml`/`.yml` file, and executes the lexically sorted suite even when an earlier scenario fails:

```sh
mobilelab run mobilelab/scenarios \
  --platform fake \
  --timeout 10s \
  --report junit \
  --output artifacts/mobilelab-junit.xml

mobilelab run mobilelab/scenarios \
  --platform fake \
  --report html \
  --output artifacts/mobilelab-report.html
```

Supported formats are `terminal`, `json`, `junit`, and `html`. JUnit emits one test suite per scenario and one test case per executed check. HTML is a standalone, responsive file with escaped scenario content and no remote assets. Output parent directories are created automatically. A suite returns a non-zero exit code if any scenario fails, while still reporting every scenario that could run.

The fake platform is intended for engine/CI tests. Assertions poll only requests observed after the run begins. The runner resets temporary faults before and after every run.

`app_event` assertions likewise consider only events received after the scenario starts. `framework` is optional; `kind` is `lifecycle`, `marker`, or `assertion`. When `passed` is omitted for an assertion, MobileLab requires the SDK assertion to report `true`.

## Record and replay

With the environment running, start an interactive recording in a second terminal:

```sh
mobilelab record login
# Exercise the app, change API/auth state, and open deep links.
# Press Ctrl+C when finished.
mobilelab replay login --platform android --device emulator-5554 --app-id com.example.app
```

For scripts, use `--duration 30s`. Existing scenario files are protected unless `--force` is explicit. Core records one session at a time and captures sanitized HTTP request/response metadata and JSON bodies (maximum 1 MiB), successful MobileLab deep links, and runtime latency/error/auth/reset changes. The generated file is written atomically under `mobilelab/scenarios/`, parsed by the same strict YAML parser, and includes HTTP synchronization points so later fault changes cannot overtake earlier observed requests during replay. An empty capture is rejected without creating a scenario.

Recorded secrets remain `[REDACTED]`; a recording never has access to the original authorization, cookie, token, password, or configured sensitive values. Replay executes actions and validates observed request/status pairs—it does not resend captured credentials or arbitrary production traffic.

## Optional framework SDKs

The optional SDKs share protocol v1 and send lifecycle, marker, and assertion events to `POST /__mobilelab/sdk/events`:

- Flutter: [`mobilelab_flutter`](sdk/flutter/mobilelab_flutter/README.md)
- React Native: [`@mobilelab/react-native`](sdk/react-native/README.md)
- Android/Kotlin: [`mobilelab-android`](sdk/android/mobilelab-android/README.md)
- iOS/Swift: [`MobileLabKit`](sdk/ios/MobileLabKit/README.md)
- Capacitor: [`@mobilelab/capacitor`](sdk/capacitor/README.md)

All expose `lifecycle`, `marker`, `assertThat`, and lifecycle reporting suited to their platform. Events are strictly validated, limited to 64 KiB, sanitized, persisted in SQLite, streamed to the dashboard, and available to scenario assertions. SDK transport failures are observable and automatic lifecycle callbacks do not crash the application.

Android Emulator applications normally use `http://10.0.2.2:4566`; iOS Simulator normally uses `http://127.0.0.1:4566`. A physical device requires an explicitly reachable trusted-network bind. Development builds may also need Android cleartext HTTP or iOS local-network/ATS configuration. Do not enable those exceptions in production builds.

The [shared SDK event example](examples/sdk-events/README.md) supplies one sandbox and equivalent Android, iOS, and Capacitor marker scenarios.

## Android and iOS

```sh
mobilelab detect
mobilelab capabilities
mobilelab deeplink open 'myapp://payments/123'
mobilelab location set -5.1945 -80.6328
mobilelab device list
mobilelab device info --platform ios --device <simulator-udid> --json
mobilelab device launch --platform android --device emulator-5554 --app-id com.example.app
mobilelab device stop --platform android --device emulator-5554 --app-id com.example.app
mobilelab device clear --platform android --device emulator-5554 --app-id com.example.app
mobilelab device boot --platform android --device avd:Pixel_9_API_35
mobilelab device boot --platform ios --device <shutdown-simulator-udid>
mobilelab network slow --platform android --device emulator-5554
mobilelab network online --platform android --device emulator-5554
mobilelab push send payment-success --platform ios --device <booted-simulator-udid> --app-id com.example.app
```

Android uses the official `adb` and Android Emulator executables. MobileLab resolves them from `PATH`, `ANDROID_HOME`, or `ANDROID_SDK_ROOT`. Configured, inactive AVDs are listed with IDs prefixed by `avd:` and can be started explicitly with `device boot`; a running AVD is represented by its attached `emulator-*` serial instead. Deep links use `adb shell am start`; location uses `adb emu geo fix` and is therefore limited to running emulators. Explicit `device clear` uses `adb shell pm clear` and deletes the selected application's local data.

On a running Android Emulator, `network slow` applies the official `network delay gprs` and `network speed gprs` console profiles; `network online` restores `delay none` and `speed full`. These capabilities are reported `partial` because the emulator documentation limits shaping to Ethernet and cellular traffic, and recent emulator Wi-Fi may use a separate network simulator. `network offline` remains unavailable: MobileLab does not toggle radios or airplane mode and pretend that all transports were disconnected. iOS Simulator network conditioning remains unavailable because `simctl` offers no portable equivalent. See the [Android Emulator console reference](https://developer.android.com/studio/run/emulator-console).

`mobilelab device info` reports runtime metadata as well as the capability matrix. For connected Android devices this includes a safe allowlist of `getprop` values (manufacturer, brand, model, OS/API version, and ABI); emulator-console features become fully available only after a successful runtime probe. For iOS Simulator it includes the CoreSimulator runtime, derived OS version, device type, and last boot time when `simctl` supplies them.

iOS uses `xcrun simctl` and is available only on macOS with Xcode. It supports discovery, simulator boot, app launch/terminate, deep links, and simulator location. On iOS, explicit `device clear` uses `simctl uninstall`, so the selected app is removed rather than merely having its data reset. `simctl` does not provide a portable network-conditioning command, so that capability is reported unavailable.

Push fixtures live under the top-level `push` configuration key. A booted iOS Simulator advertises push only when `simctl help push` succeeds; `push send` then builds a private temporary APNs JSON payload, enforces the 4096-byte simulator limit, and invokes `simctl push` for the explicit bundle ID. Android reports push unavailable because there is no generic official local delivery path without FCM credentials or an app-specific receiver.

For an Android Emulator, `localhost` is the emulator itself. Use `http://10.0.2.2:4566` to reach MobileLab on the development host. iOS Simulator can normally use the host loopback address.

## Dashboard

The embedded dashboard at `http://127.0.0.1:4566/dashboard` receives a persisted snapshot followed by typed WebSocket events for sanitized requests, active latency/errors/auth state, and scenario results. Slow dashboard clients never block the API Sandbox; their oldest buffered event is discarded in favor of current state. The dashboard and WebSocket endpoints remain loopback-only even when a mock API is explicitly bound to a network address. Node.js is not required at runtime.

## Project-local plugins

MobileLab v0.7 introduces the versioned `mobilelab.plugin/v1` protocol. A plugin is a short-lived executable plus a strict manifest at `mobilelab/plugins/<name>/plugin.yaml`; it can be written in any language that reads one JSON request from standard input and writes one JSON response to standard output.

```sh
mobilelab plugin list
mobilelab plugin inspect echo
mobilelab plugin inspect echo --json
mobilelab plugin run echo echo --input input.json --timeout 30s
mobilelab plugin run echo echo --input input.json --output artifacts/result.json
```

Discovery and inspection never execute plugin code. Execution occurs only through an explicit `plugin run`, uses no shell, defaults to a 30-second timeout, caps protocol messages at 1 MiB, and passes a minimal allowlisted environment. Output files are created with owner-only permissions. The inspect command reports the exact executable SHA-256 fingerprint.

Plugins are nevertheless trusted local programs: v0.7 does not provide an operating-system sandbox and a plugin runs with the user's filesystem and network permissions. Review third-party code and its fingerprint before execution. MobileLab deliberately has no network-backed plugin installer yet, and cloud integrations such as Firebase or Supabase will be selected only from demonstrated demand instead of becoming mandatory dependencies.

Go plugin authors can use the public [`pkg/plugin`](pkg/plugin) package. The runnable [echo example](examples/plugins/README.md) demonstrates the manifest, authoring API, discovery, inspection, and explicit invocation without a running sandbox.

## Architecture

MobileLab is a Go modular monolith with inward-pointing dependencies and explicit ports for devices, request storage, scenarios, environment control, process execution, and future events. Platform details stay in adapters. See [ARCHITECTURE.md](ARCHITECTURE.md) and the honest [`PRODUCT_SPEC.md` coverage audit](docs/spec-coverage.md).

## Testing and CI

```sh
make test
make lint
make build
```

Unit tests cover strict configuration, project detection, secret redaction, fixture confinement, auth, faults, routing, OpenAPI generation, device command construction, SDK protocol ingestion, scenario execution, suite aggregation, safe report rendering, and plugin protocol boundaries. Integration tests open temporary loopback servers for lifecycle and recorder behavior. Core CI runs on Linux, macOS, and Windows; an executable plugin example runs in CI, while separate SDK CI pins Flutter 3.44.4, Node 22.23.2, and Java 17 and validates the Swift Package on macOS.

The executable [CI example](examples/ci/README.md) starts MobileLab with `--headless`, runs a directory through the fake adapter, and publishes JUnit XML, standalone HTML, the Core log, and a response fixture. Copy-ready definitions cover GitHub Actions, GitLab CI, Azure Pipelines, and Jenkins with provider-native test/artifact publication; retention is explicit where the provider exposes it in pipeline code and documented as an Azure project setting otherwise. Real emulators/simulators remain opt-in: provision a booted device on the runner, then replace `--platform fake` and supply its explicit device/application identifiers.

## Current limitations

- Scenario device selection is first-ready unless `--device` is supplied; direct device, deep-link, location, network, and push commands accept explicit selectors.
- Android Emulator Ethernet/cellular latency and speed shaping is partial; reliable offline mode and iOS network conditioning are unavailable. Local push is limited to booted iOS Simulators with supported Xcode tooling.
- OpenAPI import does not yet support external references, callbacks, GraphQL, gRPC, or sophisticated example generation.
- All framework integrations are optional observability SDKs; they are not required for the API Sandbox.
- Plugins are manually installed, trusted project-local executables in 1.0. MobileLab limits their invocation but does not provide an OS sandbox, registry, signature verification, dependency resolution, or automatic updates.
- Automatic multi-platform `--all`, GraphQL/gRPC/SSE transports, enterprise identity providers, a remote plugin registry, and a public structured logging/`--verbose` contract are not part of 1.0.
- SQLite retention/pruning policy is not implemented yet; schema migrations currently cover versions 1 through 3. Active recordings are process-local and are finalized by the `record` command.

See [ROADMAP.md](ROADMAP.md) for planned milestones. Contributions are welcome; read the [contributing guide](CONTRIBUTING.md), [code of conduct](CODE_OF_CONDUCT.md), and [security policy](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
