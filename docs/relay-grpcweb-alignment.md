# Relay gRPC-Web Alignment (Frontend + Backend)

This doc is the single source of truth for **gRPC-web** usage in this repo.
It is intentionally strict so frontend and backend stay aligned.

---

## Must-haves (do not deviate)

**Transport**
- Use Connect (gRPC-web) to call `electrician.RelayService/Receive`.
- Base URL: `https://localhost:50051` (TLS).

**Auth headers (gRPC metadata)**
- `authorization: Bearer <token>`
- `x-tenant: local`

**Token requirements**
- `iss` matches receiver issuer (defaults to `auth-service` in the secure gRPC‑web example)
- `aud` includes `your-api`
- `scope` includes `write:data`

**Payload format (JSON only)**
- `payload`: UTF-8 JSON bytes
- `metadata.content_type`: `application/json`
- `payload_encoding`: `PAYLOAD_ENCODING_UNSPECIFIED`
- `payload_type`: `electrician.Feedback` (optional)

**Encryption (required when server has a key)**
- AES-GCM with 12-byte IV prefix
- `metadata.security.enabled = true`
- `metadata.security.suite = ENCRYPTION_AES_GCM`

---

## Token acquisition (dev)

```text
POST https://localhost:3000/api/auth/session/token
{ "scope": "write:data" }
```

Use `access_token` as the Bearer token.

If you see `issuer mismatch`, the token `iss` does not match the receiver’s
expected issuer. Check the receiver startup log line:
`Auth JWT validator installed`.

### Mock OAuth server (recommended for dev)

Use the built‑in mock server to avoid surprises when your real auth is down:

```bash
go run ./example/auth/mock_oauth_server
```

Defaults:

- Issuer: `auth-service`
- JWKS URL: `https://localhost:3000/api/auth/oauth/jwks.json`
- Session token: `https://localhost:3000/api/auth/session/token`

---

## CORS (browser only)

Your relay endpoint must answer CORS preflight (OPTIONS) requests.
See: `docs/relay-grpcweb-cors.md`

---

## Reference example (server)

```bash
OAUTH_ISSUER_BASE=auth-service \
  OAUTH_JWKS_URL=https://localhost:3000/api/auth/oauth/jwks.json \
  go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

---

## What we do NOT support in the browser

- GOB payloads
- Raw QUIC streams (use native clients or WebTransport)
- Non-TLS HTTP
