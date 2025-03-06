# ⚡ Generator Package

The **Generator** package provides **data production** capabilities in Electrician.  
It acts as a **source of data**, generating and forwarding elements to processing pipelines.

Generators can produce data from **plugs, external adapters, or internal sources**,  
and distribute it to **connected components**, ensuring seamless data flow.

---

## 📦 Package Overview

| Feature                    | Description                                                     |
| -------------------------- | --------------------------------------------------------------- |
| **Pluggable Architecture** | Supports **multiple input adapters** for dynamic data sourcing. |
| **Fault-Tolerant Design**  | Integrates **circuit breakers** to prevent system overload.     |
| **Concurrency Management** | Efficiently handles **multiple data streams** in parallel.      |
| **Sensor Integration**     | Supports **real-time monitoring** and event-driven callbacks.   |
| **Restart Mechanism**      | Allows **safe stop-and-restart** for live data adjustments.     |

---

## 📂 Package Structure

| File                  | Purpose                                                                 |
| --------------------- | ----------------------------------------------------------------------- |
| **api.go**            | Public API methods for **starting, stopping, and managing** generators. |
| **internal.go**       | Low-level logic for **handling concurrent data generation**.            |
| **notify.go**         | Handles **event logging and sensor notifications**.                     |
| **options.go**        | Functional options for **composable configuration**.                    |
| **generator.go**      | Core **Type Definition and Constructor**.                               |
| **generator_test.go** | Unit tests for **performance, reliability, and error handling**.        |

---

## 🔧 How Generators Work

Generators **produce data** and send it to downstream components.

### ✅ **Key Mechanisms**

- **Plugs & Adapters:** Connect to **various data sources** (files, APIs, in-memory functions).
- **Circuit Breaker Protection:** Prevents data flooding by **pausing under high load conditions**.
- **Controlled Concurrency:** Manages **multiple concurrent sources** efficiently.
- **Event-Driven Architecture:** Reacts to **external triggers and system events**.

### ✅ **Lifecycle Management**

| Method          | Description                                       |
| --------------- | ------------------------------------------------- |
| `Start()`       | Begins **data generation** from attached sources. |
| `Stop()`        | Gracefully **halts** all active generation.       |
| `Restart()`     | Stops and **restarts** the generator safely.      |
| `ConnectPlug()` | Attaches a **new data source** dynamically.       |

---

## 🔧 Extending the Generator Package

To **add new functionality**, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- Define the new **interface method** inside `types/generator.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file must now implement the new method.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for event logging

- If your change introduces **new events**, update **logging and sensor hooks**.

### 5️⃣ Unit Testing (`generator_test.go`)

- **Ensure performance, event handling, and failure conditions are tested**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.md)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/generator_example/)** – Demonstrates real-world **Generator usage**.

---

## 📝 License

The **Generator package** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy generating! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
