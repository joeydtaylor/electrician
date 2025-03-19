# 📜 Internal Logger Package

The **Internal Logger** package provides high-performance, structured logging for Electrician.  
Unlike most of Electrician, which exclusively relies on the **Go standard library**, this package uses **Zap**,  
the industry-standard structured logger from Uber, for **performance and efficiency**.

Zap offers **zero-allocation logging**, making it the fastest structured logger available for Go.

---

## 📦 Package Overview

| Feature                       | Description                                                         |
| ----------------------------- | ------------------------------------------------------------------- |
| **Blazing-Fast Logging**      | Uses **Zap** for **low-latency structured logging**.                |
| **Multiple Log Sinks**        | Supports **stdout, file-based, and network sinks** for flexibility. |
| **Structured JSON Logs**      | Provides **machine-readable** logs for easy ingestion.              |
| **Dynamic Log Level Control** | Adjust log levels **without restarting** the application.           |
| **Caller Tracing**            | Includes **caller depth tracking** to maintain accurate trace logs. |

---

## 📂 Package Structure

| File                       | Purpose                                                     |
| -------------------------- | ----------------------------------------------------------- |
| **api.go**                 | Public API for logging operations.                          |
| **internal.go**            | Low-level logic for managing **Zap logger configuration**.  |
| **notify.go**              | Handles **log event sinks** and structured log output.      |
| **options.go**             | Functional options for **log customization**.               |
| **internallogger.go**      | Core **Type Definition and Constructor**.                   |
| **internallogger_test.go** | Unit tests for **log performance, output, and formatting**. |

---

## 🔧 How Logging Works

Electrician’s **Internal Logger** provides high-speed logging **without sacrificing flexibility**.

### ✅ **Key Mechanisms**

- **Multiple Output Sinks:** Logs can be written to **stdout, files, or network-based sinks**.
- **Structured JSON Logging:** Ensures **logs are easily parsed and machine-readable**.
- **Dynamic Log Levels:** Supports **Debug, Info, Warn, Error, Fatal, Panic** dynamically.
- **Caller Depth Customization:** Ensures accurate **tracebacks even in deep call stacks**.

### ✅ **Lifecycle Management**

| Method                      | Description                                             |
| --------------------------- | ------------------------------------------------------- |
| `Debug() / Info() / Warn()` | Logs messages at different **severity levels**.         |
| `SetLevel()`                | Dynamically **adjusts logging level** at runtime.       |
| `AddSink()`                 | Adds a **new log output sink** (file, network, stdout). |
| `Flush()`                   | Ensures **all logs are written before shutdown**.       |

---

## 🔧 Extending the Internal Logger Package

To **add new logging functionality**, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- Define new **log levels, output sinks, or metadata options** inside `types/logger.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file contains **public API methods** – update it accordingly.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for new log events

- If new log types are introduced, add **event handlers for structured logging**.

### 5️⃣ Unit Testing (`internallogger_test.go`)

- **Ensure logging performance, output formatting, and error handling are verified**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/logging/)** – Demonstrates **real-world logging configurations**.

---

## 📝 License

The **Internal Logger package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy logging! 📝⚡** If you have questions or need support, feel free to open a GitHub issue.
