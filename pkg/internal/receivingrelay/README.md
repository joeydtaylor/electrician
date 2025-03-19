# 📡 Receiving Relay Package

The **Receiving Relay** package handles **incoming data streams** in Electrician, ensuring **secure, efficient, and scalable message ingestion** within distributed systems.

It provides **network listening, TLS security, and real-time data forwarding**, making it an essential component for **high-throughput, event-driven architectures**.

---

## 📦 Package Overview

| Feature                       | Description                                                                   |
| ----------------------------- | ----------------------------------------------------------------------------- |
| **gRPC-Based Communication**  | Handles **high-performance streaming and request/response** data exchange.    |
| **TLS Security**              | Supports **encryption and authentication** via **configurable TLS settings**. |
| **Flexible Output Routing**   | Routes messages to multiple downstream **Submitters** dynamically.            |
| **Compression Handling**      | Supports **ZSTD, Brotli, Snappy, LZ4, and Deflate** for bandwidth efficiency. |
| **Event Logging & Telemetry** | Provides **detailed structured logging and real-time monitoring**.            |

---

## ⚡ Notable Dependencies

Electrician is **almost entirely built on the Go standard library**—with two key exceptions:

1. **Logging:** Uses [Zap](https://github.com/uber-go/zap), which is widely regarded as the best structured logger for Go.
2. **Compression:** Integrates **widely adopted compression libraries** (ZSTD, Snappy, Brotli, LZ4, Deflate) to **optimize bandwidth and processing efficiency**.

These external dependencies were chosen carefully to **maximize performance and minimize bloat**.

---

## 📂 Package Structure

| File                       | Purpose                                                                      |
| -------------------------- | ---------------------------------------------------------------------------- |
| **api.go**                 | Public API methods for configuring and managing the Receiving Relay.         |
| **internal.go**            | Handles **data decompression, TLS setup, and gRPC connection management**.   |
| **notify.go**              | Event logging and **sensor integration** for system-wide telemetry.          |
| **options.go**             | Functional options for declarative **relay configuration**.                  |
| **receivingrelay.go**      | Core **Type Definition and Constructor**.                                    |
| **receivingrelay_test.go** | Unit tests for **correctness, fault tolerance, and networking reliability**. |

---

## 🔧 How Receiving Relays Work

A **Receiving Relay** acts as an **ingress point** for Electrician pipelines, **listening for messages** and **forwarding them to downstream consumers**.

### ✅ **Core Responsibilities**

- **Data Reception:** Accepts incoming **gRPC messages** via direct calls or streaming.
- **Decompression Support:** **Unpacks compressed payloads** based on metadata.
- **TLS-Encrypted Communication:** Ensures **secure, authenticated transport**.
- **Dynamic Output Routing:** Forwards messages to **multiple destinations**.

### ✅ **Lifecycle Management**

| Method            | Description                                                       |
| ----------------- | ----------------------------------------------------------------- |
| `Start()`         | Initiates the relay, **binding it to a network address**.         |
| `Receive()`       | Processes **single incoming messages** and sends acknowledgments. |
| `StreamReceive()` | Enables **continuous bidirectional streaming**.                   |
| `Stop()`          | Gracefully shuts down the relay, **closing all connections**.     |

---

## 🔧 Extending the Receiving Relay Package

To **modify or extend the Receiving Relay**, follow this **structured workflow**:

### 1️⃣ Modify `types/`

- Define new methods inside `types/receivingrelay.go`.
- This ensures **all implementations remain consistent** across Electrician.

### 2️⃣ Implement in `api.go`

- The `api.go` file contains **public methods** – update it accordingly.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for logging & telemetry

- If new events are introduced, add **sensor and logger hooks**.

### 5️⃣ Unit Testing (`receivingrelay_test.go`)

- **Validate reliability, performance, and network stability** under real-world conditions.

---

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
