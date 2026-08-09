# Contributing to MobileLab

MobileLab is in early development. Start with an issue describing the mobile scenario or capability gap before making a large change.

## Development

Use Go 1.23 or newer, then run:

```sh
make test
make lint
make build
```

Optional SDK work uses Flutter 3.44.4/Dart 3.12 and Node 22 with TypeScript 5.9. Run both package suites with:

```sh
make sdk-test
```

Keep domain packages independent of CLI, HTTP, YAML, SQLite, `adb`, `simctl`, and framework code. Platform behavior belongs in adapters and must report unsupported capabilities honestly. Add focused tests for parser, fault, security, and process-command changes. Tests must pass without Android SDK or Xcode unless explicitly marked as opt-in integration tests.

Do not commit real credentials, tokens, captured production requests, generated state files, or application customer data.
