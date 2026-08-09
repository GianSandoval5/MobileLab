# Changelog

All notable changes to MobileLab are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/), and MobileLab follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- Versioned protocol v1 SDK Bridge with strict 64 KiB ingestion, nested secret redaction, SQLite app-event history, realtime dashboard updates, and authenticated scenario access.
- Optional `mobilelab_flutter` and `@mobilelab/react-native` packages with lifecycle, marker, assertion, automatic lifecycle reporting, examples, and package-level tests.
- Portable `app_event` scenario assertions filtered to events emitted after each run begins.
- Flutter/Node toolchain detection, Node 18+ diagnostics, Android Emulator diagnostics, and pinned SDK CI on Flutter 3.44.4 and Node 22.23.2.
- Version-checked Flutter and React Native SDK archives in tag-driven GitHub Releases.

### Fixed

- Updated GitHub artifact actions to their Node 24 generations after the v0.2.0 release warning.
- Avoided the Flutter action's legacy Node 20 cache path while retaining pinned SDK validation.

## [0.2.0] - 2026-08-09

### Added

- Explicit `--platform` and `--device` selection for deep links, location, and new device lifecycle commands.
- `mobilelab device list|info|launch|stop|clear|boot`, including JSON device inspection.
- Android app-data clearing through `adb shell pm clear` and iOS Simulator boot/app uninstall through verified `simctl` commands.
- Android AVD discovery and explicit non-blocking startup by `avd:<name>`, with SDK tool resolution through `PATH`, `ANDROID_HOME`, or `ANDROID_SDK_ROOT`.
- Rich human-readable/JSON device inspection, safe Android runtime metadata, iOS Simulator metadata, and runtime-probed emulator capabilities.
- Capability-gated `mobilelab network slow|online|offline`; Android Emulator delay/speed shaping is intentionally partial while unreliable offline and iOS paths remain unavailable.
- Strict YAML push fixtures and capability-probed local iOS Simulator delivery through `simctl push`, including APNs reserved-key and 4096-byte validation.

### Fixed

- Updated GitHub Actions to their Node 24 generations and made release publication pass the repository explicitly when running without a checkout.
- Device command errors now report the capability actually requested instead of always naming app launch.

## [0.1.0] - 2026-08-09

### Added

- Single-binary Go CLI with `init`, `start`, `stop`, `status`, `doctor`, `detect`, `capabilities`, `api`, `auth`, `deeplink`, `location`, `run`, and `scenario` commands.
- Strict YAML configuration, framework/project detection, safe initialization, fixtures, local variables, parameterized mock routes, endpoint delays, and runtime HTTP fault injection.
- Development-only HS256 JWT login/refresh sandbox with valid, invalid, and forced-expired session behavior.
- Sanitized request inspection with persistent SQLite history, transactional schema migrations, WAL mode, and persisted scenario runs.
- Portable YAML Scenario Engine with fake, Android (`adb`), and iOS Simulator (`xcrun simctl`) device adapters.
- Correlated request/response assertions, OpenAPI-style path parameter matching, terminal output, and JSON reports.
- OpenAPI 3 importer using `kin-openapi`, including generated mocks, schema examples, parameterized routing, and starter scenarios.
- Typed non-blocking event bus and loopback-only live WebSocket dashboard for requests, faults, auth state, and scenario history.
- Cross-platform CI for Linux, macOS, and Windows plus tag-driven release binaries and SHA-256 checksums.
- MIT license, architecture, security, contribution, quick-start, platform limitations, and roadmap documentation.

### Security

- Loopback binding by default with an explicit warning for network exposure.
- Random token-authenticated control endpoint and owner-readable runtime state.
- Redaction of authorization headers, cookies, API keys, passwords, tokens, secrets, nested JSON values, and sensitive query parameters before persistence or event publication.
- Fixture confinement blocks lexical and symlink path traversal.
- Dashboard and WebSocket access remain loopback-only even when mock APIs are explicitly exposed.

### Known limitations

- Android/iOS network conditioning and local push delivery are not yet implemented.
- Full Flutter, React Native, Kotlin, Swift, and Capacitor SDKs/examples are planned for later milestones.
- SQLite retention/pruning and JUnit/HTML reports are not yet available.
- OpenAPI external references, callbacks, GraphQL, gRPC, and advanced example generation are not yet supported.

[Unreleased]: https://github.com/GianSandoval5/MobileLab/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/GianSandoval5/MobileLab/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/GianSandoval5/MobileLab/releases/tag/v0.1.0
