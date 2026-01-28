# Rust QUIC Relay Integration (Secure OAuth2 + AES-GCM)

This guide shows how to build a Rust QUIC sender/receiver that interoperates with the Go QUIC
examples in this repo:

- `example/relay_example/quic_secure_oauth_aes_sender`
- `example/relay_example/quic_secure_oauth_aes_receiver`

It is **not gRPC**. It is a raw QUIC stream with length-prefixed protobuf frames.

---

## 0) Quick Compatibility Checklist

Match these exactly or you will get auth/crypto failures:

- **ALPN**: `electrician-quic-relay-v1`
- **Proto**: `proto/electrician_relay.proto`
- **Stream framing**: 4-byte big-endian length prefix + protobuf bytes
- **Auth headers** (in `StreamOpen.defaults.headers`):
  - `authorization: Bearer <JWT>`
  - `x-tenant: local`
- **Issuer** must equal the JWT `iss` claim (in Go examples, this is `auth-service`).
- **AES-GCM**: payload = `IV(12 bytes) || ciphertext+tag` (same key as Go example)

---

## 1) Required Constants (match Go example)

Set these to the same values as the Go QUIC examples:

```text
relay_addr = "localhost:50072"
quic_alpn = "electrician-quic-relay-v1"
issuer = "auth-service"               // must match JWT "iss" claim
jwks_url = "https://localhost:3000/api/auth/oauth/jwks.json"
audience = "your-api"
required_scope = "write:data"
static_header_x_tenant = "local"

// AES-256 key (hex) -- must be exactly 32 bytes
key_hex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
```

---

## 2) Protobuf Codegen (Rust)

This is a simple `prost`-based setup (no gRPC):

**Cargo.toml** (dependencies you likely need):

```toml
[dependencies]
bytes = "1"
prost = "0.12"
prost-types = "0.12"
quinn = "0.11"
rcgen = "0.13"
rustls = "0.23"
serde = { version = "1", features = ["derive"] }
serde_json = "1"
hex = "0.4"
base64 = "0.22"
rand = "0.8"
aes-gcm = "0.10"
reqwest = { version = "0.12", default-features = false, features = ["json", "rustls-tls"] }
tokio = { version = "1", features = ["rt-multi-thread", "macros"] }

[build-dependencies]
prost-build = "0.12"
```

**build.rs**:

```rust
fn main() {
    prost_build::Config::new()
        .out_dir("src/gen")
        .compile_protos(
            &["proto/electrician_relay.proto"],
            &["proto"],
        )
        .unwrap();
}
```

---

## 3) QUIC Frame Format (MUST MATCH)

Every message on the stream is:

```
[u32 length big-endian] + [protobuf bytes]
```

The protobuf message is **RelayEnvelope** (from `electrician_relay.proto`).

### Rust helpers

```rust
fn write_frame(buf: &mut Vec<u8>, payload: &[u8]) {
    let len = payload.len() as u32;
    buf.extend_from_slice(&len.to_be_bytes());
    buf.extend_from_slice(payload);
}

fn read_frame(mut bytes: &[u8]) -> Option<Vec<u8>> {
    if bytes.len() < 4 { return None; }
    let len = u32::from_be_bytes(bytes[0..4].try_into().ok()?) as usize;
    if bytes.len() < 4 + len { return None; }
    Some(bytes[4..4+len].to_vec())
}
```

---

## 4) Sender Flow (Rust -> Go Receiver)

1. Dial QUIC with ALPN `electrician-quic-relay-v1`.
2. Open a bidirectional stream.
3. Send `StreamOpen` in a `RelayEnvelope`.
4. Send each `WrappedPayload` in its own `RelayEnvelope`.
5. Read `StreamAcknowledgment` frames in the background.

### StreamOpen defaults (important)

Set these in `StreamOpen.defaults`:

- `headers["authorization"] = "Bearer <token>"`
- `headers["x-tenant"] = "local"`
- `content_type = "application/json"`
- `version = { major: 1, minor: 0 }`
- `security = { enabled: true, suite: ENCRYPTION_AES_GCM }`

Also set:

- `ack_mode = ACK_BATCH`
- `ack_every_n = 1024`
- `max_in_flight = 8192`
- `omit_payload_metadata = true`

### Payload format

For JSON payloads:

- `payload` = UTF-8 JSON bytes (optionally AES-GCM encrypted)
- `payload_encoding = PAYLOAD_ENCODING_UNSPECIFIED`
- `payload_type` optional (e.g. `"electrician.Feedback"`)

### AES-GCM (same as Go)

```rust
use aes_gcm::{Aes256Gcm, Nonce};
use aes_gcm::aead::{Aead, KeyInit};
use rand::RngCore;

fn encrypt_aes_gcm(key: &[u8; 32], plaintext: &[u8]) -> Vec<u8> {
    let cipher = Aes256Gcm::new_from_slice(key).unwrap();
    let mut iv = [0u8; 12];
    rand::thread_rng().fill_bytes(&mut iv);
    let nonce = Nonce::from_slice(&iv);
    let ciphertext = cipher.encrypt(nonce, plaintext).unwrap();
    // wire format: IV || ciphertext+tag
    [iv.to_vec(), ciphertext].concat()
}
```

---

## 5) Receiver Flow (Rust <- Go Sender)

1. Run a QUIC server with ALPN `electrician-quic-relay-v1`.
2. Accept incoming connections + streams.
3. Read `RelayEnvelope` frames.
4. First message should be `StreamOpen`.
5. Validate headers (auth + x-tenant).
6. For each `WrappedPayload`:
   - decrypt if `metadata.security.enabled == true`
   - decode JSON or protobuf
   - send `StreamAcknowledgment` based on `ack_mode`

### ACKs (what Go sender expects)

If `ack_mode == ACK_BATCH`, send one ack every `ack_every_n` messages:

```text
StreamAcknowledgment {
  success: true
  stream_id: <same stream_id>
  last_seq: <last seq>
  ok_count: <batch count>
  err_count: 0
}
```

For simple interop, `ACK_PER_MESSAGE` is fine too.

---

## 6) OAuth2 Token (Client Credentials)

The Go sender uses client credentials at:

```
POST {auth_base_url}/api/auth/oauth/token
Content-Type: application/x-www-form-urlencoded
Authorization: Basic base64(client_id:client_secret)

grant_type=client_credentials&scope=write:data
```

Use any Rust HTTP client to fetch the token and place it in
`StreamOpen.defaults.headers["authorization"]`.

**Important**: the receiver validates `iss`, `aud`, and `scope`.
Make sure the `iss` claim equals `issuer` above.

---

## 7) Smoke Test Interop

**Go receiver -> Rust sender**

1) Run Go receiver:

```
go run ./example/relay_example/quic_secure_oauth_aes_receiver
```

2) Run Rust sender. You should see payloads in the Go receiver output.

**Rust receiver -> Go sender**

1) Run Rust receiver.
2) Run Go sender:

```
go run ./example/relay_example/quic_secure_oauth_aes_sender
```

---

## 8) Gotchas

- If you see `issuer mismatch`, your Rust receiver's issuer string does not match the token `iss`.
- If you see `missing bearer token`, you forgot to put `authorization` in `StreamOpen.defaults.headers`.
- If payloads don't decrypt, double-check AES key length and IV prefix.
- The QUIC stream uses **length-prefixed protobuf**, not gRPC.

---

## 9) Wire Objects to Use

From `electrician_relay.proto`:

- `RelayEnvelope`
- `StreamOpen`
- `WrappedPayload`
- `StreamAcknowledgment`

That's it. Keep the framing + headers + AES the same and Rust will interop cleanly with Go.
