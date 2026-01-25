# adapter/s3client

The s3client adapter integrates with AWS S3 (or compatible services such as LocalStack). It supports reading and writing structured data, including Parquet, and provides helpers for listing and filtering objects.

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

## References

- examples: example/adapter/s3_adapter/
- builder: pkg/builder/s3client_adapter.go
- internal contracts: pkg/internal/types/s3_adapter.go
