# Relay Examples (Index)

This folder groups all relay examples (gRPC, QUIC, secure variants) and the docs that go with them.

## Quickstart (pick one path)

### A) Browser gRPC-Web (fastest)
1) Start the mock OAuth server (dev):
   ```bash
   go run ./example/relay_example/mock_oauth_server
   ```
2) Start the secure gRPC-Web receiver:
   ```bash
   OAUTH_ISSUER_BASE=auth-service \
     go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
   ```
3) Use your browser client (Connect) and send:
   - `authorization: Bearer <token>`
   - `x-tenant: local`
   - JSON payload + `content_type: application/json`

If you see `issuer mismatch`, your token `iss` does not match the receiver's issuer.
Check the receiver log line: `Auth JWT validator installed` for the exact issuer string.

### B) QUIC (native clients)
1) Start the secure QUIC receiver:
   ```bash
   go run ./example/relay_example/quic_secure_oauth_aes_receiver
   ```
2) Send with the secure QUIC sender:
   ```bash
   go run ./example/relay_example/quic_secure_oauth_aes_sender
   ```

### C) Basic (no auth, no encryption)
```bash
go run ./example/relay_example/basic_relay_a
go run ./example/relay_example/basic_relay_b
```

### D) WebSocket (compat browser/legacy)
```bash
go run ./example/relay_example/websocket_basic_receiver
go run ./example/relay_example/websocket_basic_sender
```

### E) WebTransport (browser QUIC semantics)
Build tag required (and `github.com/quic-go/webtransport-go` dependency):
```bash
go run -tags webtransport ./example/relay_example/webtransport_basic_receiver
go run -tags webtransport ./example/relay_example/webtransport_basic_sender
```

### F) NATS (server-side fanout)
Build tag required (and `github.com/nats-io/nats.go` dependency):
```bash
go run -tags nats ./example/relay_example/nats_basic_subscriber
go run -tags nats ./example/relay_example/nats_basic_publisher
```

## Key docs

- Frontend integration (gRPC-Web + QUIC): `example/relay_example/FRONTEND_INTEGRATION_UNIFIED.md`
- QUIC relay docs (including Rust guide): `example/relay_example/quic/README.md`
- Rust QUIC guide (standalone): `example/relay_example/quic/README_RUST.md`
- Mock OAuth server (dev only): `example/relay_example/mock_oauth_server/README.md`

## Example groups

- Basic relay: `example/relay_example/basic_relay_a`, `example/relay_example/basic_relay_b`
- QUIC relay: `example/relay_example/quic_basic_receiver`, `example/relay_example/quic_basic_sender`
- WebSocket relay: `example/relay_example/websocket_basic_receiver`, `example/relay_example/websocket_basic_sender`
- WebTransport relay (build tag): `example/relay_example/webtransport_basic_receiver`, `example/relay_example/webtransport_basic_sender`
- NATS fanout (build tag): `example/relay_example/nats_basic_subscriber`, `example/relay_example/nats_basic_publisher`
- QUIC secure (OAuth + AES):
  - Sender: `example/relay_example/quic_secure_oauth_aes_sender`
  - Receiver: `example/relay_example/quic_secure_oauth_aes_receiver`
- Secure relay (JWT/JWKS + mTLS + AES):
  - Sender: `example/relay_example/secure_advanced_relay_a_oauth_mtls_aes`
  - Receiver: `example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes`
- Secure relay (gRPC-Web frontend): `example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb`
- Secure relay (introspection): `example/relay_example/secure_advanced_relay_b_oauth_rfc7662_introspection_mtls_aes`
- Secure relay (Kafka + S3):
  - Kafka: `example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_kafka`
  - S3 Parquet: `example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_s3_parquet`

## TLS assets

Shared TLS assets live in `example/relay_example/tls/`.
