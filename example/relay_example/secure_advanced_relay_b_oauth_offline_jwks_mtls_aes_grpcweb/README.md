# gRPC-Web Relay Receiver (Frontend Guide)

This example runs a gRPC-Web–compatible receiving relay on `https://localhost:50051`.
Your frontend can call the `RelayService` methods from the browser and push payloads into the
Electrician pipeline.

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

## Minimal payload shape

`WrappedPayload` must include:
- `payload` (bytes)
- `metadata.content_type` (e.g. `application/json` if sending JSON)
- Optional: `payload_encoding` (`PAYLOAD_ENCODING_GOB` or `PAYLOAD_ENCODING_PROTO`)

If `metadata.content_type` is JSON, the receiver will decode JSON by default.

## Encryption and compression (optional)

The receiver will only decrypt/decompress when the metadata says to do so:

- Encryption:
  - Set `metadata.security.enabled = true`
  - Set `metadata.security.suite = ENCRYPTION_AES_GCM`
  - Encrypt the payload with AES-GCM using the shared 32-byte key.

- Compression:
  - Set `metadata.performance.use_compression = true`
  - Set `metadata.performance.compression_algorithm`
  - Compress the plaintext before encrypting.

If you skip these options, send plaintext JSON bytes and the receiver will accept it.

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

## Common pitfalls

- CORS: lock down `AllowedOrigins` for production.
- TLS: browsers will reject self-signed certs unless the CA is trusted.
- Auth: if you need strict JWT validation, add a server-side validator (introspection is built in).
- Payload encoding: JSON is easiest for browser clients; proto/gob requires extra work.
