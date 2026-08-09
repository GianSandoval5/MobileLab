# Configuration schema v1

MobileLab reads one strict YAML document from `mobilelab.yaml`. Version 1 documents begin with:

```yaml
schema_version: 1
project:
  name: my-mobile-app
server:
  host: 127.0.0.1
  port: 4566
dashboard:
  enabled: true
sandbox:
  latency: 0
  error_rate: 0
auth:
  enabled: true
device:
  auto_detect: true
endpoints: []
```

Unknown keys, multiple YAML documents, files larger than 1 MiB, invalid ports/statuses/methods, duplicate route signatures, negative latency, and conflicting response `body`/`fixture` values are rejected before Core starts.

`server.host` defaults to loopback. A network bind must be explicit and produces a security warning. Use synthetic values only; configuration is not a secret store.

The authoritative machine-readable contract is [`schemas/mobilelab-config-v1.schema.json`](../schemas/mobilelab-config-v1.schema.json). Editors and CI can obtain the identical schema from a released binary:

```sh
mobilelab schema config > mobilelab-config-v1.schema.json
```

Legacy pre-1.0 files remain readable. Use `mobilelab migrate --check` to enforce the stable version in CI and `mobilelab migrate` to upgrade a project.
