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

Capability probing, explicit device selection across every command, emulator lifecycle, safe app clearing, technically verified platform-specific network controls, and local push paths where official tools support them.

## 0.3 — Framework integrations

Optional Flutter and React Native SDKs for advanced lifecycle/assertion hooks. API Sandbox remains SDK-free.

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
