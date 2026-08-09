# MobileLab Roadmap

The roadmap follows vertical, testable slices. Capabilities appear in CLI output only after a real adapter implements them.

## 0.1 — Local scenario foundation

- [x] Go CLI, strict YAML config, project init/detection, doctor
- [x] HTTP mocks, fixtures, variables, latency/errors, local JWT auth
- [x] request capture and secret redaction
- [x] portable scenario parser/runner, assertions, fake adapter, JSON reports
- [x] basic `adb` and `simctl` device adapters, deep links, emulator/simulator location
- [x] basic OpenAPI 3 importer and embedded dashboard
- [x] SQLite request/scenario repositories
- [x] typed WebSocket event stream and live dashboard request table
- [x] scenario list/run/history commands
- [x] correlated request/response assertions and parameterized-path matching
- [x] install/release packaging with cross-platform binaries and checksums

Framework SDKs and complete runnable framework applications intentionally begin in 0.3/0.4; the 0.1 core remains framework-agnostic and SDK-free.

## 0.2 — Device engine

- [x] explicit platform/device selection for direct device commands
- [x] app launch, stop, and explicit data clearing/uninstall
- [x] iOS Simulator boot from a detected shutdown state
- [x] Android AVD discovery and emulator startup by AVD name
- [x] richer device information and runtime capability probes
- [x] technically verified platform-specific network controls
- [x] local push paths where official tools support them

## 0.3 — Framework integrations

- [x] versioned framework-neutral SDK event protocol and persistent Core ingestion
- [x] optional `mobilelab_flutter` lifecycle, marker, and assertion hooks
- [x] optional `@mobilelab/react-native` lifecycle, marker, and assertion hooks
- [x] scenario assertions over app events emitted after a run starts
- [x] framework tooling diagnostics, package-level tests, examples, and CI

The API Sandbox remains SDK-free; these packages add only advanced observability hooks.

## 0.4 — Native and Capacitor integrations

Optional Kotlin, Swift, and Capacitor packages plus equivalent example applications consuming the same sandbox and scenarios.

## 0.5 — Record and replay

Typed recorder ports, request/environment/deep-link capture, deterministic replay, and scenario editing workflows.

## 0.6 — CI reporting

JUnit XML and HTML reports, GitHub/GitLab/Azure/Jenkins examples, artifact retention, and headless device workflows.

## 0.7 — Plugin ecosystem

Versioned plugin protocol and first-party integrations selected by demonstrated demand. Plugins must not weaken local-first installation.

## 1.0 — Stable platform

Stable scenario/config schemas, compatibility policy, migrations, long-lived plugin API, complete documentation, and production-quality cross-platform release distribution.
