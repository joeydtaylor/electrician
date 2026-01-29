# Relay Transports: WebSocket, WebTransport, NATS (last three)

This doc covers the **last three** transports on the matrix:

- **WebSocket / WSS** (browser-compatible fallback)
- **WebTransport** (browser QUIC semantics)
- **NATS / PubSub** (server-side fanout)

If you are a frontend engineer, only **WebSocket** and **WebTransport** apply to you.
NATS is **server-side only**.

---

## Shared contract (all three)

- Proto: `proto/electrician_relay.proto`
- Message type: `RelayEnvelope` (stream open + payloads + close)
- Payload type: `WrappedPayload`
- Encryption: AES-GCM with **12-byte IV prefix**
- JSON payloads are recommended for browser clients

**Headers & auth:**
- Always send `authorization: Bearer <token>` and `x-tenant: local`
- Browsers cannot set custom headers on a WebSocket handshake. Send these headers inside
  `StreamOpen.defaults.headers` instead.

**AES-GCM payload format:**
```
IV (12 bytes) || ciphertext+tag
```

---

## 1) WebSocket / WSS (browser fallback)

### Why
- Works everywhere
- Reliable/ordered (TCP)
- Good for UI feeds, notifications, admin consoles

### Server endpoint
- Example: `wss://localhost:8084/relay`
- The server expects a **StreamOpen** message first

### Message framing
Each WebSocket **binary** message is a protobuf payload.

- `RelayEnvelope` messages for open/payload/close
- `StreamAcknowledgment` messages from server

### Minimal browser flow (TypeScript)

```ts
import { RelayEnvelope, StreamOpen, WrappedPayload, SecurityOptions, EncryptionSuite } from "./gen/relay_pb";

const ws = new WebSocket("wss://localhost:8084/relay");
ws.binaryType = "arraybuffer";

ws.onopen = () => {
  const open = new StreamOpen({
    streamId: crypto.randomUUID(),
    defaults: {
      headers: {
        authorization: `Bearer ${token}`,
        "x-tenant": "local",
      },
      contentType: "application/json",
      security: new SecurityOptions({
        enabled: true,
        suite: EncryptionSuite.ENCRYPTION_AES_GCM,
      }),
    },
    ackMode: 2, // ACK_BATCH
    ackEveryN: 1024,
    maxInFlight: 8192,
    omitPayloadMetadata: true,
  });

  const env = new RelayEnvelope({ open });
  ws.send(env.toBinary());
};

function sendPayload(obj: any) {
  const plaintext = new TextEncoder().encode(JSON.stringify(obj));
  const encrypted = aesGcmEncrypt(plaintext, keyBytes); // IV || ciphertext

  const payload = new WrappedPayload({
    id: crypto.randomUUID(),
    payload: encrypted,
    payloadEncoding: 0,
    payloadType: "electrician.Feedback",
    metadata: {
      contentType: "application/json",
      security: { enabled: true, suite: EncryptionSuite.ENCRYPTION_AES_GCM },
    },
  });

  const env = new RelayEnvelope({ payload });
  ws.send(env.toBinary());
}
```

---

## 2) WebTransport (browser QUIC semantics)

### Why
- QUIC streams + datagrams (when supported)
- Lower latency than WebSocket
- Modern replacement path for WSS (not universal yet)

### Server endpoint
- Example: `https://localhost:8443/relay`
- Uses WebTransport over HTTP/3

### Streams (reliable)
Each stream frame is a protobuf `RelayEnvelope` with **length-prefix framing**:

```
[u32 length BE] + [protobuf bytes]
```

### Datagrams (unreliable)
If enabled, send a protobuf-encoded `WrappedPayload` as the datagram body.
No acknowledgments.

### Minimal browser flow (TypeScript)

```ts
const transport = new WebTransport("https://localhost:8443/relay");
await transport.ready;

const stream = await transport.createBidirectionalStream();
const writer = stream.writable.getWriter();
const reader = stream.readable.getReader();

const open = new StreamOpen({
  streamId: crypto.randomUUID(),
  defaults: {
    headers: { authorization: `Bearer ${token}`, "x-tenant": "local" },
    contentType: "application/json",
    security: { enabled: true, suite: EncryptionSuite.ENCRYPTION_AES_GCM },
  },
  ackMode: 2,
  ackEveryN: 1024,
  maxInFlight: 8192,
  omitPayloadMetadata: true,
});

const openEnv = new RelayEnvelope({ open });
await writer.write(frame(openEnv.toBinary())); // length prefix

// Send payloads over stream
const payloadEnv = new RelayEnvelope({ payload });
await writer.write(frame(payloadEnv.toBinary()));
```

---

## 3) NATS / PubSub (server-side only)

NATS is a **server-side fanout fabric**. It is **not** a browser transport.
Use it to broadcast inside your cluster (shards/rooms/cells), not to clients.

### Expected usage
- `ForwardRelay` publishes `WrappedPayload` (protobuf) to a subject
- `ReceivingRelay` subscribes and unwraps payloads

---

## Build tags (repo)

These transports are implemented behind build tags so the repo builds without
optional deps by default:

- WebTransport: `-tags webtransport`
- NATS: `-tags nats`

Example:

```bash
go run -tags webtransport ./example/relay_example/webtransport_basic_receiver
```

---

## AES-GCM helper (browser)

```ts
async function aesGcmEncrypt(plaintext: Uint8Array, keyBytes: Uint8Array) {
  const key = await crypto.subtle.importKey(
    "raw",
    keyBytes,
    { name: "AES-GCM" },
    false,
    ["encrypt"],
  );
  const iv = crypto.getRandomValues(new Uint8Array(12));
  const ciphertext = new Uint8Array(
    await crypto.subtle.encrypt({ name: "AES-GCM", iv }, key, plaintext),
  );
  const out = new Uint8Array(iv.length + ciphertext.length);
  out.set(iv, 0);
  out.set(ciphertext, iv.length);
  return out; // IV || ciphertext+tag
}
```
