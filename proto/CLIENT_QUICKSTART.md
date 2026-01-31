# Client Quickstart (Interop)

This is the shortest path for an external client to prove full interoperability with Electrician.
If you can complete the gRPC + QUIC checks below, you are fully integrated.

## 0) Compile the proto (required)

Generate types from `proto/electrician_relay.proto` in your language.

See `proto/README.md` for protoc instructions.

## 1) Start the mock OAuth server (required for local testing)

From repo root:

```bash
go run ./example/auth/mock_oauth_server
```

Defaults:
- Issuer: `auth-service`
- Audience: `your-api`
- Scope: `write:data`
- Client ID/Secret: `steeze-local-cli` / `local-secret`
- JWKS: `https://localhost:3000/api/auth/oauth/jwks.json`
- Token: `https://localhost:3000/api/auth/oauth/token` (client credentials)

Shared test values used by the secure examples:
- AES-256 key (hex):
  `ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec`
- Static tenant header: `x-tenant: local`

## 2) Secure gRPC relay (required)

Start the Go receiver:

```bash
go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes
```

Start the Go sender:

```bash
go run ./example/relay_example/secure_advanced_relay_a_oauth_mtls_aes
```

Your integration is valid when both directions work:
- Your client sends → Go receiver accepts and decodes.
- Go sender sends → your client receives and decodes.

Required gRPC details:
- Service: `electrician.RelayService`
- Method: `Receive(WrappedPayload)` (unary) or `StreamReceive(stream RelayEnvelope)`
- TLS 1.3 with certs from `example/relay_example/tls/` (examples configure client certs)
- Metadata headers:
  - `authorization: Bearer <JWT>`
  - `x-tenant: local`
- JWT claims:
  - `iss` = `auth-service`
  - `aud` includes `your-api`
  - `scope` includes `write:data`
- Encryption: AES-GCM, payload bytes are `IV(12 bytes) || ciphertext+tag`
- Compression: Snappy when `metadata.performance.use_compression = true`

## 3) Secure QUIC relay (required)

Start the Go receiver:

```bash
go run ./example/relay_example/quic_secure_oauth_aes_receiver
```

Start the Go sender:

```bash
go run ./example/relay_example/quic_secure_oauth_aes_sender
```

Required QUIC details:
- ALPN: `electrician-quic-relay-v1`
- Frame format: 4-byte big-endian length prefix + protobuf bytes
- Message: `RelayEnvelope` (from `proto/electrician_relay.proto`)
- TLS 1.3 with CA from `example/relay_example/tls/`
- StreamOpen defaults headers must include:
  - `authorization: Bearer <token>`
  - `x-tenant: local`
- AES-GCM payload format matches gRPC section

Reference: `example/relay_example/quic/README_RUST.md`

## 4) Payload encodings + type registry (required)

Do not hardcode `Feedback`. Treat payloads as generic and decode by `payload_type` + encoding.

Recommended registry pattern:
- Key: fully qualified type name (e.g., `acme.TelemetryEvent`)
- Value: decoder for that type

Encodings you must support for full parity:
- **PROTO**: `payload_encoding = PAYLOAD_ENCODING_PROTO`, `payload_type` set to the
  fully qualified message name, and `payload` contains serialized protobuf bytes.
- **JSON**: set `metadata.content_type = application/json` and send JSON bytes.
- **GOB**: `payload_encoding = PAYLOAD_ENCODING_GOB` with Go gob bytes.

Notes:
- Go senders use **GOB by default**. Full interop with Go senders requires GOB support.
- If you are not ready to implement GOB, start by sending JSON/PROTO into the Go receivers,
  then add GOB later for full parity.

## 5) Bonus: UDS (optional)

If your gRPC client/server can also run over Unix domain sockets, that’s extra credit.
We do not ship a UDS example yet.

## Success criteria

You are fully integrated if:
- gRPC secure relay works both directions (your client ↔ Go examples).
- QUIC secure relay works both directions (your client ↔ Go examples).
- You support PROTO, JSON, and GOB payload encodings.

## Troubleshooting

- `issuer mismatch`: JWT `iss` must match `auth-service`.
- `unauthorized`: check `aud`, `scope`, and the `authorization` header.
- TLS errors: trust the CA in `example/relay_example/tls/` or allow self-signed for local dev.
