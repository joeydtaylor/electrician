# 📡 Receiving Relay Package

The **Receiving Relay** is Electrician’s **ingress gRPC server**.

It implements the relay service defined in the protobuf contract and turns inbound gRPC messages into submissions into your pipeline (typically a `Wire` or other `types.Submitter`).

What it is:

* a network-facing receiver (unary + streaming)
* an adapter layer from protobuf envelopes → pipeline submissions
* a place to enforce ingress policy (TLS config, optional auth checks, payload handling)

What it is not:

* a durability layer
* a “guaranteed delivery” system

---

## 📦 What it does

| Capability         | Meaning                                                                           |
| ------------------ | --------------------------------------------------------------------------------- |
| gRPC ingress       | Accept `Receive` (unary) and `StreamReceive` (bi-di streaming).                   |
| Envelope handling  | Read `WrappedPayload` / `RelayEnvelope` fields and pass along payload + metadata. |
| Output forwarding  | Forward received items to one (or more) configured downstream submitters.         |
| Transport security | Runs over gRPC transport security when you configure TLS/mTLS credentials.        |
| Acknowledgments    | Returns `StreamAcknowledgment` as application-level status.                       |

Notes:

* “Compression”, “payload encryption”, and “auth hints” exist in the schema as metadata. Whether this relay **enforces** or **interprets** those fields is an implementation/config decision.

---

## 📂 Package structure

| File                | Purpose                                              |
| ------------------- | ---------------------------------------------------- |
| `receivingrelay.go` | Type definition + constructor                        |
| `api.go`            | Public methods / component wiring                    |
| `internal.go`       | gRPC server wiring + message handling implementation |
| `options.go`        | Functional options for configuration                 |
| `*_test.go`         | Tests                                                |

(If a package doesn’t have `notify.go`, that’s intentional — telemetry hooks live next to the code that emits them.)

---

## 🧠 How receiving works

### Unary: `Receive(WrappedPayload)`

1. gRPC handler receives a `WrappedPayload`.
2. The relay extracts:

   * `payload` bytes
   * `metadata` (headers/content_type/version/trace_id/etc.)
   * `seq`, `id`, `timestamp`
3. The relay forwards the message into the configured downstream submitter(s).
4. The relay returns a `StreamAcknowledgment` indicating application-level success/failure.

### Streaming: `StreamReceive(stream RelayEnvelope)`

1. The client opens a stream and sends `StreamOpen` (optional but recommended).
2. The client sends many `WrappedPayload` messages.
3. The server may respond with acknowledgments per message or in batches depending on the negotiated `AckMode`.
4. The client sends `StreamClose`.

The streaming contract is defined in protobuf. The receiving relay’s behavior should follow that contract without inventing extra semantics.

---

## 🗜️ Compression / 🔐 Encryption / 🪪 Auth (what’s real)

The protobuf schema includes:

* `PerformanceOptions` + `CompressionAlgorithm`
* `SecurityOptions` + `EncryptionSuite`
* `AuthenticationOptions` and receiver-populated `AuthContext`

Important reality checks:

* gRPC/protobuf won’t automatically transform payload bytes based on these fields.
* Treat `AuthenticationOptions` as **advisory**. Enforcement should be server policy.
* Don’t ship secrets (client secrets, bearer tokens) in message metadata unless you control both ends and your threat model explicitly allows it.

---

## ✅ Acknowledgments are not durability

A `StreamAcknowledgment` is an application-level status signal.

It does **not** guarantee:

* persistence
* exactly-once processing
* replay across restarts

If you need those guarantees, design them explicitly (idempotency keys, durable queues/brokers, retries with backoff, etc.).

---

## 🔧 Extending the receiving relay

When you add features, keep layering clean:

* Proto changes → update `.proto`, regenerate bindings, then update receiving/forward relays.
* Cross-component behavior → update `types/receivingrelay.go` first.
* User-facing configuration knob → add an option in `options.go` and expose it through `pkg/builder`.

Add tests that cover:

* unary + streaming behavior
* cancellation + shutdown
* ack modes
* forwarding to downstream submitters
* (if implemented) compression/decompression and auth policy behavior

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Protobuf README](../../../proto/README.md)** – Full details on **Relay’s gRPC message format**.
- **[Examples Directory](../../../example/relay_example/relay_b/)** – Demonstrates **Receiving Relay in a basic real-world deployment**.
- **[Examples Directory](../../../example/relay_example/advanced_relay_b/)** – Demonstrates **Receiving Relay in a more advanced real-world deployment**.
- **[Examples Directory](../../../example/relay_example/blockchain_node/)** – Demonstrates **Receiving Relay in a contrived blockchain node deployment**.

---

## 📝 License

The **Receiving Relay package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
