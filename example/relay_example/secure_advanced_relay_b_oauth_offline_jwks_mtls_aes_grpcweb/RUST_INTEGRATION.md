# Rust Integration: Secure Relay (gRPC + JWKS + AES‑GCM)

This guide shows how a Rust service can send encrypted `WrappedPayload` messages
to the secure relay example:

`example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb`

> Note: gRPC‑web is for browsers. Rust uses **standard gRPC** (HTTP/2). The same
> receiver accepts both gRPC‑web and gRPC.

---

## Requirements (match the server)

- **Endpoint:** `https://localhost:50051`
- **Proto:** `proto/electrician_relay.proto`
- **RPC:** `electrician.RelayService/Receive`
- **Metadata headers (gRPC metadata):**
  - `authorization: Bearer <token>`
  - `x-tenant: local`
- **JWT claims:**
  - `iss` matches `OAUTH_ISSUER_BASE` (default `auth-service`)
  - `aud` includes `your-api`
  - `scope` includes `write:data`
- **Payload format:** JSON only (browser standardization)
- **Encryption:** AES‑GCM with 12‑byte IV prefix (required when the server has a key)
  - `metadata.security.enabled = true`
  - `metadata.security.suite = ENCRYPTION_AES_GCM`

---

## Token (dev)

Run the mock OAuth server:

```bash
go run ./example/auth/mock_oauth_server
```

Get a token:

```bash
curl -k -X POST https://localhost:3000/api/auth/session/token \
  -H 'Content-Type: application/json' \
  -d '{"scope":"write:data"}'
```

Use the returned `access_token` as the Bearer token.

---

## Cargo.toml

```toml
[dependencies]
tokio = { version = "1", features = ["macros", "rt-multi-thread"] }
tonic = { version = "0.12", features = ["tls"] }
prost = "0.13"
prost-types = "0.13"
serde_json = "1"
aes-gcm = "0.10"
rand = "0.8"
hex = "0.4"
uuid = { version = "1", features = ["v4"] }

[build-dependencies]
tonic-build = "0.12"
```

---

## Generate Rust stubs (tonic-build)

Create `build.rs` in your Rust crate:

```rust
fn main() -> Result<(), Box<dyn std::error::Error>> {
    tonic_build::configure()
        .build_client(true)
        .build_server(false)
        .compile(&["proto/electrician_relay.proto"], &["proto"])?;
    Ok(())
}
```

Place `electrician_relay.proto` in a local `proto/` folder (copy from this repo), then run:

```bash
cargo build
```

---

## Minimal Rust sender (tonic)

```rust
use aes_gcm::{Aes256Gcm, KeyInit, Nonce};
use aes_gcm::aead::Aead;
use prost_types::Timestamp;
use rand::RngCore;
use std::time::SystemTime;
use tonic::metadata::MetadataValue;
use tonic::transport::{Certificate, Channel, ClientTlsConfig};
use tonic::Request;

use relay::electrician::relay_service_client::RelayServiceClient;
use relay::electrician::{MessageMetadata, SecurityOptions, EncryptionSuite, WrappedPayload};

fn now_ts() -> Timestamp {
    let now = SystemTime::now();
    Timestamp::from(now)
}

fn aes_gcm_encrypt(plaintext: &[u8], key_hex: &str) -> Vec<u8> {
    let key = hex::decode(key_hex).expect("bad hex key");
    let cipher = Aes256Gcm::new_from_slice(&key).expect("bad key");

    let mut iv = [0u8; 12];
    rand::thread_rng().fill_bytes(&mut iv);
    let nonce = Nonce::from_slice(&iv);

    let ciphertext = cipher.encrypt(nonce, plaintext).expect("encrypt");
    let mut out = iv.to_vec();
    out.extend_from_slice(&ciphertext);
    out
}

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let token = std::env::var("RELAY_BEARER_TOKEN").expect("RELAY_BEARER_TOKEN");

    // TLS: trust the local CA used by the example server.
    // In this repo: example/relay_example/tls/ca.crt
    let ca = tokio::fs::read("example/relay_example/tls/ca.crt").await?;
    let tls = ClientTlsConfig::new()
        .ca_certificate(Certificate::from_pem(ca))
        .domain_name("localhost");

    let channel = Channel::from_static("https://localhost:50051")
        .tls_config(tls)?
        .connect()
        .await?;

    let mut client = RelayServiceClient::new(channel);

    let payload_json = serde_json::json!({
        "customerId": "rust",
        "content": "Relay test payload from Rust",
        "category": "feedback",
        "isNegative": false,
        "tags": ["rust", "grpc"]
    });
    let plaintext = serde_json::to_vec(&payload_json)?;
    let encrypted = aes_gcm_encrypt(&plaintext, "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec");

    let payload = WrappedPayload {
        id: uuid::Uuid::new_v4().to_string(),
        timestamp: Some(now_ts()),
        payload: encrypted,
        payload_encoding: 0, // PAYLOAD_ENCODING_UNSPECIFIED
        payload_type: "electrician.Feedback".to_string(),
        metadata: Some(MessageMetadata {
            content_type: "application/json".to_string(),
            security: Some(SecurityOptions {
                enabled: true,
                suite: EncryptionSuite::EncryptionAesGcm as i32,
            }),
            ..Default::default()
        }),
        ..Default::default()
    };

    let mut req = Request::new(payload);
    req.metadata_mut().insert(
        "authorization",
        MetadataValue::from_str(&format!("Bearer {}", token))?,
    );
    req.metadata_mut().insert("x-tenant", MetadataValue::from_static("local"));

    let ack = client.receive(req).await?;
    println!("{:?}", ack.into_inner());
    Ok(())
}
```

---

## Common pitfalls

- `issuer mismatch`: token `iss` must match server `OAUTH_ISSUER_BASE`.
- `x-tenant` must be **gRPC metadata**, not payload headers.
- TLS: your Rust client must trust `example/relay_example/tls/ca.crt`.
- If encryption is enabled on the server, you must encrypt payloads and set `metadata.security`.
