# ⚡ Wire Package

The **Wire** package is the **core processing unit** of Electrician, responsible for **data ingestion, transformation, and forwarding** in a **concurrent and structured manner**.

Wires act as the foundation of Electrician’s **event-driven architecture**, enabling seamless integration with **circuit breakers, sensors, generators, loggers, and surge protectors**.

---

## 📦 Package Overview

The Wire package provides a **modular, high-performance framework** for handling **data pipelines**. It includes **component connectivity, concurrency control, and lifecycle management** to ensure reliability.

| Feature                       | Description                                                                                       |
| ----------------------------- | ------------------------------------------------------------------------------------------------- |
| **Component Connectivity**    | Connects to **circuit breakers, generators, loggers, transformers, and surge protectors**.        |
| **Concurrency Management**    | Supports **high-throughput, multi-routine** data processing with configurable concurrency limits. |
| **Immutability Model**        | Pipelines are **pre-configured**, ensuring deterministic execution (no runtime modifications).    |
| **Fault Tolerance**           | Uses **insulators, retries, and circuit breakers** to recover from failures.                      |
| **Event Logging & Telemetry** | Supports **real-time monitoring** via **sensors and structured logging**.                         |

---

## 📂 Package Structure

Each file in the `wire` package follows Electrician’s **structured approach**, separating concerns across **public APIs, private internals, logging, and options**.

| File             | Purpose                                                                                              |
| ---------------- | ---------------------------------------------------------------------------------------------------- |
| **api.go**       | Public methods for interacting with Wires (**connect components, manage lifecycle**).                |
| **internal.go**  | Low-level implementation details, **hidden from external use**.                                      |
| **notify.go**    | Handles **event logging, sensor notifications, and telemetry hooks**.                                |
| **options.go**   | Functional options for configuring Wires in a **composable, declarative manner**.                    |
| **wire.go**      | The **primary implementation** of Wires, including concurrency, buffering, and transformation logic. |
| **wire_test.go** | Unit tests ensuring **correctness, performance, and fault tolerance**.                               |

---

## 🔧 How Wires Work

A **Wire** is responsible for **receiving, transforming, and outputting data** while managing **concurrency, component integration, and lifecycle control**.

### ✅ **Connection Methods**

- `ConnectCircuitBreaker(cb CircuitBreaker[T])` – Attach a circuit breaker to regulate flow.
- `ConnectGenerator(generator Generator[T])` – Attach one or more data sources.
- `ConnectLogger(logger Logger)` – Attach logging components.
- `ConnectSensor(sensor Sensor[T])` – Attach sensors to monitor performance.
- `ConnectSurgeProtector(protector SurgeProtector[T])` – Attach a surge protection mechanism.
- `ConnectTransformer(transformer Transformer[T])` – Attach a transformation function.

### ✅ **Lifecycle Management**

- `Start(ctx context.Context) error` – Initiate the Wire's processing.
- `Stop() error` – Gracefully shut down the Wire.
- `Restart(ctx context.Context) error` – Fully restart the Wire, applying new configurations.

### ✅ **Data Handling**

- `Submit(ctx context.Context, elem T) error` – Push data into the Wire for processing.
- `LoadAsJSONArray() ([]byte, error)` – Retrieve processed output as a JSON array.
- `Load() *bytes.Buffer` – Retrieve processed output as raw bytes.
- `GetInputChannel() chan T` / `GetOutputChannel() chan T` – Get the Wire’s data channels.

### ✅ **Concurrency & Configuration**

- `SetConcurrencyControl(bufferSize int, maxRoutines int)` – Configure buffer size and max concurrency.
- `SetEncoder(e Encoder[T])` – Define a serialization mechanism.
- `SetComponentMetadata(name string, id string)` – Assign metadata for logging and tracking.

---

## 🔒 Immutability & Best Practices

Electrician **does not support runtime modifications** to components once a pipeline has started.

✔️ **Pipelines should be fully configured before execution.**  
✔️ **Components must be connected before calling `Start()`.**  
✔️ **All modifications should happen at initialization using functional options.**

This ensures:

- **Predictability** – Pipelines behave consistently without unexpected state changes.
- **Concurrency Safety** – Eliminates race conditions and synchronization issues.
- **Performance Optimization** – Avoids costly reconfiguration overhead.

While Go technically allows modifying running components, **Electrician strongly discourages this**. The framework is designed for **deterministic, event-driven pipelines** that are **pre-configured and stable**.

---

## ⚡ Standard Library First

Electrician is **99% based on the Go standard library**.  
The **only** external dependencies used in the Wire package are:

- **`zap` (Uber’s logging library)** – Used in `internallogger/` for **high-performance structured logging**.
- **`protobuf` (gRPC and serialization)** – Used in `relay/` to support **cross-service messaging**.

Everything else—including **networking, data transformation, and concurrency management**—is built using **pure Go**, ensuring:

✅ **Maximum compatibility** – No unnecessary dependencies.  
✅ **Minimal attack surface** – Secure and easy to audit.  
✅ **High performance** – Optimized for **low-latency, high-throughput pipelines**.

---

## 🔧 Extending the Wire Package

To **add new functionality**, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- If a new method is needed, **update the interface** in `types/wire.go`.
- This ensures **consistent contracts** across components.

### 2️⃣ Implement in `api.go`

- Add the **actual implementation** of the new method.

### 3️⃣ Create a Functional Option in `options.go`

- Supports **composable configuration** alongside traditional method calls.

### 4️⃣ Enhance `notify.go` (if applicable)

- If your change introduces **new events**, **add corresponding logging and telemetry hooks**.

### 5️⃣ Unit Testing (`wire_test.go`)

- Ensure new functionality is **fully covered** before merging.

By enforcing these steps, Electrician maintains **consistency, safety, and extensibility**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – In-depth explanation of how the `internal/` directory works.
- **[Examples Directory](../../../example/wire_example/)** – Real-world use cases demonstrating Wires in action.

---

## 📝 License

The **Wire package** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
