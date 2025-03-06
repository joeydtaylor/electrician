# 🛡️ Surge Protector Package

The **Surge Protector** is a **load-balancing and overload protection mechanism** for Electrician pipelines.  
It prevents **system failures** by controlling data flow under high load conditions, redistributing traffic to **backup systems**, and enforcing **rate limits**.

Surge Protectors are essential for **stabilizing throughput**, preventing **queue overflows**, and ensuring **graceful degradation** in high-demand environments.

---

## 📦 Package Overview

| Feature                       | Description                                                               |
| ----------------------------- | ------------------------------------------------------------------------- |
| **Rate Limiting**             | Limits the number of requests processed over time.                        |
| **Blackout Periods**          | Temporarily disables processing during predefined time windows.           |
| **Backup Systems**            | Redirects traffic to alternate pipelines during overload.                 |
| **Resister Integration**      | Uses a **priority queue** to buffer and manage excess traffic.            |
| **Event Logging & Telemetry** | Supports **real-time monitoring** via **sensors and structured logging**. |

---

## 📂 Package Structure

Each file follows **Electrician’s structured approach**, ensuring a **clear separation of concerns**.

| File                       | Purpose                                                                         |
| -------------------------- | ------------------------------------------------------------------------------- |
| **api.go**                 | Public API methods for managing Surge Protectors.                               |
| **internal.go**            | Low-level logic for managing surge conditions and backpressure.                 |
| **notify.go**              | Handles **event logging, sensor notifications, and telemetry hooks**.           |
| **options.go**             | Functional options for configuring Surge Protectors in a **composable manner**. |
| **surgeprotector.go**      | Core **Type Definition and Constructor**.                                       |
| **surgeprotector_test.go** | Unit tests ensuring **correctness, performance, and fault tolerance**.          |

---

## 🔧 How Surge Protectors Work

A **Surge Protector** acts as a **guardrail** for processing pipelines, ensuring that bursts of incoming data do not overwhelm the system.

### ✅ **Core Responsibilities**

- **Rate Limiting:** Prevents excessive processing load by enforcing token-based request control.
- **Backup Routing:** Automatically redirects excess traffic to **alternative processing units**.
- **Blackout Periods:** Allows scheduled downtime or controlled failure handling.
- **Resister Queue:** Implements a **priority-based buffering system** for deferred processing.

### ✅ **Lifecycle Management**

| Method           | Description                                                    |
| ---------------- | -------------------------------------------------------------- |
| `Trip()`         | Activates surge protection, halting normal data flow.          |
| `Reset()`        | Restores normal operation and restarts managed components.     |
| `IsTripped()`    | Checks if the Surge Protector is actively restricting traffic. |
| `AttachBackup()` | Connects **backup processing systems** for failover.           |

---

## 🔒 Standard Library First

Like most of Electrician, the **Surge Protector package is built entirely on Go’s standard library**.  
This ensures:

✅ **Maximum compatibility** – No unnecessary third-party dependencies.  
✅ **Minimal attack surface** – Secure and easy to audit.  
✅ **High performance** – Optimized for **low-latency, high-throughput processing**.

Electrician adheres to a **strict standard-library-first** philosophy, ensuring long-term maintainability.

---

## 🔧 Extending the Surge Protector Package

To **add new functionality** to the Surge Protector package, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- **Define the new interface method** inside `types/surgeprotector.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file inside the **surgeprotector** package must now implement this method.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Ensure `notify.go` handles event logging (if applicable)

- If your change introduces **new events**, add corresponding **logging and telemetry hooks**.

### 5️⃣ Unit Testing (`surgeprotector_test.go`)

- **Write tests** to verify that the new functionality works as expected.

By following these steps, Electrician maintains **stability, compatibility, and strict type safety**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/surge_protector_example/)** – Demonstrates how Surge Protectors work in real-world use cases.

---

## 📝 License

The **Surge Protector package** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
