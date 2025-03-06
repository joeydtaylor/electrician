# 📡 Sensor Package – Event Monitoring & Telemetry

The **Sensor** package in Electrician provides a **flexible event monitoring system** that captures structured data from various **pipeline components**, including **Wires, Circuit Breakers, Surge Protectors, HTTP Clients, and more**.

It enables **real-time observability** by allowing developers to register **event-driven callbacks**, push structured logs, and update performance metrics via **attached Meters**.

---

## 📦 Package Overview

| Feature                        | Description                                                              |
| ------------------------------ | ------------------------------------------------------------------------ |
| **Event-Driven Callbacks**     | Hooks into key lifecycle events (`OnStart`, `OnStop`, `OnError`, etc.).  |
| **Multi-Component Monitoring** | Observes activity across **Wires, Circuit Breakers, HTTP Clients**, etc. |
| **Logging & Metrics**          | Supports **structured logging** and updates **attached meters**.         |
| **Modular & Extensible**       | Easily extendable to track new event types.                              |
| **Performance Tracking**       | Monitors **latency, error rates, and processing efficiency**.            |

---

## 📂 Package Structure

Each file follows **Electrician’s structured approach**, ensuring a **clear separation of concerns**.

| File               | Purpose                                                                    |
| ------------------ | -------------------------------------------------------------------------- |
| **api.go**         | Public API for **registering event listeners and invoking callbacks**.     |
| **internal.go**    | Internal logic for **event handling, logging, and metric updates**.        |
| **notify.go**      | Handles **event logging, structured telemetry, and sensor notifications**. |
| **options.go**     | Functional options for configuring Sensors **in a declarative manner**.    |
| **sensor.go**      | Core **Type Definition and Constructor**.                                  |
| **sensor_test.go** | Unit tests ensuring **correctness and performance**.                       |

---

## 🔧 How Sensors Work

Sensors act as **observability agents** within Electrician's event-driven architecture.  
They listen for **lifecycle events**, log structured data, and invoke **registered callbacks** for real-time monitoring.

### ✅ **Event Monitoring**

- **Lifecycle Events** – `OnStart`, `OnStop`, `OnRestart`, `OnComplete`, etc.
- **Error Handling** – `OnError` captures failures across the pipeline.
- **Data Processing Events** – `OnElementProcessed`, `OnSubmit`, `OnCancel`, etc.
- **Circuit Breaker Events** – Tracks trips, resets, and dropped elements.
- **HTTP Client Events** – Logs request lifecycle (`OnRequestStart`, `OnResponseReceived`, etc.).
- **Surge Protector Events** – Monitors rate limits, token releases, and backup activations.
- **Resister Events** – Tracks elements being queued, dequeued, and requeued.

### ✅ **Metrics & Logging**

- Sensors can **attach meters** to track event counts and performance statistics.
- **Structured logging** enables detailed **tracing** and **debugging** of system events.

---

## 🔒 Standard Library First

Like most of Electrician, the **Sensor package is built entirely on Go’s standard library**, ensuring:

✅ **Maximum compatibility** – No unnecessary third-party dependencies.  
✅ **Minimal attack surface** – Secure and easy to audit.  
✅ **High performance** – Optimized for **low-latency, high-throughput monitoring**.

Electrician adheres to a **strict standard-library-first** philosophy, ensuring long-term maintainability.

---

## 🔧 Extending the Sensor Package

To **add new functionality** to the Sensor package, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- **Define the new event method** inside `types/sensor.go`.
- This ensures **all components remain consistent**.

### 2️⃣ Implement the logic in `api.go`

- The `api.go` file inside the **sensor** package must now implement this event.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **declarative event registration**.

### 4️⃣ Ensure `notify.go` handles event logging (if applicable)

- If your change introduces **new events**, add corresponding **logging and telemetry hooks**.

### 5️⃣ Unit Testing (`sensor_test.go`)

- **Write tests** to verify that the new event hooks function correctly.

By following these steps, Electrician maintains **stability, compatibility, and strict type safety**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../../internal/README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/sensor_example/)** – Demonstrates how Sensors integrate with Electrician’s event system.

---

## 📝 License

The **Sensor package** is part of Electrician and is released under the [MIT License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
