# Embedded business data API

MobileLab can expose an optional persistent REST API backed by a project-local SQLite file. It requires no database server, Docker container, account, driver installation, or framework SDK.

The feature uses two files with deliberately different responsibilities:

- `mobilelab/data.yaml` declares business resources and their HTTP paths.
- `mobilelab/data.db` stores their mutable JSON documents and is ignored by Git.

MobileLab's internal history remains in `mobilelab/mobilelab.db`. Business data never shares tables with request records, app events, or scenario runs.

## Create a database

From a project that already has `mobilelab.yaml`:

```sh
mobilelab db init
```

This creates a starter `mobilelab/data.yaml`, `mobilelab/seeds/items.json`, and `mobilelab/data.db`. Edit the generated resource or replace it with your own:

```yaml
schema_version: 1
resources:
  products:
    path: /api/products
    id: id
    seed: seeds/products.json
  profile:
    path: /api/profile
    singleton: true
    seed: seeds/profile.json
  private-notes:
    path: /api/private-notes
    id: id
    protected: true
```

A collection seed is an array of JSON objects. Every seeded object needs the configured string ID field:

```json
[
  { "id": "product-1", "name": "Keyboard", "price": 79.9 }
]
```

A singleton seed is one JSON object. Seed paths are relative to `mobilelab/` and cannot escape that directory through `..` or symlinks.

## Generated REST routes

Collection resources provide:

| Method | Path | Result |
| --- | --- | --- |
| `GET` | `/api/products` | List documents |
| `POST` | `/api/products` | Create a document; generates a string ID when omitted |
| `GET` | `/api/products/{id}` | Read one document |
| `PUT` | `/api/products/{id}` | Replace or create one document |
| `PATCH` | `/api/products/{id}` | Merge top-level fields into an existing document |
| `DELETE` | `/api/products/{id}` | Delete one document (`204`) |

Singleton resources provide `GET`, `PUT`, and `PATCH` at their declared path. Bodies must be JSON objects no larger than 1 MiB. Collection IDs are strings. A duplicate `POST` returns `409`; a missing item returns `404`; unsupported methods return `405` with an `Allow` header.

Database routes participate in the same MobileLab request history, redaction, recording, global latency, forced errors, random error rate, and optional bearer-token protection as fixture routes. A static endpoint and database resource cannot own the same method/path; startup rejects the conflict.

## Lifecycle commands

```sh
mobilelab db status
mobilelab db status --json
mobilelab db seed
mobilelab db reset
mobilelab start
```

- `start` creates `data.db` when necessary and applies seeds only to empty resources. Existing changes survive restarts.
- `db seed` upserts the declared seed documents by ID without deleting other documents.
- `db reset` deletes every business document and then reapplies all seeds. This is intentionally destructive for `data.db` only.
- `db status` reports paths, resource kinds, and document counts.

The declarative contract is available from every release binary:

```sh
mobilelab schema data > mobilelab-data-v1.schema.json
```

Do not commit `data.db`; commit `data.yaml` and synthetic seed JSON instead. Never place production exports, credentials, personal data, or real access tokens in a MobileLab project.

## Framework usage

Flutter, React Native, Kotlin, Swift, and Capacitor use these routes as ordinary HTTP APIs. No MobileLab SDK is required for CRUD. Resolve the host URL with `mobilelab endpoint`; standard Android Emulator uses `10.0.2.2`, while iOS Simulator normally reaches host loopback.

The runnable [Flutter](../examples/apps/flutter/README.md) and [React Native](../examples/apps/react-native/README.md) shops include equivalent `data.yaml` files. Their profile, catalog, businesses, cart items, purchase history, and owned products are backed by SQLite; authentication, payments, and per-business catalog views remain deterministic fixture mocks to demonstrate that both modes compose in one project.
