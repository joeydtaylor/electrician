# ⚡ Electrician Relay Protocol (Protobuf + gRPC)

This directory defines Electrician’s **relay envelope** and **gRPC service contract**.

It’s a **canonical wire format** for moving opaque application payloads (`bytes`) between Electrician components (in-process, over LAN, or across services). The protobuf schema describes *structure and intent*; enforcement of auth/crypto/compression is implemented by the relay/adapters and depends on configuration.

---

## 📌 Where the Go code lives

The `.proto` file uses:

```proto
package electrician;
option go_package = "github.com/joeydtaylor/electrician/pkg/internal/relay";
```

Generated Go bindings land in:

* `pkg/internal/relay`

---

## 🧬 Core types

### `WrappedPayload`

One message = one envelope around an opaque `payload`:

* `id` — message identifier (string)
* `timestamp` — event time (`google.protobuf.Timestamp`)
* `payload` — raw bytes (may be compressed/encrypted depending on metadata + implementation)
* `metadata` — headers, content type, perf/security/auth options
* `error_info` — optional error attachment (if the sender wants to include structured error details)
* `seq` — sender-defined sequence number (commonly used for ordering/ack tracking)

`payload` is intentionally unopinionated: Electrician treats it as bytes; `content_type` and headers explain what those bytes mean.

---

### `MessageMetadata`

Per-message context and options:

* `headers` — arbitrary key/value metadata
* `content_type` — MIME-ish string for `payload` (e.g. `application/json`, `application/x-protobuf`, etc.)
* `version` — `major`/`minor` for payload/schema compatibility
* `performance` — compression preferences
* `trace_id` — distributed tracing correlation
* `priority` — simple integer priority hint
* `security` — optional payload-level encryption declaration
* `authentication` — optional sender→receiver auth expectations (advisory)
* `auth_context` — receiver-emitted facts after authentication (no raw tokens)

Important: these fields are part of the wire format; what a receiver actually enforces is a **policy decision**.

---

### `ErrorInfo`

Structured error attachment:

* `code` — int status code (project-defined)
* `message` — human-readable summary
* `details` — freeform strings

This is not the same as an RPC failure. It’s an optional part of the payload envelope.

---

## 🗜️ Compression (PerformanceOptions)

Compression is expressed as **metadata** plus an algorithm enum:

* `use_compression`
* `compression_algorithm`
* `compression_level`

```proto
enum CompressionAlgorithm {
  COMPRESS_NONE = 0;
  COMPRESS_DEFLATE = 1;
  COMPRESS_SNAPPY = 2;
  COMPRESS_ZSTD = 3;
  COMPRESS_BROTLI = 4;
  COMPRESS_LZ4 = 5;
}
```

Semantics:

* If compression is enabled, `payload` is expected to contain the compressed bytes for the selected algorithm.
* The enum is the **wire label**; the actual codec implementation is adapter/relay-specific.
* Receivers may ignore compression hints and enforce their own behavior.

---

## 🔐 Payload encryption (SecurityOptions)

`SecurityOptions` is a **declaration** about the `payload` bytes:

```proto
enum EncryptionSuite {
  ENCRYPTION_NONE = 0;
  ENCRYPTION_AES_GCM = 1;
}
```

Semantics:

* If `security.enabled=true`, the sender is declaring that the `payload` bytes are encrypted using `suite`.
* Protobuf does **not** provide encryption; it only carries bytes + metadata.
* Transport security (TLS/mTLS) is separate from payload encryption.

Do not oversell this: encryption support depends on the concrete relay/adapter implementation and key management strategy.

---

## 🪪 Auth hints + verified auth facts

### `AuthenticationOptions` (sender intent / advisory)

This message lets a sender communicate expected auth mode and constraints:

* `AUTH_NONE`
* `AUTH_OAUTH2`
* `AUTH_MUTUAL_TLS`

OAuth2 and mTLS sub-messages include fields for issuer/audience/scope requirements and principal allowlists.

**Critical note:** Some fields (e.g. introspection credentials) are inherently sensitive. Electrician treats authentication *policy* and credentials as configuration concerns; do not ship secrets inside message metadata unless you control both ends and have an explicit threat model.

### `AuthContext` (receiver output)

Populated by the receiver after validation:

* mode + authenticated
* principal/subject/client_id
* scopes + curated claims
* expires_at / issuer / audience / token_id

This is for auditing and downstream policy decisions. It should not contain raw tokens.

---

## 🌊 Streaming protocol (`RelayEnvelope`)

For high-throughput streams, the protocol uses a `oneof` envelope:

* `StreamOpen` — establishes stream defaults + ack strategy
* `WrappedPayload` — the data messages
* `StreamClose` — closes the stream with an optional reason

### `StreamOpen`

Key fields:

* `stream_id` — logical stream identifier
* `defaults` — default `MessageMetadata` applied to messages
* `ack_mode` — ack behavior
* `ack_every_n` — for batch ack
* `max_in_flight` — flow-control hint
* `omit_payload_metadata` — if true, payload messages can omit metadata to reduce overhead (defaults apply)

---

## ✅ Acknowledgments (`StreamAcknowledgment`)

Acknowledgment fields support both unary and streaming:

* `success` + `message`
* `stream_id`, `id`, `seq`
* `code` + `retryable`
* `last_seq`, `ok_count`, `err_count` (useful for batch/stream summaries)

Receivers should treat ack as **best-effort application status**, not a replacement for transport reliability.

---

## 🛰️ gRPC service

```proto
service RelayService {
  rpc Receive(WrappedPayload) returns (StreamAcknowledgment);
  rpc StreamReceive(stream RelayEnvelope) returns (stream StreamAcknowledgment);
}
```

* `Receive` is a unary “send one, get one ack”.
* `StreamReceive` is bidirectional streaming for continuous/high-volume pipelines.

---

## 🛠️ Generating Go code

From repo root (assuming the proto is under `proto/`):

```bash
protoc \
  -I=proto \
  --go_out=pkg/internal/relay \
  --go_opt=paths=source_relative \
  --go-grpc_out=pkg/internal/relay \
  --go-grpc_opt=paths=source_relative \
  proto/electrician_relay.proto
```

If you move/rename files, keep field numbers stable and avoid breaking the `go_package` path.

---

## 🔁 Compatibility rules (non-negotiable)

* Never reuse field numbers.
* Only add new fields with new numbers.
* Don’t change the meaning of existing fields.
* Keep enums append-only.

---

## 📝 License

[Apache 2.0 License](../LICENSE).  

---
