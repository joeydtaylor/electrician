# 📊 Meter Package

The **Meter** package provides **real-time performance monitoring** and **metric tracking** for Electrician pipelines.  
It collects, evaluates, and reports various **processing metrics** to ensure system **efficiency and reliability**.

---

## 📦 Package Overview

| Feature                      | Description                                                            |
| ---------------------------- | ---------------------------------------------------------------------- |
| **Real-Time Monitoring**     | Tracks **throughput, error rates, and processing efficiency**.         |
| **Threshold-Based Alerts**   | Triggers actions when **metrics exceed defined limits**.               |
| **Idle Timeout Detection**   | Automatically stops processing if the system becomes idle.             |
| **Event Logging & Sensors**  | Supports **structured logging and real-time telemetry hooks**.         |
| **Performance Optimization** | Measures **CPU, RAM, and Go routine usage** for fine-tuning pipelines. |

---

## 📂 Package Structure

| File              | Purpose                                                              |
| ----------------- | -------------------------------------------------------------------- |
| **api.go**        | Public API methods for **configuring meters and accessing metrics**. |
| **internal.go**   | Low-level logic for **performance tracking and metric evaluation**.  |
| **notify.go**     | Handles **event logging, telemetry, and alert notifications**.       |
| **options.go**    | Functional options for **customizing metric tracking behavior**.     |
| **meter.go**      | Core **Type Definition and Constructor**.                            |
| **meter_test.go** | Unit tests ensuring **accuracy and stability of metric tracking**.   |

---

## 🔧 How Meters Work

A **Meter** continuously monitors **pipeline performance**, tracking various metrics such as:

- **Throughput:** Number of processed elements per second.
- **Error Rates:** Percentage of failed transformations.
- **Resource Utilization:** CPU, memory, and active goroutines.
- **Pipeline Health:** Overall progress and remaining workload.

### ✅ **Key Mechanisms**

- **Threshold Monitoring:** Stops execution if errors exceed a **configurable limit**.
- **Idle Detection:** Auto-terminates after a **specified idle duration**.
- **Live Metrics Reporting:** Continuously updates and prints system statistics.

### ✅ **Lifecycle Management**

| Method                 | Description                                                   |
| ---------------------- | ------------------------------------------------------------- |
| `IncrementCount()`     | Increases a specific **metric counter**.                      |
| `SetMetricThreshold()` | Defines **alert thresholds** for monitored metrics.           |
| `Monitor()`            | Starts **real-time tracking** of pipeline performance.        |
| `GetMetricCount()`     | Retrieves the **current value** of a specific metric.         |
| `SetIdleTimeout()`     | **Terminates processing** if no activity occurs for a period. |

---

## 🔧 Extending the Meter Package

To **add new metrics or functionality**, follow this structured **workflow**:

### 1️⃣ Modify `types/`

- Define new **metrics and counters** inside `types/meter.go`.
- This ensures **all implementations remain consistent**.

### 2️⃣ Implement in `api.go`

- The `api.go` file contains **public API methods** – update it accordingly.

### 3️⃣ Add a Functional Option in `options.go`

- Supports **composable, declarative-style configuration**.

### 4️⃣ Extend `notify.go` for logging & telemetry

- If new events are introduced, add **sensor and logger hooks**.

### 5️⃣ Unit Testing (`meter_test.go`)

- **Ensure new metrics are correctly tracked and reported**.

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/meter_example/)** – Demonstrates **real-world Meter usage**.

---

## 📝 License

The **Meter package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy monitoring! 📈🚀** If you have questions or need support, feel free to open a GitHub issue.
