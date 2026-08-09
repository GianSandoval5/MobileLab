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

- [x] extend protocol v1 validation to Android, iOS, and Capacitor without adding framework dependencies to Core
- [x] optional `mobilelab-android` Kotlin/JVM-compatible SDK with HTTP transport, lifecycle, marker, and assertion hooks
- [x] optional `MobileLabKit` Swift Package with async HTTP transport and UIKit lifecycle integration
- [x] optional `@mobilelab/capacitor` package with Capacitor App lifecycle integration
- [x] package-level tests and integration examples consuming the same sandbox and `app_event` scenarios
- [x] pinned Java, Swift, and Node SDK CI plus version-checked GitHub Release archives

All three SDKs remain transport adapters around protocol v1. Applications can continue to use the API Sandbox without installing them.

## 0.5 — Record and replay

- [x] typed, single-active-session recorder owned by Core
- [x] sanitized HTTP request/response, environment mutation, and deep-link capture
- [x] deterministic YAML scenario generation with atomic create/update writes
- [x] `mobilelab record <name>` interactive and duration-limited workflows
- [x] `mobilelab replay <name>` through the existing platform-neutral runner
- [x] recorder unit/integration tests, documentation, CI, and release packaging

## 0.6 — CI reporting

- [x] recursive, preflighted scenario-directory suites with aggregate domain results
- [x] JUnit XML mapping scenarios/checks to standard test suites and cases
- [x] standalone responsive HTML reports with escaped user-controlled content
- [x] configurable CI timeout, output directory creation, complete-suite execution, and failure exit codes
- [x] executable headless fake-device workflow with retained logs and report artifacts
- [x] GitHub Actions, GitLab CI, Azure Pipelines, and Jenkins examples

## 0.7 — Plugin ecosystem

- [x] versioned, language-neutral `mobilelab.plugin/v1` request/response contract
- [x] strict project-local manifests, confined executable resolution, and SHA-256 inspection
- [x] explicit short-lived execution with deadlines, bounded I/O, response correlation, and minimal inherited environment
- [x] `plugin list`, `plugin inspect`, and `plugin run` CLI workflows
- [x] public Go authoring package and executable protocol example in CI
- [x] documented trusted-code boundary with no registry, implicit execution, or mandatory cloud dependency

First-party Firebase, Supabase, GraphQL, or gRPC integrations remain candidates selected by demonstrated demand; v0.7 does not invent credentials or weaken local-first installation to claim them.

## 1.0 — Stable platform

- [x] stable configuration and scenario schema v1 with embedded/published JSON Schemas
- [x] compatibility policy for YAML, CLI JSON, SDK protocol, plugin protocol, and SQLite
- [x] legacy project preflight and atomic `migrate` / CI `migrate --check` workflows
- [x] long-lived `mobilelab.plugin/v1` contract and public Go authoring API
- [x] device-aware `mobilelab endpoint` with honest emulator/simulator/physical-device behavior
- [x] complete contract documentation and honest `PRODUCT_SPEC.md` coverage audit
- [x] production cross-platform binaries, coordinated SDK packages, schema assets, and checksums
