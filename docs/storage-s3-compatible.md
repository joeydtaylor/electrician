# S3-Compatible Storage (Storj, MinIO, LocalStack)

Electrician uses the AWS SDK v2 S3 client under the hood. Any S3-compatible
provider that implements the standard APIs (PutObject, GetObject, ListObjectsV2)
will work, including **Storj S3 Gateway**, **MinIO**, and **LocalStack**.

## What is supported

- Read/write via the S3 adapter (Parquet, NDJSON, raw bytes).
- Standard object operations (list, put).
- Prefix-based partitioning and rolling writers.

## What is not required (or not used)

- S3 Select is **not** used by the adapter today.
- STS assume-role is AWS-specific; use static credentials for most S3-compatible providers.
- SSE-KMS is AWS-specific; use AES-GCM at the relay layer or provider-native encryption.

## Quick start: Storj S3 Gateway

```go
cli, err := builder.NewS3ClientStorj(ctx, builder.StorjS3Config{
    AccessKey: "<storj-access-key>",
    SecretKey: "<storj-secret-key>",
    RequireTLS: builder.BoolPtr(true),
    // Endpoint and Region default to gateway.storjshare.io and us-east-1.
})
if err != nil {
    log.Fatal(err)
}

adapter := builder.NewS3ClientAdapter[Feedback](
    ctx,
    builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, "my-bucket"),
    builder.S3ClientAdapterWithFormat[Feedback]("parquet", ""),
    builder.S3ClientAdapterWithStorjSecureDefaults[Feedback]("<32-byte-hex-key>"),
)
```

For HIPAA‑grade guidance (TLS + client‑side encryption + SSE), see:
`docs/storage-storj-secure.md`

## Generic S3-compatible endpoint

```go
cli, err := builder.NewS3ClientStaticCompatible(ctx, builder.S3CompatibleStaticConfig{
    Endpoint:  "http://localhost:4566", // LocalStack / MinIO / other gateway
    Region:    "us-east-1",
    AccessKey: "test",
    SecretKey: "test",
    // ForcePathStyle defaults to true; override if your provider prefers virtual-hosted.
    // RequireTLS can enforce https-only endpoints.
})
```

## Notes

- For many gateways, **path-style addressing** is required. The helper defaults
  to path-style (compatible with LocalStack/MinIO/Storj). If you need
  virtual-hosted style, set `ForcePathStyle` to `builder.BoolPtr(false)`.
- If you are encrypting payloads at the relay layer (AES-GCM), that is independent
  of S3 storage and works with all S3-compatible providers.
- Client-side encryption (AES‑GCM) is now supported at the S3 adapter level for
  object‑at‑rest protection.
