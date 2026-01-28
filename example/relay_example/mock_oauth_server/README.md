# Mock OAuth Server (Dev Only)

This is a tiny local OAuth server for dev/testing so you are not blocked if the real auth
service is down. It issues JWTs, serves JWKS, and supports introspection.

## Run

```bash
go run ./example/relay_example/mock_oauth_server
```

It starts on:

```
https://localhost:3000
```

The server uses a self-signed TLS cert generated at startup. Your clients should allow
insecure TLS for local dev (the Go examples already do this).

## Defaults (match QUIC examples)

- issuer: `auth-service`
- audience: `your-api`
- client_id: `steeze-local-cli`
- client_secret: `local-secret`
- scope: `write:data` (default)

## Endpoints

- Token (client credentials):
  - `POST /api/auth/oauth/token`
  - Basic auth with `client_id:client_secret`
  - Form body: `grant_type=client_credentials&scope=write:data`
- JWKS:
  - `GET /api/auth/oauth/jwks.json`
- Introspection:
  - `POST /api/auth/oauth/introspect`
  - Form body: `token=<jwt>`
- Session token (frontend-style):
  - `POST /api/auth/session/token`
  - JSON body: `{ "scope": "write:data" }`

## Example curl

```bash
curl -k -u steeze-local-cli:local-secret \
  -d grant_type=client_credentials \
  -d scope=write:data \
  https://localhost:3000/api/auth/oauth/token
```

## Notes

- Tokens are signed with a fresh RSA key on each server start.
- JWKS is generated from the in-memory key.
- Issuer must match the value your receivers validate against.
