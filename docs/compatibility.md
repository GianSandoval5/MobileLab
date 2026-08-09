# MobileLab compatibility policy

MobileLab 1.0 follows Semantic Versioning for the Core CLI, public YAML contracts, the Go plugin authoring package, and coordinated framework SDK packages.

## Stable contracts

The following interfaces are stable throughout the MobileLab 1.x line:

- configuration documents with `schema_version: 1`;
- scenario documents with `schema_version: 1`;
- SDK event protocol v1;
- `mobilelab.plugin/v1` and the public `pkg/plugin` Go package;
- documented CLI commands, flags, exit status behavior, and JSON field meanings;
- the two JSON Schemas in `schemas/` and printed by `mobilelab schema`.

A valid v1 configuration or scenario remains valid in later 1.x releases. Existing required fields will not change meaning, and existing enum values will not be removed. New optional capabilities may be added in a minor release; documents using them naturally require a Core version that implements them. Human-readable CLI spacing is not a machine contract. Scripts should prefer `--json` where offered; JSON additions are compatible, while existing fields retain their types and meanings during 1.x.

Breaking changes to these contracts require a new major MobileLab version or a separately versioned protocol/schema. Deprecations are documented in the changelog before removal and are not removed during 1.x.

## YAML versions and migration

MobileLab 1.0 emits schema version 1 for `mobilelab.yaml`, recorder output, OpenAPI-generated scenarios, and initialization templates. Legacy pre-1.0 documents without `schema_version` remain readable during 1.x.

Use:

```sh
mobilelab migrate --check
mobilelab migrate
```

The check performs no writes and returns non-zero when a project needs migration, making it suitable for CI. Migration preflights the configuration and every recursive scenario before writing anything, rejects symlinks, preserves YAML comments and file permissions, and replaces each changed file atomically. A newer schema version is rejected with an instruction to upgrade MobileLab instead of being interpreted incorrectly.

## SQLite

SQLite migrations are ordered, transactional, and forward-only. Core automatically upgrades older supported databases. An older Core refuses a database created by a newer unsupported schema version. Downgrades are not automatic; keep project backups when testing prerelease builds.

## Plugins and SDKs

Plugin protocol v1 is language-neutral and remains supported for the MobileLab 1.x line. Manifest and wire-contract changes that break existing plugins require another explicit protocol version. `internal/` Go packages are not public APIs.

Framework SDK releases share the Core release number for traceability. Their HTTP event protocol remains independently identified as protocol v1, so applications do not need a plugin or SDK merely to use the API Sandbox.

## Platform tooling

Android SDK, Emulator, ADB, Xcode, and `simctl` behavior can vary independently of MobileLab. The compatibility guarantee covers MobileLab's capability reporting and refusal to claim unsupported behavior, not undocumented behavior in third-party tool versions.
