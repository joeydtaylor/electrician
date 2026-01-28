# QUIC Relay Examples

These examples mirror the gRPC relay patterns, but use QUIC for transport. QUIC is UDP-based,
reliable, multiplexed, and uses TLS 1.3 by default.

## Files

- `quic_basic_receiver/` - minimal QUIC receiver
- `quic_basic_sender/` - minimal QUIC sender
- `quic_secure_oauth_aes_receiver/` - secure receiver (OAuth2 JWKS + AES-GCM + static headers)
- `quic_secure_oauth_aes_sender/` - secure sender (OAuth2 client_credentials + AES-GCM + static headers)
- `README_RUST.md` - Rust integration guide (QUIC framing, auth, AES-GCM)

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
go run ./example/relay_example/quic_basic_sender
```

## Secure flow (OAuth2 + AES-GCM)

Before running, edit the constants near the top of these files to match your auth server,
TLS paths, and AES key:

- `example/relay_example/quic_secure_oauth_aes_receiver/main.go`
- `example/relay_example/quic_secure_oauth_aes_sender/main.go`

Start the secure receiver:

```bash
go run ./example/relay_example/quic_secure_oauth_aes_receiver
```

Send secure messages:

```bash
go run ./example/relay_example/quic_secure_oauth_aes_sender
```

Notes:

- QUIC requires TLS 1.3.
- The receiver enforces `x-tenant: local` via StreamOpen defaults.
- AES-GCM uses the same 32-byte key as the gRPC secure example.
- The sender fetches a JWT using client credentials; the receiver validates via JWKS.

## Logging

Set the `logLevel` constant in the sender/receiver `main.go` files to `debug` to see
verbose, structured logs (including auth headers and tokens).

## Mock OAuth server (dev only)

If you do not want to rely on a real auth server during development, use the mock OAuth
server at:

- `example/relay_example/mock_oauth_server/README.md`
