# Relay Integration (Rust Parity Guide)

This guide is the **Rust parity** companion to the Go examples and frontend docs.
It keeps Rust clients aligned with the same relay contract, auth, and encryption
standards used across this repo.

If you only need native gRPC, you can also use:
`docs/relay-client-native.md`

If you need QUIC in Rust, use:
`example/relay_example/quic/README_RUST.md`

If you are building browser UI, use:
`docs/relay-grpcweb-alignment.md`

---

## TL;DR (copy/paste checklist)

**Proto**
- `proto/electrician_relay.proto`

**gRPC method**
- `electrician.RelayService/Receive`

**Metadata headers (gRPC metadata)**
- `authorization: Bearer <token>`
- `x-tenant: local`

**Payload (JSON only)**
- `payload`: UTF‑8 JSON bytes
- `metadata.content_type`: `application/json`
- `payload_encoding`: `PAYLOAD_ENCODING_UNSPECIFIED`
- `payload_type`: `electrician.Feedback` (optional)

**Encryption (required if server has a key)**
- AES‑GCM with 12‑byte IV prefix
- `metadata.security.enabled = true`
- `metadata.security.suite = ENCRYPTION_AES_GCM`

---

## 1) gRPC (native) — recommended baseline

Use the secure gRPC receiver example (also accepts gRPC‑web):

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

Rust example code is in:
`example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb/RUST_INTEGRATION.md`

---

## 2) QUIC (native, low‑latency)

Use the Rust QUIC guide:
`example/relay_example/quic/README_RUST.md`

**Frame format** (stream):
```
[u32 length BE] + [protobuf bytes of RelayEnvelope]
```

**Auth + headers** go in `StreamOpen.defaults.headers`:
- `authorization: Bearer <token>`
- `x-tenant: local`

---

## 3) WebSocket / WSS (native fallback)

Rust can use any WebSocket client (tokio‑tungstenite, tungstenite, etc.).
The server expects:
- Binary frames containing **protobuf-encoded** `RelayEnvelope`
- First message is `StreamOpen`
- Subsequent messages are `WrappedPayload`

Example server you can run:
```bash
go run ./example/relay_example/websocket_basic_receiver
```

Docs:
`docs/relay-last-three.md` (WebSocket section)

---

## 4) WebTransport (QUIC‑like, advanced)

WebTransport is optional and behind a build tag:

```bash
go run -tags webtransport ./example/relay_example/webtransport_basic_receiver
```

Rust can speak HTTP/3 + WebTransport, but it’s an advanced stack.
Use this only if you need QUIC‑like semantics in a browser‑compatible way.

Docs:
`docs/relay-last-three.md` (WebTransport section)

---

## 5) NATS / PubSub (server‑side fanout)

NATS is **server‑side only**. It is not a client transport.

Examples:
```bash
go run -tags nats ./example/relay_example/nats_basic_subscriber
go run -tags nats ./example/relay_example/nats_basic_publisher
```

Docs:
`docs/relay-last-three.md` (NATS section)

---

## Common pitfalls

- `issuer mismatch`: token `iss` doesn’t match `OAUTH_ISSUER_BASE`.
- `missing bearer token`: auth required and `authorization` header not set.
- `x-tenant` must be **transport metadata**, not payload headers.
- TLS: Rust must trust `example/relay_example/tls/ca.crt`.
- Encryption enabled on server means payload must be AES‑GCM encrypted.
