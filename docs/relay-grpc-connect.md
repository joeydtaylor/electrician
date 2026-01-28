# Relay gRPC-Web / Connect Integration

This guide explains how a browser client (Connect / gRPC-web) calls
`electrician.RelayService/Receive` using the Electrician relay contract.

## Endpoint
- Base URL: `https://localhost:50051`
- Method: `electrician.RelayService/Receive`
- Proto: `proto/electrician_relay.proto`

## Required transport headers (gRPC metadata)
- `authorization: Bearer <token>`
- `x-tenant: local`

> These must be sent as transport metadata, not only in payload metadata.

## Token requirements (JWT)
The secure gRPC-web receiver validates JWTs with JWKS. Tokens must contain:
- `iss`: issuer that matches the receiver config
  - Default example: `https://localhost:3000`
- `aud`: includes `your-api`
- `scope`: includes `write:data`

### Token endpoint (example)
```text
POST https://localhost:3000/api/auth/session/token
{ "scope": "write:data" }
```
Use the returned `access_token` as the Bearer token.

## Payload rules
- `payload_encoding`: `PAYLOAD_ENCODING_UNSPECIFIED` (0)
- `content_type`: `application/json`
- `payload_type`: optional (e.g. `electrician.Feedback`)

## AES-GCM (if enabled)
Encrypt the payload bytes and set:
- `metadata.security.enabled = true`
- `metadata.security.suite = ENCRYPTION_AES_GCM`

Payload format: `IV (12 bytes) || ciphertext+tag`

## Reference implementation
Use the unified frontend doc for a full TypeScript sample:
`example/relay_example/FRONTEND_INTEGRATION_UNIFIED.md`
