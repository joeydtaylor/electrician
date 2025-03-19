# 📡 Forward Relay Package

The **Forward Relay** package enables **secure, high-performance data transmission** in distributed systems.  
It acts as a **message router**, forwarding data from **input sources** to **target destinations** over gRPC.

Forward Relays support **compression, TLS encryption, and configurable logging**,  
ensuring **reliable and efficient data forwarding** between system components.

---

## 📦 Package Overview

| Feature                 | Description                                                           |
| ----------------------- | --------------------------------------------------------------------- |
| **Multi-Input Support** | Receives data from **multiple sources** (e.g., APIs, data streams).   |
| **High Throughput**     | Optimized for **parallel data processing and transmission**.          |
| **Compression Support** | Reduces bandwidth usage with **gzip, zstd, snappy, brotli, and LZ4**. |
| **TLS Encryption**      | Secures data transmission with **configurable TLS settings**.         |
| **Performance Tuning**  | Supports **batching, buffering, and rate control options**.           |

---

## ⚡ Notable Dependencies

Electrician is **almost entirely built on the Go standard library**—with two key exceptions:

1. **Logging:** Uses [Zap](https://github.com/uber-go/zap), widely regarded as the best structured logger for Go.
2. **Compression:** Integrates **widely used compression libraries** (ZSTD, Snappy, Brotli, LZ4, Deflate)  
   to **optimize bandwidth and processing efficiency**.

These dependencies were chosen carefully to **maximize performance while minimizing external bloat**.

---

## 📂 Package Structure

| File                     | Purpose                                                             |
| ------------------------ | ------------------------------------------------------------------- |
| **api.go**               | Public API for **starting, stopping, and managing** Forward Relays. |
| **internal.go**          | Low-level logic for **handling TLS, compression, and data flow**.   |
| **notify.go**            | Handles **event logging and monitoring**.                           |
| **options.go**           | Functional options for **composable configuration**.                |
| **forwardrelay.go**      | Core **Type Definition and Constructor**.                           |
| **forwardrelay_test.go** | Unit tests ensuring **performance, reliability, and security**.     |

---

## 🔧 How Forward Relays Work

A **Forward Relay** acts as an **intelligent data router**,  
efficiently transmitting messages between services.

### ✅ **Key Mechanisms**

- **Multi-Input Handling:** Receives data from **one or more sources**.
- **Compression Support:** **Reduces payload size** for optimized network performance.
- **Secure TLS Encryption:** Ensures **confidentiality and data integrity**.
- **Dynamic Performance Tuning:** Adjusts **batch sizes, compression, and concurrency**.
- **Event-Driven Logging & Monitoring:** Integrated with **Zap logger and sensors**.

### ✅ **Lifecycle Management**

| Method           | Description                                          |
| ---------------- | ---------------------------------------------------- |
| `Start()`        | Begins **data forwarding** from connected sources.   |
| `Stop()`         | Gracefully **terminates all forwarding operations**. |
| `ConnectInput()` | Attaches **one or more data sources** dynamically.   |
| `Submit()`       | **Forwards data** to target addresses securely.      |

---

## 🔧 Extending the Forward Relay Package

To **add new functionality**, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- Define the new **interface method** inside `types/forwardrelay.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file must now implement the new method.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for event logging

- If your change introduces **new events**, update **logging and sensor hooks**.

### 5️⃣ Unit Testing (`forwardrelay_test.go`)

- **Ensure performance, event handling, and security are tested**.

---

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
