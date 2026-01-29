# Postgres Adapter (Secure Defaults)

This adapter provides **generic** Postgres read/write support with strong defaults
for enterprise usage. It stores payloads as JSON bytes (or raw bytes) and can
optionally encrypt payloads client‑side with AES‑GCM.

## Security baseline (defaults)

- **TLS required** (`sslmode=require` or `verify-full`)
- **Client‑side encryption** required when a key is provided

To force TLS + client‑side encryption:

```
builder.PostgresAdapterWithSecureDefaults[YourType]("<32-byte-hex-key>")
```

## Driver requirement

The adapter uses `database/sql` and **does not import a driver**.
Your app must register one, for example:

```
import _ "github.com/jackc/pgx/v5/stdlib" // driver "pgx"
```

## Examples

- Writer: `example/adapter/postgres_adapter/secure_writer/main.go`
- Reader: `example/adapter/postgres_adapter/secure_reader/main.go`

## Default table schema

When auto‑create is enabled, the adapter uses a safe default schema:

```
CREATE TABLE IF NOT EXISTS electrician_events (
  id TEXT PRIMARY KEY,
  created_at TIMESTAMPTZ NOT NULL,
  trace_id TEXT,
  payload BYTEA NOT NULL,
  payload_encrypted BOOLEAN NOT NULL DEFAULT false,
  metadata JSONB,
  content_type TEXT,
  payload_type TEXT,
  payload_encoding TEXT
)
```

You can override with `CreateTableDDL`.

## TLS enforcement

For `sslmode=disable|allow|prefer`, the adapter will **hard‑fail** by default.
To allow insecure connections (not recommended):

```
builder.PostgresAdapterWithAllowInsecure[YourType](true)
```
