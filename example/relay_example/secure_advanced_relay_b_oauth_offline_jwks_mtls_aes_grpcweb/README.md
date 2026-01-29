# gRPC-Web Relay Receiver (Frontend Guide)

This example runs a gRPC-Web-compatible receiving relay on `https://localhost:50051`.
Use it to send encrypted or plaintext payloads from a browser client (Connect).

Source of truth for gRPC-web alignment:
- `docs/relay-grpcweb-alignment.md`

## 60-second quickstart (copy/paste)

1) Start the mock OAuth server (dev):

```bash
go run ./example/auth/mock_oauth_server
```

2) Start the gRPC-Web receiver (match issuer):

```bash
OAUTH_ISSUER_BASE=auth-service \
  OAUTH_JWKS_URL=https://localhost:3000/api/auth/oauth/jwks.json \
  go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

3) Use your browser client to send:
   - `authorization: Bearer <token>`
   - `x-tenant: local`
   - JSON payload (`content_type: application/json`)

If you see `issuer mismatch`, your token `iss` does not match the receiver's issuer.
Check the receiver log line **Auth JWT validator installed** for the exact `issuer`.

---

## What the server expects

Endpoint:
- `https://localhost:50051` (TLS enabled by default)

Service:
- `electrician.RelayService` from `proto/electrician_relay.proto`

Methods:
- `Receive(WrappedPayload) -> StreamAcknowledgment` (unary)
- `StreamReceive(stream RelayEnvelope) -> stream StreamAcknowledgment` (streaming)

Metadata headers (gRPC-Web):
- `authorization: Bearer <token>` (required if auth is enabled)
- `x-tenant: local` (required in this example)

Notes:
- CORS allow-all in the example only affects browser origins. It does not bypass auth or TLS.
- TLS uses the certificates under `example/relay_example/tls/`. Browsers must trust the CA.

## Minimal payload shape (works every time)

`WrappedPayload` must include:
- `payload` (bytes)
- `metadata.content_type` (e.g. `application/json` if sending JSON)
- Optional: `payload_encoding` (`PAYLOAD_ENCODING_GOB` or `PAYLOAD_ENCODING_PROTO`)

If `metadata.content_type` is JSON, the receiver will decode JSON by default.

Recommended:
- `payload_encoding = PAYLOAD_ENCODING_UNSPECIFIED`
- `content_type = application/json`

## Encryption and compression (required in this example)

This receiver is configured with a decryption key, so **encryption is required**.
If `metadata.security` is missing or disabled, the payload will be rejected.

- Encryption:
  - Set `metadata.security.enabled = true`
  - Set `metadata.security.suite = ENCRYPTION_AES_GCM`
  - Encrypt the payload with AES-GCM using the shared 32-byte key.

- Compression:
  - Set `metadata.performance.use_compression = true`
  - Set `metadata.performance.compression_algorithm`
  - Compress the plaintext before encrypting.

If you skip these options, the receiver will reject the payload.

## Browser client outline

1) Generate a gRPC-Web client from `proto/electrician_relay.proto`.
2) Call `RelayService.Receive` with a `WrappedPayload`.
3) Pass gRPC metadata headers for `authorization` and `x-tenant`.

The example server also supports gRPC-Web websocket transport. If your client library
does not use websockets, it will use HTTP/1.1 with gRPC-Web framing automatically.

## Local dev quick start (receiver)

From the repo root:

```sh
go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

## Common pitfalls (fast fixes)

- `issuer mismatch`: token `iss` must match the receiver's issuer (see log line).
- CORS: lock down `AllowedOrigins` for production.
- TLS: browsers will reject self-signed certs unless the CA is trusted.
- Auth: missing `aud` or `scope` claims will fail auth.
- Payload encoding: JSON is easiest for browser clients; proto/gob requires extra work.
