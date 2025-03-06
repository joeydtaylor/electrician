# 🔌 Relay Package – gRPC Communication Layer

The **Relay package** provides **gRPC-based communication** for Electrician components, enabling **high-performance, structured message exchange** between distributed services.

This package **facilitates reliable transport** of Electrician pipeline data, ensuring **fault-tolerant message delivery** with **metadata handling, error reporting, and optional compression**.

---

## 📦 Package Overview

| Feature                       | Description                                                                  |
| ----------------------------- | ---------------------------------------------------------------------------- |
| **gRPC Streaming**            | Supports **both unary and bidirectional streaming** for efficient transport. |
| **Message Wrapping**          | Encapsulates data with **metadata, timestamps, and compression settings**.   |
| **Error Handling**            | Provides **detailed error metadata** for robust debugging.                   |
| **Tracing & Priority Queues** | Supports **trace IDs** and **priority-based message handling**.              |
| **Compression Support**       | Optimizes bandwidth with **ZSTD, Brotli, LZ4, Snappy, and Deflate**.         |

---

## 📂 Package Structure

| File                 | Purpose                                                       |
| -------------------- | ------------------------------------------------------------- |
| **relay.pb.go**      | **Compiled Go bindings** for Protobuf messages.               |
| **relay_grpc.pb.go** | **Compiled Go bindings** for gRPC client & server interfaces. |

---

## 🔧 How Relay Works

A **Relay Service** in Electrician acts as a **high-efficiency messaging layer** that supports **streaming and direct message passing** between components.

### ✅ **Core Message Structure**

Each message is **wrapped in a structured payload** that includes:

- **Unique ID** – Ensures **traceability** in distributed environments.
- **Timestamp** – Supports **event ordering** and **latency tracking**.
- **Payload Data** – The **actual message content**.
- **Metadata** – **Headers, priority, versioning, and tracing** details.
- **Error Info** – **Standardized error codes and descriptions**.

### ✅ **Key RPC Methods**

| Method            | Description                                                          |
| ----------------- | -------------------------------------------------------------------- |
| **Receive**       | Sends a single `WrappedPayload` and returns an acknowledgment.       |
| **StreamReceive** | Supports **bidirectional streaming** for real-time message exchange. |

---

## 🔧 Extending the Relay Package

To **modify or extend the relay service**, follow this structured **workflow**:

### 1️⃣ Modify `proto/electrician_relay.proto`

- **Add new fields** or **define additional RPC methods**.
- **Ensure backward compatibility** with existing messages.

### 2️⃣ Regenerate Go Bindings

- Recompile the gRPC service using the standard Protobuf tooling.

### 3️⃣ Implement New gRPC Methods

- Add new methods in `relay_grpc.pb.go` **(or manually extend via an interface wrapper)**.

### 4️⃣ Ensure Compatibility with Electrician Pipelines

- Update **relay clients** across Electrician components.
- Validate that **metadata, compression, and error handling remain intact**.

### 5️⃣ Unit Testing & Benchmarking

- **Simulate network conditions** to ensure **fault-tolerant message delivery**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Protobuf README](../../../proto/README.md)** – Full details on the Electrician **gRPC message format**.
- **[Examples Directory](../../../example/relay_example/)** – Demonstrates **Relay’s role in distributed processing**.

---

## 📝 License

The **Relay package** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
