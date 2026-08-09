# Security Policy

## Supported versions

MobileLab has not released a stable version. Security fixes currently target the main development branch.

## Reporting a vulnerability

Do not open a public issue containing exploit details, credentials, or sensitive request captures. Contact the maintainers privately through the repository security advisory workflow once the public repository is available.

## Local security model

MobileLab is a development tool and binds to `127.0.0.1` by default. Its local auth sandbox is not a production identity provider. Runtime control uses a random token stored in an owner-readable state file. Dashboard and WebSocket endpoints reject non-loopback clients even when mocks are explicitly exposed. Request metadata and SDK event attributes are redacted before SQLite storage, but users should still use synthetic data and must not configure or report real secrets.

The SDK event endpoint is intentionally reachable by the application under test and therefore does not receive the private CLI control token. It accepts only a strict versioned schema with a 64 KiB request limit. If a physical device needs access, bind MobileLab only on a trusted development network and remove debug cleartext/local-network exceptions from production application configurations.

## Plugin trust model

Plugins are trusted project-local executables and run only after an explicit `mobilelab plugin run`. Discovery and inspection parse manifests, confine resolved executable paths to the plugin directory, and compute a SHA-256 fingerprint without executing code. Invocation uses no shell, passes a minimal environment rather than arbitrary parent variables, enforces a deadline, limits each protocol message to 1 MiB, and requires a correlated strict `mobilelab.plugin/v1` response.

These controls are not an operating-system sandbox. An invoked plugin still has the user's filesystem and network permissions, and a SHA-256 fingerprint identifies bytes but does not prove their publisher. MobileLab v0.7 intentionally provides no remote registry, downloader, automatic update, or signature claim. Review third-party code and compare an independently obtained fingerprint before placing it in `mobilelab/plugins`.
