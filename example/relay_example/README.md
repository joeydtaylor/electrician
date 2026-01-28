# Relay Examples (Index)

This folder groups all relay examples (gRPC, QUIC, secure variants) and the docs that go with them.

## Key docs

- Frontend integration (gRPC-Web + QUIC): `example/relay_example/FRONTEND_INTEGRATION_UNIFIED.md`
- QUIC relay docs (including Rust guide): `example/relay_example/quic/README.md`
- Rust QUIC guide (standalone): `example/relay_example/quic/README_RUST.md`
- Mock OAuth server (dev only): `example/relay_example/mock_oauth_server/README.md`

## Example groups

- Basic relay: `example/relay_example/basic_relay_a`, `example/relay_example/basic_relay_b`
- QUIC relay: `example/relay_example/quic_basic_receiver`, `example/relay_example/quic_basic_sender`
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
