# S3 Client Adapter

The S3 Client Adapter writes records to S3 and can read them back into `T`. It supports NDJSON by default
and includes optional Parquet handling for bulk payloads. The adapter is typically wired through a `Plug`
and driven by a `Generator`.

## Responsibilities

- Encode records and upload S3 objects with batching.
- Emit telemetry for write attempts and flushes.
- Read objects back into `T` via NDJSON or Parquet.
- Provide optional polling for reader workflows.

## Pipeline Composition

A common arrangement is:

```
S3 Client Adapter -> Plug -> Generator -> Wire
```

## Configuration

Writer configuration includes:

- Prefix and filename templates for object keys.
- Format selection and format options (NDJSON defaults, Parquet options).
- Batching thresholds (records, bytes, max age).
- Optional SSE configuration.

Reader configuration includes:

- Prefix filters and page sizing.
- Polling interval for `Serve`.
- Decoder format and options (gzip for NDJSON, Parquet options).

## Telemetry

Sensors can subscribe to:

- Writer start/stop
- Key rendering
- Put attempts, successes, and errors
- Parquet roll flushes

Loggers receive formatted messages from adapter operations.

## Package Layout

- `s3client.go`: core type and constructor
- `config.go`: dependency and configuration setters
- `connect.go`: logger/sensor/wire wiring
- `helpers.go`: internal helpers
- `lifecycle.go`: stop handling
- `metadata.go`: component metadata accessors
- `notify.go`: logger emission
- `reader.go`: fetch + serve loops
- `parquet_reader.go`: Parquet decode helpers
- `parquet_writer.go`: Parquet roll/encode helpers
- `writer.go`: writer serve loop
- `writer_raw.go`: raw writer loop
- `writer_helpers.go`: retries, flush helpers, key rendering
- `options.go`: functional options
- `*_test.go`: tests

## References

- S3 adapter examples: `example/adapter/s3_adapter/`
