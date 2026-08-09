# MobileLab

MobileLab is a local-first API sandbox, device adapter layer, and portable scenario runner for mobile development. It lets Flutter, React Native, Android, iOS, and Capacitor applications exercise repeatable failure and device scenarios without a real backend, cloud account, Docker, or a framework SDK.

> Status: **v0.1.0** local scenario foundation. See [Current limitations](#current-limitations) and the [changelog](CHANGELOG.md).

## Why MobileLab?

Mobile applications fail at boundaries: expired sessions, slow responses, malformed backends, deep links, device location, and unavailable networks. MobileLab brings those boundaries into one local CLI and one framework-neutral scenario format:

> Write the mobile scenario once. Run it everywhere.

The Go core is a single process, binds to `127.0.0.1` by default, and does not require Node.js or Docker. Optional framework SDKs will add advanced integrations later; basic HTTP mocks do not need them.

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
```

Android uses the official `adb` and Android Emulator executables. MobileLab resolves them from `PATH`, `ANDROID_HOME`, or `ANDROID_SDK_ROOT`. Configured, inactive AVDs are listed with IDs prefixed by `avd:` and can be started explicitly with `device boot`; a running AVD is represented by its attached `emulator-*` serial instead. Deep links use `adb shell am start`; location uses `adb emu geo fix` and is therefore limited to running emulators. Explicit `device clear` uses `adb shell pm clear` and deletes the selected application's local data. MobileLab does not currently claim device-wide Android network conditioning.

iOS uses `xcrun simctl` and is available only on macOS with Xcode. It supports discovery, simulator boot, app launch/terminate, deep links, and simulator location. On iOS, explicit `device clear` uses `simctl uninstall`, so the selected app is removed rather than merely having its data reset. `simctl` does not provide a portable network-conditioning command, so that capability is reported unavailable.

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

Unit tests cover strict configuration, project detection, secret redaction, fixture confinement, auth, faults, routing, OpenAPI generation, device command construction, scenario parsing, and scenario execution. An integration test opens a temporary loopback server and verifies start/status/API fault/reset/stop. CI runs build, tests, and vet on Linux, macOS, and Windows without requiring Android or Xcode.

## Current limitations

- Device selection is first-ready unless `--device` is supplied to `run`; direct deep-link/location commands do not yet expose selection flags.
- Android/iOS network conditioning and local push delivery are not implemented and are reported unavailable.
- OpenAPI import does not yet support external references, callbacks, GraphQL, gRPC, or sophisticated example generation.
- Framework SDKs and runnable Flutter/React Native/native example apps are not part of this first slice.
- SQLite retention/pruning policy and migrations beyond schema v1 are not implemented yet.

See [ROADMAP.md](ROADMAP.md) for planned milestones. Contributions are welcome; read [CONTRIBUTING.md](CONTRIBUTING.md) and [SECURITY.md](SECURITY.md).

## License

MIT. See [LICENSE](LICENSE).
