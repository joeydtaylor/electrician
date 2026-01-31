# Mock OAuth Server (JWKS + Introspection + Session Token)

This is a tiny OAuth mock used for local relay testing. It serves:

- `GET /api/auth/oauth/jwks.json` (JWKS)
- `POST /api/auth/oauth/token` (client credentials)
- `POST /api/auth/session/token` (issues JWT access tokens)
- `POST /api/auth/oauth/introspect` (RFC 7662‑style introspection)

## Quick start

```bash
go run ./example/auth/mock_oauth_server
```

Defaults:

- Address: `https://localhost:3000`
- Issuer: `auth-service`
- Audience: `your-api`
- Scope: `write:data`
- Token TTL: `300s`
- Client ID: `steeze-local-cli`
- Client Secret: `local-secret`
- JWKS URL: `https://localhost:3000/api/auth/oauth/jwks.json`

Secure relay defaults (matches the secure gRPC + QUIC examples):
- AES-256 key (hex): `ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec`
- Required header: `x-tenant: local`
- Compression: Snappy when `metadata.performance.use_compression = true`

## Use with the secure gRPC‑web relay example

Start the mock OAuth server:

```bash
go run ./example/auth/mock_oauth_server
```

Start the receiver (JWKS):

```bash
OAUTH_ISSUER_BASE=auth-service \
OAUTH_JWKS_URL=https://localhost:3000/api/auth/oauth/jwks.json \
go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

Fetch a token for the frontend:

```bash
curl -k -X POST https://localhost:3000/api/auth/session/token \
  -H 'Content-Type: application/json' \
  -d '{"scope":"write:data"}'
```

Use `access_token` as `Authorization: Bearer <token>` and send `x-tenant: local` as gRPC metadata.

Client credentials token (non-browser clients):

```bash
curl -k -u steeze-local-cli:local-secret \
  -d grant_type=client_credentials \
  -d scope=write:data \
  https://localhost:3000/api/auth/oauth/token
```

## Use with QUIC introspection

The mock server supports introspection with Basic auth by default:

```bash
OAUTH_INTROSPECTION_URL=https://localhost:3000/api/auth/oauth/introspect \
OAUTH_INTROSPECTION_AUTH_TYPE=basic \
OAUTH_INTROSPECTION_CLIENT_ID=steeze-local-cli \
OAUTH_INTROSPECTION_CLIENT_SECRET=local-secret \
go run ./example/relay_example/quic_secure_oauth_aes_receiver
```

## Environment variables

- `OAUTH_ADDR` (default `localhost:3000`)
- `OAUTH_TLS_DISABLE` (default `false`)
- `TLS_CERT`, `TLS_KEY` (defaults: auto-detected from `example/relay_example/tls/server.crt` and `example/relay_example/tls/server.key`)
- `OAUTH_CLIENT_ID` (default `steeze-local-cli`)
- `OAUTH_CLIENT_SECRET` (default `local-secret`)
- `OAUTH_ISSUER_BASE` (default `auth-service`)
- `OAUTH_AUDIENCE` (default `your-api`)
- `OAUTH_SCOPE` (default `write:data`)
- `OAUTH_TOKEN_TTL` (default `300s`)
- `OAUTH_KID` (default `mock-key-1`)
- `OAUTH_SUBJECT` (default `user-local`)
- `OAUTH_STATIC_TOKEN` (default `token-123`)

Introspection auth:

- `INTROSPECT_AUTH` (`basic`, `bearer`, or `none`; default `basic`)
- `INTROSPECT_CLIENT_ID` (default `steeze-local-cli`)
- `INTROSPECT_CLIENT_SECRET` (default `local-secret`)
- `INTROSPECT_BEARER_TOKEN` (default empty)

Logging:

- `LOG_LEVEL` (default `info`; set `debug` to log tokens)
