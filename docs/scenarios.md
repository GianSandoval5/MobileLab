# Scenario schema v1

Scenario v1 is framework-neutral and starts with `schema_version: 1`:

```yaml
schema_version: 1
name: Payment with expired session
backend:
  latency: 2000
  error: 500
auth:
  token: expired
steps:
  - launch_app
  - open_deeplink: myapp://payments
expect:
  - request:
      method: POST
      path: /payments
  - response:
      status: 500
```

Supported steps are `launch_app`, `stop_app`, `open_deeplink`, `set_latency`, `set_error`, `reset_api`, `expire_auth`, `reset_auth`, and `wait_for_http`. Assertions cover a request, response status, or SDK app event. Each assertion defines exactly one of those forms.

The parser rejects unknown fields, multiple YAML documents, inputs above 1 MiB, unsupported actions, malformed waits, invalid status values, and future schema versions. Directory runs recursively preflight every YAML file before the first scenario executes.

The authoritative contract is [`schemas/mobilelab-scenario-v1.schema.json`](../schemas/mobilelab-scenario-v1.schema.json), also available from the binary:

```sh
mobilelab schema scenario > mobilelab-scenario-v1.schema.json
```

Legacy scenarios without a version remain readable in MobileLab 1.x and can be upgraded with `mobilelab migrate`.
