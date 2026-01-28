# Frontend Integration: Secure Relay (gRPC-web + OAuth + AES-GCM)

This guide is for the frontend team. It targets the most secure relay example in this repo:
`example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb`.

If you want a single doc that covers gRPC-web + QUIC side-by-side, use:
`example/relay_example/FRONTEND_INTEGRATION_UNIFIED.md`.

It assumes the browser uses Connect (gRPC-web compatible) to call
`electrician.RelayService/Receive` from `proto/electrician_relay.proto`.

## 90-second checklist (copy/paste)

1) Start mock OAuth (dev):
```bash
go run ./example/relay_example/mock_oauth_server
```

2) Start receiver:
```bash
OAUTH_ISSUER_BASE=auth-service \
  go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

3) In the browser client:
- Base URL: `https://localhost:50051`
- Headers (gRPC metadata):
  - `authorization: Bearer <token>`
  - `x-tenant: local`
- Payload:
  - JSON bytes
  - `content_type = application/json`
  - `payload_encoding = UNSPECIFIED`
- If encrypting:
  - AES-GCM encrypt payload (IV + ciphertext+tag)
  - `metadata.security.enabled = true`
  - `metadata.security.suite = ENCRYPTION_AES_GCM`

If you hit `issuer mismatch`, the token `iss` does not match the receiver's issuer.
Check the receiver log line **Auth JWT validator installed** to see the expected issuer.

## Backend: run the secure receiver

From repo root:

```bash
go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

Defaults:

- Receiver address: `https://localhost:50051`
- TLS certs: `example/relay_example/tls/*` (browser must trust the CA)
- Auth issuer: `https://localhost:3000`
- JWKS URL: `https://localhost:3000/api/auth/oauth/jwks.json`
- Required audience: `your-api`
- Required scope: `write:data`
- Required metadata header: `x-tenant: local`
- AES-GCM key (hex):
  `ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec`

If you need different values, set these env vars before running:

- `RX_ADDR` (receiver address)
- `TLS_CERT`, `TLS_KEY`, `TLS_CA`
- `OAUTH_ISSUER_BASE`, `OAUTH_JWKS_URL`

## Client requirements (must-haves)

1. Use `proto/electrician_relay.proto` (not `proto/relay.proto`).
2. Call `electrician.RelayService/Receive` over gRPC-web/Connect.
3. Send gRPC metadata headers:
   - `authorization: Bearer <token>`
   - `x-tenant: local`
4. If encryption is enabled, AES-GCM encrypt the payload and set metadata:
   - `metadata.security.enabled = true`
   - `metadata.security.suite = ENCRYPTION_AES_GCM`

Note: `x-tenant` must be gRPC metadata. Putting it in `WrappedPayload.metadata.headers`
will not satisfy the server's static header check.

## Token requirements

The backend validates JWTs with JWKS and requires:

- `iss` matches `OAUTH_ISSUER_BASE`
- `aud` includes `your-api`
- `scope` includes `write:data`

The UI can fetch a token from:

```text
POST {authBaseUrl}/api/auth/session/token
{ "scope": "write:data" }
```

Use the returned `access_token` as `Authorization: Bearer <token>`.

Tip: the receiver logs the issuer it expects on startup:
`Auth JWT validator installed`

## WrappedPayload expectations

For JSON (recommended in browser):

- `payload`: UTF-8 JSON bytes
- `metadata.content_type`: `application/json`
- `payload_encoding`: leave `UNSPECIFIED`
- `payload_type`: optional (unused for JSON decode)

If the server expects the Feedback shape, JSON should look like:

```json
{
  "customerId": "cust-123",
  "content": "Relay test payload",
  "category": "feedback",
  "isNegative": false,
  "tags": ["ui", "grpc-web"]
}
```

## AES-GCM encryption details

The receiver expects the payload to be encrypted as:

```
IV (12 bytes) || ciphertext+tag
```

- Key: 32 bytes (AES-256), derived from the hex string above.
- Algorithm: AES-GCM
- Additional authenticated data: none

The browser should:

1. Decode the hex key into raw bytes.
2. Generate a random 12-byte IV.
3. Encrypt the plaintext payload bytes.
4. Prefix IV to the ciphertext+tag and send as `WrappedPayload.payload`.

## Minimal Connect (TypeScript) sketch

```ts
import { createPromiseClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { RelayService } from "./gen/relay_connect";
import { WrappedPayload, SecurityOptions, EncryptionSuite } from "./gen/relay_pb";

const transport = createGrpcWebTransport({
  baseUrl: "https://localhost:50051",
  credentials: "include",
});

const client = createPromiseClient(RelayService, transport);

const headers = {
  "authorization": `Bearer ${token}`,
  "x-tenant": "local",
};

const payloadBytes = new TextEncoder().encode(JSON.stringify({
  customerId: "cust-123",
  content: "Relay test payload",
  category: "feedback",
  isNegative: false,
  tags: ["ui", "grpc-web"],
}));

const encryptedBytes = aesGcmEncrypt(payloadBytes, hexKeyBytes); // IV + ciphertext

const request = new WrappedPayload({
  id: crypto.randomUUID(),
  timestamp: new Date(),
  payload: encryptedBytes,
  payloadEncoding: 0, // PAYLOAD_ENCODING_UNSPECIFIED
  metadata: {
    contentType: "application/json",
    security: new SecurityOptions({
      enabled: true,
      suite: EncryptionSuite.ENCRYPTION_AES_GCM,
    }),
  },
});

const ack = await client.receive(request, { headers });
console.log(ack);
```

## Common pitfalls

- gRPC-web vs gRPC: Connect in the browser requires a gRPC-web compatible endpoint.
- TLS: browsers must trust `example/relay_example/tls/ca.crt`.
- Headers: `x-tenant` must be gRPC metadata, not payload metadata.
- Auth: missing `aud` or `scope` claims will fail auth.
- Ack semantics: `Receive` returns before payload decode/processing completes.

## Need changes from frontend team

If anything in the current UI does not support the requirements above, please flag it.
We control the backend and can adjust, but the current secure example expects the
exact headers, auth, and AES-GCM format described here.
