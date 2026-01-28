# Relay Client Integration (Rust + Python + native SDKs)

This guide is for non-browser clients (Rust, Python, Go, etc.). Use it to build
clients that talk to the Electrician relay with the same contract as the Go
examples.

If you are building a browser client, use:
`example/relay_example/FRONTEND_INTEGRATION_UNIFIED.md`

---

## TL;DR (copy/paste checklist)

**Transport**
- gRPC (native) calling `electrician.RelayService/Receive`
- TLS enabled (trust the CA)

**Auth headers (transport metadata)**
- `authorization: Bearer <token>`
- `x-tenant: local`

**Payload (JSON recommended)**
- `payload`: JSON bytes
- `content_type`: `application/json`
- `payload_encoding`: `PAYLOAD_ENCODING_UNSPECIFIED`
- `payload_type`: `electrician.Feedback` (optional)

**If encrypting**
- AES-GCM with 12-byte IV prefix
- Set:
  - `metadata.security.enabled = true`
  - `metadata.security.suite = ENCRYPTION_AES_GCM`

---

## 1) Which endpoint to hit

Use the secure gRPC receiver example:

```bash
OAUTH_ISSUER_BASE=auth-service \
  go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

Defaults:
- Address: `https://localhost:50051`
- JWKS: `https://localhost:3000/api/auth/oauth/jwks.json`
- Required `aud`: `your-api`
- Required `scope`: `write:data`
- Required header: `x-tenant: local`
- AES key (hex): `ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec`

> The example is labeled “gRPC-Web”, but it still runs a native gRPC server.
> Native clients can call it directly over TLS.

---

## 2) Token rules (JWT)

Tokens must contain:
- `iss` = configured issuer (example above uses `auth-service`)
- `aud` includes `your-api`
- `scope` includes `write:data`

To get a token in dev:

```text
POST https://localhost:3000/api/auth/session/token
{ "scope": "write:data" }
```

Use the returned `access_token` as the Bearer token.

If you see `issuer mismatch`, the token `iss` does not match the receiver’s
expected issuer. Check the receiver log line **Auth JWT validator installed**.

---

## 3) TLS / CA trust

The receiver uses TLS. Your client must trust:
- `example/relay_example/tls/ca.crt`

In Rust/Python, configure the gRPC channel with that CA certificate.

---

## 4) Payload format (JSON)

For JSON payloads:
- `payload`: UTF-8 JSON bytes
- `content_type`: `application/json`
- `payload_encoding`: `PAYLOAD_ENCODING_UNSPECIFIED`

Example JSON:

```json
{
  "customerId": "cust-123",
  "content": "Relay test payload",
  "category": "feedback",
  "isNegative": false,
  "tags": ["native", "grpc"]
}
```

---

## 5) Encryption (AES-GCM)

Encrypt **before** sending if enabled. Format:

```
IV (12 bytes) || ciphertext+tag
```

Metadata flags:
- `metadata.security.enabled = true`
- `metadata.security.suite = ENCRYPTION_AES_GCM`

Key is a 32-byte hex string (AES-256). Example in secure receiver:
`ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec`

---

## 6) QUIC option (native only)

If you need QUIC, see:
- `example/relay_example/quic/README.md`
- `example/relay_example/quic/README_RUST.md`

QUIC uses a length-prefixed `RelayEnvelope` frame:
```
[u32 length BE] + [protobuf bytes]
```

---

## 7) Quick sanity test (gRPC)

If your client can make a single unary call and receives a `StreamAcknowledgment`
with `success=true`, the integration is correct.

Common failures:
- `issuer mismatch`: token issuer does not match receiver
- `missing bearer token`: no `authorization` header
- `unauthenticated`: missing `x-tenant` header
- TLS errors: CA not trusted
