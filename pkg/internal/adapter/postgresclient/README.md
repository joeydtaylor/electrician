# adapter/postgresclient

The postgresclient adapter provides **secure, generic** read/write integration
with PostgreSQL using `database/sql`. It stores payloads as JSON (or raw bytes)
and supports optional client‑side AES‑GCM encryption before insert.

## Responsibilities

- Connect to Postgres with TLS requirements enforced.
- Batch inserts for throughput.
- Optional table auto‑create with a default schema.
- Read back rows and decode into typed structs.

## Security defaults

- TLS is **required by default** (you must set `sslmode=require` or `verify-full`).
- If a client‑side key is provided, AES‑GCM is required.

Use `builder.PostgresAdapterWithSecureDefaults(...)` to enforce TLS + CSE.

## Example usage

See:

- `example/adapter/postgres_adapter/secure_writer/main.go`
- `example/adapter/postgres_adapter/secure_reader/main.go`

## Notes

- The adapter does **not** import a driver. Your application must register one
  (e.g., `pgx` or `lib/pq`).
- Table schema can be customized via `CreateTableDDL`.
