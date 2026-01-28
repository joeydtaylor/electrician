# quicrelay

Package quicrelay provides QUIC-based Forward and Receiving relays with the same payload
contract as the gRPC relays (`relay.WrappedPayload`, `relay.RelayEnvelope`, `relay.StreamAcknowledgment`).

Highlights:

- QUIC transport (UDP + TLS 1.3)
- Stream-based relay with ack modes (per-message, batch, none)
- Optional payload compression and AES-GCM encryption (reuse existing metadata)
- Static headers + optional OAuth2 introspection (bearer validation)
- Works with existing relay proto definitions

See examples under `example/relay_example/quic/*`.
