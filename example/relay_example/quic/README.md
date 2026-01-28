# QUIC Relay Examples

These examples mirror the gRPC relay patterns, but use QUIC for transport. QUIC is UDP-based,
reliable, multiplexed, and uses TLS 1.3 by default.

## Files

- `quic_basic_receiver/` – minimal QUIC receiver
- `quic_basic_sender/` – minimal QUIC sender
- `quic_secure_oauth_aes_receiver/` – secure receiver (OAuth2 introspection + AES-GCM + static headers)
- `quic_secure_oauth_aes_sender/` – secure sender (AES-GCM + bearer token + static headers)

## Shared TLS assets

These examples reuse the TLS assets under `example/relay_example/tls/`.
Browsers are not involved here, but clients must trust the CA.

## Basic flow

Start the receiver:

```bash
go run ./example/relay_example/quic_basic_receiver
```

Send messages:

```bash
RELAY_ADDR=localhost:50071 go run ./example/relay_example/quic_basic_sender
```

## Secure flow (OAuth2 + AES-GCM)

Start the secure receiver:

```bash
OAUTH_INTROSPECTION_URL=https://localhost:3000/api/auth/oauth/introspect \
RELAY_ADDR=localhost:50072 \
go run ./example/relay_example/quic_secure_oauth_aes_receiver
```

Send secure messages (requires a bearer token and matching `x-tenant`):

```bash
RELAY_ADDR=localhost:50072 \
RELAY_BEARER_TOKEN=your-token \
go run ./example/relay_example/quic_secure_oauth_aes_sender
```

Notes:

- QUIC requires TLS 1.3.
- The receiver enforces `x-tenant: local` via StreamOpen defaults.
- AES-GCM uses the same 32-byte key as the gRPC secure example.
