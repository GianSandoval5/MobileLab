# Security Policy

## Supported versions

Security fixes target the latest stable 1.x release and the main development branch.

## Reporting a vulnerability

Do not open a public issue containing exploit details, credentials, or sensitive request captures. Contact the maintainers privately through the repository security advisory workflow once the public repository is available.

## Local security model

MobileLab is a development tool and binds to `127.0.0.1` by default. Its local auth sandbox is not a production identity provider. Runtime control uses a random token stored in an owner-readable state file. Dashboard and WebSocket endpoints reject non-loopback clients even when mocks are explicitly exposed. Request metadata and SDK event attributes are redacted before SQLite storage, but users should still use synthetic data and must not configure or report real secrets.

The SDK event endpoint is intentionally reachable by the application under test and therefore does not receive the private CLI control token. It accepts only a strict versioned schema with a 64 KiB request limit. If a physical device needs access, bind MobileLab only on a trusted development network and remove debug cleartext/local-network exceptions from production application configurations.

The optional business data API stores synthetic JSON documents separately in `mobilelab/data.db`. Request bodies and seed files are limited to 1 MiB, seed paths are confined beneath `mobilelab/` including symlink resolution, and SQLite statements bind document values as parameters. A resource marked `protected` uses the development-only local JWT sandbox; it is not production access control. Do not import production databases, personal information, credentials, or real customer data. `mobilelab db reset` deliberately deletes only business documents before restoring declared seeds.

## Plugin trust model

Plugins are trusted project-local executables and run only after an explicit `mobilelab plugin run`. Discovery and inspection parse manifests, confine resolved executable paths to the plugin directory, and compute a SHA-256 fingerprint without executing code. Invocation uses no shell, passes a minimal environment rather than arbitrary parent variables, enforces a deadline, limits each protocol message to 1 MiB, and requires a correlated strict `mobilelab.plugin/v1` response.

These controls are not an operating-system sandbox. An invoked plugin still has the user's filesystem and network permissions, and a SHA-256 fingerprint identifies bytes but does not prove their publisher. MobileLab 1.0 intentionally provides no remote registry, downloader, automatic update, or signature claim. Review third-party code and compare an independently obtained fingerprint before placing it in `mobilelab/plugins`.

## Schema migration safety

Configuration and scenario inputs are limited to one 1 MiB YAML document and reject unknown fields or newer unsupported schema versions. `mobilelab migrate` rejects a symlinked configuration, does not traverse scenario symlinks, preflights every target before writing, preserves file modes/comments, and replaces changed files atomically. Review migration diffs before committing them; migrations do not make unsafe scenario actions trustworthy.
