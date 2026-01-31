# Frontend Integration (gRPC-Web + QUIC)

This guide explains how the frontend should integrate with Electrician relay:

- **gRPC-Web (Connect)**: supported in browsers today.
- **QUIC relay**: raw QUIC stream framing (not directly available in browsers without WebTransport).

If you only ship a browser UI, implement **gRPC-Web** now. QUIC is for native/desktop/mobile
clients or a proxy service that speaks QUIC on behalf of the browser.

Source of truth for gRPC-web alignment:
- `docs/relay-grpcweb-alignment.md`

---

## 60-second sanity path (browser)

1) Start mock OAuth (dev):
```bash
go run ./example/auth/mock_oauth_server
```

2) Start secure gRPC-web receiver:
```bash
OAUTH_ISSUER_BASE=auth-service \
  go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

3) Use Connect and send:
- `authorization: Bearer <token>`
- `x-tenant: local`
- JSON payload, `content_type = application/json`
- If encrypted: AES-GCM + `metadata.security.enabled = true`

If you see `issuer mismatch`, the token `iss` does not match the receiver's issuer.
Check the receiver log line **Auth JWT validator installed** for the exact issuer.

---

## 0) Shared Rules (Applies to both gRPC-Web and QUIC)

**Proto**
- Use `proto/electrician_relay.proto`.
- Service: `electrician.RelayService/Receive`.

**Auth**
- Always send `authorization: Bearer <token>` as transport metadata.
- Tokens must include:
  - `iss` = configured issuer (check receiver log on startup)
  - `aud` includes `your-api`
  - `scope` includes `write:data`

**Static header**
- Always send `x-tenant: local` as **transport metadata**.
  - For gRPC-Web, this is request headers.
  - For QUIC, this lives in `StreamOpen.defaults.headers`.

**Security (AES-GCM)**
- If the server is configured with a decryption key (secure examples), encryption is required.
  Encrypt the payload bytes and set:
  - `metadata.security.enabled = true`
  - `metadata.security.suite = ENCRYPTION_AES_GCM`
- Payload format: `IV (12 bytes) || ciphertext+tag`.

**Payload JSON shape** (if you are sending Feedback):

```json
{
  "customerId": "cust-123",
  "content": "Relay test payload",
  "category": "feedback",
  "isNegative": false,
  "tags": ["ui", "grpc-web"]
}
```

---

## 1) gRPC-Web (Connect) - Browser Ready

### Required headers
- `authorization: Bearer <token>`
- `x-tenant: local`

### Browser payload rule (JSON only)
- Do not send GOB from the browser.
- Always set `content_type = application/json`.

### Minimal TypeScript (Connect)

```ts
import { createPromiseClient } from "@connectrpc/connect";
import { createGrpcWebTransport } from "@connectrpc/connect-web";
import { RelayService } from "./gen/relay_connect";
import { WrappedPayload, SecurityOptions, EncryptionSuite } from "./gen/relay_pb";

const baseUrl = "https://localhost:50051";
const token = "<your bearer token>";
const headers = {
  "authorization": `Bearer ${token}`,
  "x-tenant": "local",
};

const transport = createGrpcWebTransport({
  baseUrl,
  credentials: "include",
});

const client = createPromiseClient(RelayService, transport);

const payloadBytes = new TextEncoder().encode(JSON.stringify({
  customerId: "cust-123",
  content: "Relay test payload",
  category: "feedback",
  isNegative: false,
  tags: ["ui", "grpc-web"],
}));

const encrypted = aesGcmEncrypt(payloadBytes, keyBytes); // IV + ciphertext

const req = new WrappedPayload({
  id: crypto.randomUUID(),
  timestamp: new Date(),
  payload: encrypted,
  payloadEncoding: 0, // PAYLOAD_ENCODING_UNSPECIFIED
  metadata: {
    contentType: "application/json",
    security: new SecurityOptions({
      enabled: true,
      suite: EncryptionSuite.ENCRYPTION_AES_GCM,
    }),
  },
});

const ack = await client.receive(req, { headers });
console.log(ack);
```

### Token acquisition (frontend)

```text
POST {authBaseUrl}/api/auth/session/token
{ "scope": "write:data" }
```

Use the `access_token` as the Bearer token.

### Required CORS / server support
- Backend must expose a **Connect or gRPC-Web compatible** HTTP endpoint.
- If cookies are used, allow credentialed CORS.

---

## 2) QUIC Relay (Native/Desktop/Mobile)

### Important browser limitation
Browsers do **not** expose raw QUIC streams. If you want QUIC from the browser, you must use
**WebTransport over HTTP/3** and a server that supports it (not implemented in this repo).

**Options:**
1) Use gRPC-Web in the browser.
2) Use a backend proxy (browser -> HTTP -> proxy -> QUIC).
3) Build a native client that speaks QUIC directly.

### QUIC wire format (for native clients)

- ALPN: `electrician-quic-relay-v1`
- Stream framing: 4-byte big-endian length prefix + protobuf bytes
- Protobuf message: `RelayEnvelope`

**Frame format:**

```
[u32 length BE] + [protobuf bytes]
```

### QUIC stream open (first message)
Send a `RelayEnvelope` with `StreamOpen`:

- `defaults.headers["authorization"] = "Bearer <token>"`
- `defaults.headers["x-tenant"] = "local"`
- `defaults.content_type = "application/json"`
- `defaults.security.enabled = true`
- `defaults.security.suite = ENCRYPTION_AES_GCM`
- `ack_mode = ACK_BATCH`
- `ack_every_n = 1024`
- `max_in_flight = 8192`
- `omit_payload_metadata = true`

### QUIC sender flow
1) Dial QUIC with ALPN `electrician-quic-relay-v1`.
2) Open a bidirectional stream.
3) Send `StreamOpen` envelope.
4) For each log/event, send `WrappedPayload` envelope.
5) Read `StreamAcknowledgment` frames (avoid backpressure).

### QUIC receiver flow
1) Accept QUIC connections.
2) Read `RelayEnvelope` frames.
3) Validate headers on `StreamOpen`.
4) Decrypt (AES-GCM) if `security.enabled == true`.
5) Unmarshal JSON into your domain type.

---

## 3) AES-GCM helper (browser)

```ts
async function aesGcmEncrypt(plaintext: Uint8Array, keyBytes: Uint8Array) {
  const key = await crypto.subtle.importKey(
    "raw",
    keyBytes,
    { name: "AES-GCM" },
    false,
    ["encrypt"]
  );
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const ciphertext = new Uint8Array(
    await crypto.subtle.encrypt({ name: "AES-GCM", iv }, key, plaintext)
  );
  const out = new Uint8Array(iv.length + ciphertext.length);
  out.set(iv, 0);
  out.set(ciphertext, iv.length);
  return out; // IV || ciphertext+tag
}
```

---

## 4) Common Pitfalls

- `issuer mismatch`: your receiver expects `issuer` that does not match JWT `iss`.
- `missing bearer token`: you didn't send `authorization` in transport metadata.
- `x-tenant` must be transport metadata, not payload metadata.
- AES-GCM requires **12-byte IV prefix** and correct key length (32 bytes).
- QUIC requires ALPN `electrician-quic-relay-v1` and length-prefixed frames.

---

## 5) What to Implement First

- Browser UI: gRPC-Web (Connect) only.
- Native clients: QUIC relay for lower latency.
- If you need browser QUIC: plan for WebTransport + a new server implementation.
