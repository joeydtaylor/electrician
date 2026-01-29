# adapter/s3client

The s3client adapter integrates with AWS S3 and S3-compatible services (LocalStack, MinIO, Storj S3 Gateway). It supports reading and writing structured data, including Parquet, and provides helpers for listing and filtering objects.

## Responsibilities

- Configure AWS S3 clients and endpoints.
- Read and write objects with structured formats.
- Support Parquet read/write with configurable options.
- Emit telemetry for read/write operations.

## Key types and functions

- S3ClientAdapter[T]: main adapter type.
- Fetch(): read objects into typed records.
- Write(): write records to S3 with format selection.

## Configuration

Common options include:

- Client, bucket, and prefix
- List and pagination settings
- Format options (JSON, NDJSON, Parquet)
- Parquet settings (schema, memory thresholds)
- Client-side encryption (AES‑GCM) for object‑at‑rest protection
- Sensor and logger

## Observability

Sensors emit metrics for reads, writes, and errors. Loggers capture structured S3 events.

## Usage

```go
adapter := builder.NewS3ClientAdapter[Item](
    ctx,
    builder.S3ClientAdapterWithClientAndBucket[Item](client, "bucket"),
    builder.S3ClientAdapterWithReaderListSettings[Item]("prefix/", ".parquet", 5000, 0),
)
```

## S3-compatible providers (Storj, MinIO, LocalStack)

Use static credentials + endpoint override:

```go
cli, err := builder.NewS3ClientStaticCompatible(ctx, builder.S3CompatibleStaticConfig{
    Endpoint:  "https://gateway.storjshare.io",
    Region:    "us-east-1",
    AccessKey: "<access>",
    SecretKey: "<secret>",
})
```

Storj convenience helper:

```go
cli, err := builder.NewS3ClientStorj(ctx, builder.StorjS3Config{
    AccessKey: "<access>",
    SecretKey: "<secret>",
})
```

Storj secure defaults (client‑side encryption required):

```go
adapter := builder.NewS3ClientAdapter[Item](
    ctx,
    builder.S3ClientAdapterWithClientAndBucket[Item](cli, "bucket"),
    builder.S3ClientAdapterWithStorjSecureDefaults[Item]("<32-byte-hex-key>"),
)
```

Secure Storj guidance:

- `docs/storage-storj-secure.md`

## References

- examples: example/adapter/s3_adapter/
- builder: pkg/builder/s3client_adapter.go
- internal contracts: pkg/internal/types/s3_adapter.go
