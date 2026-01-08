# 📡 Forward Relay Package

The **Forward Relay** is Electrician’s **egress gRPC client**.

It takes items from your local pipeline, wraps them in the Electrician relay envelope (`WrappedPayload` / `RelayEnvelope`), and forwards them to a remote `RelayService`.

This package is about **transporting bytes + metadata**. Delivery guarantees, retry strategy, compression/encryption behavior, and auth policy are implementation/config decisions.

---

## ✅ What it does

| Capability              | Meaning                                                                         |
| ----------------------- | ------------------------------------------------------------------------------- |
| 🛰️ gRPC egress         | Connect to a remote relay server and send messages via unary or streaming RPCs. |
| 📦 Envelope wrapping    | Populate `id`, `timestamp`, `seq`, `metadata`, and `payload` bytes.             |
| ✅ Acknowledgments       | Read `StreamAcknowledgment` as application-level status from the receiver.      |
| 🔐 Transport security   | Use TLS/mTLS on the gRPC channel when configured.                               |
| 🧾 Metadata propagation | Forward headers/content type/version/trace/priority hints to receivers.         |

What it does **not** do by definition:

* durable delivery
* exactly-once semantics
* “priority queues” or guaranteed ordering across restarts

---

## 🧠 How it fits in a pipeline

Common shape:

**Wire → ForwardRelay → Remote ReceivingRelay**

* Your local pipeline produces values.
* The forward relay turns those into relay envelopes.
* The remote receiving relay ingests and submits into the remote pipeline.

Forward relay is a sink/egress adapter. It should not contain business logic.

---

## 🧬 Message model (what actually goes on the wire)

The relay protocol is defined in protobuf:

* `WrappedPayload` is the core envelope.
* `RelayEnvelope` wraps streaming messages (`StreamOpen`, `payload`, `StreamClose`).

Key fields you should care about:

* `payload` is `bytes` — you decide serialization (`application/json`, protobuf, msgpack, custom).
* `metadata.content_type` describes the payload bytes.
* `metadata.headers` carries arbitrary key/value context.
* `trace_id`, `priority`, and `version` are hints that receivers may use.

---

## 🗜️ Compression / 🔐 Encryption / 🪪 Auth (metadata vs enforcement)

The schema includes fields for:

* compression preferences (`PerformanceOptions`, `CompressionAlgorithm`)
* payload encryption declaration (`SecurityOptions`, `EncryptionSuite`)
* auth hints (`AuthenticationOptions`) and receiver-populated `AuthContext`

Reality checks:

* Protobuf/gRPC will not automatically compress/encrypt your payload based on these fields.
* Transport security (TLS/mTLS) is separate from payload encryption.
* `AuthenticationOptions` is **advisory**. Enforcement belongs to the receiver’s configured policy.
* Don’t ship secrets (client secrets, bearer tokens) in metadata unless both ends explicitly agree and your threat model allows it.

---

## ✅ Acks are application status, not durability

`StreamAcknowledgment` tells you what the receiver claims happened.

It does **not** guarantee persistence or replay across failures.

If you need stronger guarantees, design them explicitly:

* idempotency keys (`id`/`seq` usage)
* retry with backoff
* durable upstream queues (Kafka/SQS/etc.)

---

## ⚙️ Lifecycle + configuration contract

Forward relays follow Electrician’s standard operational model:

✅ Configure → Start → Submit/Run → Stop/Restart

* Configure target address/credentials/options before `Start()`.
* Don’t mutate configuration while running.
* Respect contexts on submit paths (cancellation/shutdown behavior should be clean).

---

## 📂 Package structure

| File              | Purpose                                  |
| ----------------- | ---------------------------------------- |
| `forwardrelay.go` | Type definition + constructor            |
| `api.go`          | Public methods / wiring                  |
| `internal.go`     | gRPC client + send/stream implementation |
| `options.go`      | Functional options for configuration     |
| `*_test.go`       | Tests                                    |

---

## 🔧 Extending the forward relay

* Proto change → update `.proto`, regenerate bindings, then update send/receive relays.
* Cross-component contract → update `types/forwardrelay.go` first.
* User-facing knob → expose via `pkg/builder` (`ForwardRelayWith…`).

Tests should cover:

* unary send path
* streaming path + ack handling
* cancellation/shutdown
* error propagation (network errors vs receiver errors)
* (if implemented) payload compression/encryption handling

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.md)** – How `internal/` packages interact with `types/`.
- **[Protobuf README](../../../proto/README.md)** – Full details on **Relay’s gRPC message format**.
- **[Examples Directory](../../../example/relay_example/relay_a/)** – Demonstrates **Forward Relay in a basic real-world deployment**.
- **[Examples Directory](../../../example/relay_example/advanced_relay_a/)** – Demonstrates **Forward Relay in a more advanced real-world deployment**.
- **[Examples Directory](../../../example/relay_example/blockchain_hub/)** – Demonstrates **Forward Relay in a contrived blockchain hub deployment**.

---

## 📝 License

The **Forward Relay package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy forwarding! ⚡📡** If you have questions or need support, feel free to open a GitHub issue.
