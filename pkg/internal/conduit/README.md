# 🔌 Conduit Package

The **Conduit** package provides a **modular, high-performance data processing pipeline**  
that efficiently routes data through multiple **Wires**, **Circuit Breakers**, **Generators**, and **Surge Protectors**.

Conduits act as **central processing hubs**, enabling **composable, fault-tolerant workflows**  
within **streaming, batch, and event-driven architectures**.

---

## 📦 Package Overview

| Feature                     | Description                                                        |
| --------------------------- | ------------------------------------------------------------------ |
| **Multi-Wire Processing**   | Supports **chained, parallel, or sequential** data pipelines.      |
| **Circuit Breaker Support** | Prevents failures from cascading by isolating **faulty segments**. |
| **Surge Protection**        | Dynamically throttles throughput during **high-load conditions**.  |
| **Generator Integration**   | Seamlessly connects **Plug-based data sources**.                   |
| **Logging & Telemetry**     | **Zap-based logging** + **Sensor metrics** for real-time insights. |

---

## 📂 Package Structure

| File                | Purpose                                                              |
| ------------------- | -------------------------------------------------------------------- |
| **api.go**          | Public API for **configuring and managing Conduits**.                |
| **options.go**      | Functional options for **customizing conduit behavior**.             |
| **conduit.go**      | Core **Type Definition and Constructor**.                            |
| **conduit_test.go** | Unit tests ensuring **stability, performance, and fault tolerance**. |

---

## ⚡ Notable Dependencies

Electrician is **almost entirely built on the Go standard library**, with two key exceptions:

1. **Logging:** Uses [Zap](https://github.com/uber-go/zap) – the most performant structured logger for Go.
2. **Compression & Encoding (Protobuf Relay Only):**  
   Uses widely adopted **ZSTD, Snappy, Brotli, LZ4, and Deflate** for optimized performance.

Everything else is built **directly on Go’s standard library**, keeping Electrician **lean, efficient, and self-contained**.

---

## 🔧 How Conduits Work

A **Conduit** acts as a **data processing hub**, coordinating **wires, sensors, and flow control mechanisms**.

### ✅ **Key Mechanisms**

- **Multi-Wire Execution:** Supports **parallel and sequential processing chains**.
- **Dynamic Data Flow:** Routes elements **through multiple transformation steps**.
- **Concurrency & Buffering:** Configurable **throughput and parallelism controls**.
- **Integrated Error Handling:** Works with **Circuit Breakers, Surge Protectors, and Resisters**.

### ✅ **Lifecycle Management**

| Method     | Description                                                  |
| ---------- | ------------------------------------------------------------ |
| `Start()`  | Begins **data processing**, activating connected components. |
| `Stop()`   | Gracefully **shuts down** all managed wires.                 |
| `Submit()` | Pushes data **into the processing pipeline**.                |
| `Load()`   | Retrieves **processed output from the last wire**.           |

---

## 🔧 Extending the Conduit Package

To **add new functionality**, follow this **structured workflow**:

### 1️⃣ Modify `types/`

- Define the new **interface method** inside `types/conduit.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file must now **implement the new method**.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for event logging

- If your change introduces **new events**, update **logging and sensor hooks**.

### 5️⃣ Unit Testing (`conduit_test.go`)

- **Ensure performance, event handling, and fault tolerance are tested**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.md)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/conduit_example/)** – Demonstrates **Conduits in a real-world pipeline**.

---

## 📝 License

The **Conduit package** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚡🔗** If you have questions or need support, feel free to open a GitHub issue.
