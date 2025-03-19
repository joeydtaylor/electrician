# 🔌 Plug Package

The **Plug** package provides **adapter-based connectivity** for integrating external data sources into Electrician pipelines.  
It acts as a **bridge** between raw data sources and processing components, allowing for **flexible ingestion and transformation**.

---

## 📦 Package Overview

| Feature                         | Description                                                                 |
| ------------------------------- | --------------------------------------------------------------------------- |
| **Adapter-Based Connectivity**  | Supports **multiple input sources**, such as APIs, databases, and files.    |
| **Functional Adapters**         | Allows dynamic **adapter function injection** for **custom data handling**. |
| **Sensor & Logger Integration** | Supports **event monitoring and structured logging**.                       |
| **Concurrency Safe**            | Uses **lock-based synchronization** for **thread-safe operations**.         |

---

## 📂 Package Structure

| File             | Purpose                                                                     |
| ---------------- | --------------------------------------------------------------------------- |
| **api.go**       | Public API methods for configuring **Plugs and Adapters**.                  |
| **internal.go**  | Low-level logic for managing **adapter execution and connection handling**. |
| **notify.go**    | **Logging, telemetry, and sensor event hooks**.                             |
| **options.go**   | Functional options for **Plug configuration**.                              |
| **plug.go**      | Core **Type Definition and Constructor**.                                   |
| **plug_test.go** | Unit tests ensuring **data ingestion reliability and adapter correctness**. |

---

## 🔧 How Plugs Work

A **Plug** is an **ingestion layer** that connects **external data sources** to Electrician's processing components.

### ✅ **Core Responsibilities**

- **Adapter Management:** Handles **multiple adapter functions** for different data sources.
- **Input Normalization:** Ensures **data consistency** before forwarding it into the pipeline.
- **Sensor & Logger Support:** Enables **real-time monitoring and debugging**.

### ✅ **Lifecycle Management**

| Method              | Description                                   |
| ------------------- | --------------------------------------------- |
| `ConnectAdapter()`  | Attaches **external adapters** to fetch data. |
| `AddAdapterFunc()`  | Registers **adapter functions** dynamically.  |
| `GetConnectors()`   | Retrieves **all connected adapters**.         |
| `GetAdapterFuncs()` | Lists **all registered adapter functions**.   |

---

## 🔧 Extending the Plug Package

To **add new functionality** to the Plug package, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- Define new methods inside `types/plug.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file contains **public API methods** – update it accordingly.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for logging & telemetry

- If new events are introduced, add **sensor and logger hooks**.

### 5️⃣ Unit Testing (`plug_test.go`)

- **Ensure data adapters behave correctly under real-world conditions**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/plug_example/)** – Demonstrates **Plug usage with real-world adapters**.

---

## 📝 License

The **Plug package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
