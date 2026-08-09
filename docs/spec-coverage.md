# PRODUCT_SPEC coverage

This document records the honest MobileLab 1.0 coverage of `PRODUCT_SPEC.md`. “Future” means the specification itself described the item as progressive, eventual, optional, or outside the first local MVP; it is not advertised as implemented.

## Implemented

- Sections 1–4, 48, 63–64: one local-first Go Core, clean inward dependencies, framework-neutral scenarios, adapters/capabilities, no mandatory Node, Docker, cloud account, or backend.
- Sections 5–13, 31, 33–37, 61: CLI lifecycle, strict YAML, API mocks, fixtures/variables, runtime latency/errors, local JWT login/refresh/expiry, sanitized request capture, safe shutdown/control, and actionable errors.
- Sections 15–21, 26–27, 49: real ADB/Android Emulator and conditional `simctl` adapters, device lifecycle, deep links, location, honest network capability probing, and technically valid local iOS Simulator push.
- Sections 22, 42–47: optional Flutter, React Native, Kotlin, Swift, and Capacitor SDK packages with runnable/minimal examples around the same protocol and sandbox. Basic Sandbox use remains SDK-free.
- Sections 23–25: embedded dashboard, typed WebSocket events, SQLite request/scenario/app-event persistence, and transactional migrations.
- Sections 28–29, 38: terminal/JSON/JUnit/HTML reports, headless multi-provider CI examples, complete suite execution, record, and replay.
- Sections 39–41: project-local out-of-process plugin protocol/SDK, Semantic Versioning, fake adapters, unit/integration tests, and cross-platform CI.
- Sections 50–60, 62, 65: the complete mandatory v0.1 definition of done, MIT/open-source files, Make targets, dependency discipline, release packaging, documentation, and incremental validation.
- Mobile endpoint guidance from section 43 is implemented by `mobilelab endpoint`, including the standard Android Emulator `10.0.2.2` alias, iOS Simulator loopback, JSON output, and explicit physical-device limitations.
- MobileLab 1.0 adds stable versioned configuration/scenario contracts, embedded JSON Schemas, safe project migrations, and the compatibility policy.

## Partial by platform reality

- Section 14: scenarios are portable across Android, iOS, and the deterministic fake adapter, with framework-specific observability expressed only as optional event assertions. Framework names are not fake device platforms, and `--all` is not implemented.
- Sections 18 and 21: Android Emulator latency/speed is partial, reliable cross-transport offline mode is unavailable, iOS network shaping is unavailable, and generic Android push is unavailable without an app-specific/FCM path.
- Section 23: the dashboard covers Core/fault status, requests, responses, scenarios, and app events, but it is intentionally not yet a full device/log management console.
- Section 30: operational errors and typed realtime events are observable, but a public structured logging/`--verbose` contract is not yet implemented.
- Section 32: Core is measured continuously by tests/build size and remains a single process, but a formal benchmark/performance budget suite is not yet published.
- Sections 43–47: each SDK includes a minimal consumable example; the repository does not duplicate five full standalone production application projects.

## Future or intentionally absent

- GraphQL, gRPC, and SSE transports (section 8), advanced/full OpenAPI generation and external references (section 9), and enterprise OAuth2/OIDC/PKCE/SSO providers (section 10).
- A device farm, automatic `--all` platform orchestration, and cloud infrastructure.
- Firebase, Supabase, GraphQL, and gRPC first-party plugins until demand justifies their credentials, maintenance, and trust model.
- A remote plugin registry, downloader, implicit execution, automatic updater, or claim that local plugin processes are OS-sandboxed.
- Docker as a requirement. Any future container integration remains optional.

These omissions are reflected in capability output and the README rather than represented by empty methods or false success.
