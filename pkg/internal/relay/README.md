# 🔌 Relay Package (gRPC + Protobuf bindings)

The `relay` package contains the **generated Go bindings** for Electrician’s relay protocol.

This package is intentionally thin:

* ✅ protobuf message types used across Electrician services
* ✅ gRPC client/server interfaces for the relay service
* ❌ no “reliable delivery” guarantees by itself
* ❌ no compression/encryption/auth implementation (those are handled by the relays/adapters that *use* these types)

If you’re looking for implementation logic, it lives in packages like `forwardrelay/` and `receivingrelay/`. This package is the shared contract.

---

## 📦 What’s in this package

| File               | What it is                                |
| ------------------ | ----------------------------------------- |
| `relay.pb.go`      | Generated Go types for messages/enums     |
| `relay_grpc.pb.go` | Generated gRPC client + server interfaces |

These files are generated from the repo’s `.proto` definition. Do not hand-edit them.

---

## 🧬 Key message types

### `WrappedPayload`

A single envelope around an opaque `payload` (`bytes`) plus context:

* `id`, `timestamp`, `seq`
* `payload` bytes
* `metadata` (headers/content_type/version/trace_id/priority + perf/security/auth fields)
* `error_info` (optional structured error attachment)

`payload` is deliberately unopinionated: Electrician treats it as bytes. The meaning is conveyed via `content_type` and headers.

### `RelayEnvelope` (streaming)

Bidirectional streaming uses a `oneof` envelope:

* `StreamOpen` — establishes defaults + ack strategy
* `WrappedPayload` — the data messages
* `StreamClose` — close reason

### `StreamAcknowledgment`

Application-level ack fields used by unary and streaming calls:

* `success`, `message`, `code`, `retryable`
* `id`, `seq`, `stream_id`
* batch/stream summary fields: `last_seq`, `ok_count`, `err_count`

Acks are **application status**, not transport guarantees.

---

## 🗜️ Compression / 🔐 Encryption / 🪪 Auth (metadata, not magic)

The schema includes fields for:

* compression preferences (`PerformanceOptions`, `CompressionAlgorithm`)
* payload encryption declaration (`SecurityOptions`, `EncryptionSuite`)
* auth hints (`AuthenticationOptions`) and verified auth facts (`AuthContext`)

Important:

* Protobuf/gRPC does not automatically compress or encrypt your `payload` based on these fields.
* These fields describe **intent and context**. Implementations decide what to enforce.
* `priority` and similar fields are **hints**; there is no built-in priority scheduling in this package.

---

## 🛰️ gRPC service surface

The generated service is:

* `Receive(WrappedPayload) -> StreamAcknowledgment`
* `StreamReceive(stream RelayEnvelope) <-> stream StreamAcknowledgment`

Server implementations should live outside this generated package (e.g., in a relay/adapter component). Do not modify generated files to add behavior.

---

## 🔧 Making changes (safe workflow)

1. Update the `.proto` file.
2. Regenerate bindings.
3. Update the server/client implementations that use the new fields.

Compatibility rules:

* never reuse field numbers
* only add new fields with new numbers
* keep enums append-only
* don’t change meaning of existing fields

---

## 🛠️ Regenerating Go bindings

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

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Protobuf README](../../../proto/README.md)** – Full details on the Electrician **gRPC message format**.
- **[Examples Directory](../../../example/relay_example/)** – Demonstrates **Relay’s role in distributed processing**.

---

## 📝 License

The **Relay package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
