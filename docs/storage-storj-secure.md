# Storj S3 Secure Usage (HIPAA‑grade guidance)

This document shows how to use Storj’s S3‑compatible gateway with Electrician’s
S3 adapter **securely**. It is **not legal advice**, and it does **not** certify
HIPAA compliance. HIPAA compliance also requires operational controls and a
BAA with your storage provider.

---

## Security baseline (what we enforce in code)

1) **TLS only** (no HTTP)
2) **Client‑side encryption** before upload (AES‑256‑GCM)
3) **Server‑side encryption (SSE)** headers on write
4) **Least‑privilege credentials** (no wildcards)
5) **Auditable logs** (S3 adapter sensors/loggers)

---

## Code examples (read + write)

Use the new Storj examples:

- Writer: `example/adapter/s3_adapter/storj_secure_writer/main.go`
- Reader: `example/adapter/s3_adapter/storj_secure_reader/main.go`

These examples:
- enforce HTTPS
- use AES‑GCM client‑side encryption
- require SSE headers on write
- use Parquet with Zstd
- default to TLS 1.2+ for Storj HTTP clients
- force signed payload SHA‑256 (avoids `x-amz-content-sha256` mismatches)
- disable automatic AWS SDK request checksums (avoids aws‑chunked trailers)

The adapters now ship a Storj-focused helper that **requires client‑side encryption**
and **SSE headers** by default:

```
builder.S3ClientAdapterWithStorjSecureDefaults[YourType](clientSideKeyHex)
```

If your gateway does not accept SSE headers, you can explicitly disable it
(this makes it less secure; not recommended unless required):

```
builder.S3ClientAdapterWithSSE[YourType]("", "")
builder.S3ClientAdapterWithRequireSSE[YourType](false)
```

---

## Required inputs (replace placeholders)

You must set these values in the example files:

- Storj access key
- Storj secret key
- Bucket name
- Client‑side AES‑GCM key (32‑byte hex)

**Do not** commit credentials to git.

---

## HIPAA‑grade checklist (operational)

These are outside code and required for HIPAA readiness:

- **BAA** with Storj or your storage provider
- **Key management** policy (rotation, escrow, revocation)
- **Access logging** and alerting
- **Least privilege** IAM policies (read vs write separation)
- **Data retention & deletion** policies
- **Incident response** procedures

---

## Encryption details

Client‑side encryption uses:

```
AES‑256‑GCM
IV (12 bytes) || ciphertext+tag
```

Metadata set on objects:

- `x-electrician-cse: aes-gcm`
- `x-electrician-content-type: <original>`
- `x-electrician-content-encoding: <original>` (if gzip)

These are used to decrypt and decode on reads.

---

## Notes

- SSE headers are still sent even with client‑side encryption (defense in depth).
- If Storj rejects SSE headers, disable `RequireSSE` and keep client‑side encryption enabled.
- This doc is guidance only; validate compliance with your security team.
