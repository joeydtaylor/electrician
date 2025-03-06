# 🛑 Circuit Breaker Package

The **Circuit Breaker** package prevents **system overloads and cascading failures**  
by **blocking failing operations** after reaching a predefined error threshold.

It provides **automatic resets, neutral wires for buffering, and structured logging**,  
ensuring **resilience and fault tolerance** in distributed systems.

---

## 📦 Package Overview

| Feature                         | Description                                                       |
| ------------------------------- | ----------------------------------------------------------------- |
| **Automatic Failure Detection** | Opens the circuit when error count exceeds a threshold.           |
| **Time-Based Resets**           | **Automatically resets** after a cooldown period.                 |
| **Neutral Wire Support**        | Routes traffic to **backup processing channels** during failures. |
| **Event Logging & Telemetry**   | Logs failures, resets, and trips **for real-time monitoring**.    |
| **Thread-Safe Design**          | Uses **atomic operations and locks** for concurrency safety.      |

---

## 📂 Package Structure

| File                       | Purpose                                                                       |
| -------------------------- | ----------------------------------------------------------------------------- |
| **api.go**                 | Public API for **tripping, resetting, and managing** Circuit Breakers.        |
| **internal.go**            | Low-level logic for **error tracking, state transitions, and lock handling**. |
| **options.go**             | Functional options for **configurable Circuit Breakers**.                     |
| **circuitbreaker.go**      | Core **Type Definition and Constructor**.                                     |
| **circuitbreaker_test.go** | Unit tests ensuring **fault tolerance, resets, and performance**.             |

---

## 🔧 How Circuit Breakers Work

A **Circuit Breaker** **monitors failures**,  
blocking operations when error rates exceed a **defined threshold**.

### ✅ **Key Mechanisms**

- **Failure Tracking:** Counts **errors per time window**, triggering when exceeded.
- **Automatic Resets:** Reopens after a **cooldown period**.
- **Neutral Wires:** Redirects traffic **to alternate processing paths**.
- **Event Logging:** Reports trips, resets, and failures **for diagnostics**.
- **Concurrency Safe:** Uses **atomic counters and locks** for efficiency.

---

## 🔧 Extending the Circuit Breaker Package

To **customize failure handling**, follow this **structured workflow**:

### 1️⃣ Modify `types/`

- Define new methods inside `types/circuitbreaker.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file contains **public methods** – update it accordingly.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **declarative configuration** via functional options.

### 4️⃣ Extend `notify.go` for logging & telemetry

- If new events are introduced, add **sensor and logger hooks**.

### 5️⃣ Unit Testing (`circuitbreaker_test.go`)

- **Ensure resets, failure handling, and neutral wire routing are tested**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Examples Directory](../../../example/circuit_breaker_example/)** – Demonstrates **Circuit Breakers in action**.

---

## 📝 License

The **Circuit Breaker package** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy fault-tolerant processing! ⚙️🛑** If you have questions or need support, feel free to open a GitHub issue.
