# MobileLab

MobileLab is a local-first API sandbox, device adapter layer, and portable scenario runner for mobile development. It lets Flutter, React Native, Android, iOS, and Capacitor applications exercise repeatable failure and device scenarios without a real backend, cloud account, Docker, or a framework SDK.

> Status: **v0.4.0** optional Flutter, React Native, Android, iOS, and Capacitor integrations. See [Current limitations](#current-limitations) and the [changelog](CHANGELOG.md).

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

```sh
./bin/mobilelab doctor

cd /path/to/your/mobile/project
/path/to/mobilelab/bin/mobilelab init
/path/to/mobilelab/bin/mobilelab start
```

`init` detects real project evidence (`pubspec.yaml`, React Native dependencies, Gradle/manifest files, Xcode projects, Ionic and Capacitor configuration), then creates:

```text
mobilelab.yaml
mobilelab/fixtures/
mobilelab/mocks/
mobilelab/scenarios/
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

`start` remains in the foreground and shuts down cleanly on Ctrl+C. From another terminal, `stop` uses a token-authenticated loopback control endpoint—not arbitrary process termination. Use `start --headless` to disable the embedded dashboard.

## API Sandbox

Endpoints are declared in strict YAML. Unknown fields and invalid methods, ports, paths, statuses, or duplicate route templates fail before startup.

```yaml
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

Scenario YAML is independent of Flutter, React Native, and other application frameworks:

```yaml
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

The fake platform is intended for engine/CI tests. Assertions poll only requests observed after the run begins. The runner resets temporary faults before and after every run.

`app_event` assertions likewise consider only events received after the scenario starts. `framework` is optional; `kind` is `lifecycle`, `marker`, or `assertion`. When `passed` is omitted for an assertion, MobileLab requires the SDK assertion to report `true`.

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

## Architecture

MobileLab is a Go modular monolith with inward-pointing dependencies and explicit ports for devices, request storage, scenarios, environment control, process execution, and future events. Platform details stay in adapters. See [ARCHITECTURE.md](ARCHITECTURE.md).

## Testing and CI

```sh
make test
make lint
make build
```

Unit tests cover strict configuration, project detection, secret redaction, fixture confinement, auth, faults, routing, OpenAPI generation, device command construction, SDK protocol ingestion, and scenario execution. An integration test opens a temporary loopback server and verifies start/status/app events/API fault/reset/stop. Core CI runs on Linux, macOS, and Windows; separate SDK CI pins Flutter 3.44.4, Node 22.23.2, and Java 17, and validates the Swift Package on macOS.

## Current limitations

- Scenario device selection is first-ready unless `--device` is supplied; direct device, deep-link, location, network, and push commands accept explicit selectors.
- Android Emulator Ethernet/cellular latency and speed shaping is partial; reliable offline mode and iOS network conditioning are unavailable. Local push is limited to booted iOS Simulators with supported Xcode tooling.
- OpenAPI import does not yet support external references, callbacks, GraphQL, gRPC, or sophisticated example generation.
- All framework integrations are optional observability SDKs; they are not required for the API Sandbox.
- SQLite retention/pruning policy is not implemented yet; schema migrations currently cover versions 1 and 2.

See [ROADMAP.md](ROADMAP.md) for planned milestones. Contributions are welcome; read [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
