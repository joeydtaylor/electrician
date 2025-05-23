# 📊 Meter Package

The **Meter** package provides **real-time performance monitoring** and **metric tracking** for Electrician pipelines.  
It collects, evaluates, and reports various **processing metrics** to ensure system **efficiency and reliability**.

---

## 📦 Package Overview

| Feature                      | Description                                                            |
| ---------------------------- | ---------------------------------------------------------------------- |
| **Real-Time Monitoring**     | Tracks **throughput, error rates, and processing efficiency**.         |
| **Idle Timeout Detection**   | Automatically stops processing if the system becomes idle.             |
| **Event Logging & Sensors**  | Supports **structured logging and real-time telemetry hooks**.         |
| **Performance Optimization** | Measures **CPU, RAM, and Go routine usage** for fine-tuning pipelines.   |

---

## 📂 Package Structure

| File              | Purpose                                                              |
| ----------------- | -------------------------------------------------------------------- |
| **api.go**        | Public API methods for **configuring meters and accessing metrics**. |
| **internal.go**   | Low-level logic for **performance tracking and metric evaluation**.  |
| **notify.go**     | Handles **event logging, telemetry, and alert notifications**.       |
| **options.go**    | Functional options for **customizing metric tracking behavior**.     |
| **meter.go**      | Core **Type Definition and Constructor**.                            |
| **meter_test.go** | Unit tests ensuring **accuracy and stability of metric tracking**.    |

---

## 🔧 How Meters Work

A **Meter** continuously monitors **pipeline performance**, tracking various metrics such as:

- **Throughput:** Number of processed elements per second.  
- **Error Rates:** Percentage of failed transformations.  
- **Resource Utilization:** CPU, memory, and active goroutines.  
- **Pipeline Health:** Overall progress and remaining workload.

### ✅ Key Mechanisms

- **Idle Detection:** Auto-terminates after a **specified idle duration**.  
- **Live Metrics Reporting:** Continuously updates and prints system statistics.

### ✅ Lifecycle Management

| Method             | Description                                                   |
| ------------------ | ------------------------------------------------------------- |
| `IncrementCount()` | Increases a specific **metric counter**.                      |
| `Monitor()`        | Starts **real-time tracking** of pipeline performance.        |
| `GetMetricCount()` | Retrieves the **current value** of a specific metric.         |
| `SetIdleTimeout()` | **Terminates processing** if no activity occurs for a period.  |

---

## 📑 Built-In Metrics Reference

Below is a list of **string constants** used when monitoring. Pass these string values (e.g., `"error_count"`) into your Meter methods or functional options to monitor each metric.

### System Usage Metrics

- **`"current_cpu_percentage"`**  
- **`"peak_cpu_percentage"`**  
- **`"current_ram_percentage"`**  
- **`"peak_ram_percentage"`**  
- **`"current_go_routines_active"`**  
- **`"peak_go_routines_active"`**  

### Throughput & Performance

- **`"processed_per_second"`**  
- **`"transforms_per_second"`**  
- **`"max_processed_per_second"`**  
- **`"max_transformed_per_second"`**  
- **`"errors_per_second"`**  
- **`"max_transform_errors_per_second"`**  

### Counts & Totals

- **`"element_submitted_total_count"`**  
- **`"element_processed_total_count"`**  
- **`"element_transformed_total_count"`**  
- **`"element_pending_total_count"`**  
- **`"total_error_count"`**  
- **`"progress_percentage"`**  

### Transformation & Error Tracking

- **`"transform_percentage"`**, **`"error_percentage"`**  
- **`"element_transform_count"`**  
- **`"element_transform_error_count"`**  
- **`"element_error_count"`**  
- **`"element_retry_count"`**, **`"element_recover_success_count"`**, **`"element_recover_failure_count"`**  

### Circuit Breaker Metrics

- **`"circuit_breaker_trip_count"`**, **`"circuit_breaker_reset_count"`**  
- **`"circuit_breaker_current_trip_count"`**, **`"circuit_breaker_recorded_error_count"`**  
- **`"circuit_breaker_diverted_element_count"`**  
- **`"circuit_breaker_dropped_element_count"`**  
- **`"circuit_breaker_last_trip_time"`**, **`"circuit_breaker_next_reset_time"`**  
- **`"circuit_breaker_count"`**  

### Surge Protector (High-Load) Metrics

- **`"surge_attached_count"`**, **`"surge_current_trip_count"`**  
- **`"surge_trip_count"`**, **`"surge_reset_count"`**  
- **`"surge_rate_limit_exceed_count"`**, **`"surge_drop_count"`**  
- **`"surge_backup_wire_submission_count"`**, **`"surge_backup_wire_submission_percentage"`**  

### HTTP & Networking Metrics

- **`"http_request_made_count"`**, **`"http_request_received_count"`**  
- **`"http_client_error_count"`**, **`"http_client_retry_count"`**  
- **`"http_response_error_count"`**, **`"http_client_json_unmarshal_error_count"`**  
- **`"http_client_fetch_successful_count"`**  

### Resister & Relay Metrics

- **`"resister_element_queued_count"`**, **`"resister_element_dequeued"`**  
- **`"resister_connected_count"`, `"resister_cleared_count"`**  
- **`"receiving_relay_received_count"`, `"receiving_relay_relayed_count"`**  
- **`"forward_relay_submit_count"`, `"forward_relay_error_count"`**  
- **`"forward_relay_relayed_count"`, `"forward_relay_payload_compression_count"`**

### Miscellaneous

- **`"process_duration"`**  
- **`"generator_running_count"`, `"generator_submit_count"`**  
- **`"conduit_running_count"`, `"wires_running_count"`**  
- **`"meter_connected_component_count"`, `"logger_connected_component_count"`**  
- **`"component_lifecycle_error_count"`, `"component_restart_count"`**  
- **`"error_count"`**

---

## 🔧 Extending the Meter Package

To add new metrics or functionality, follow this structured workflow:

1. Modify `types/` to define new metrics and counters in `types/meter.go`.  
2. Implement or update public API methods in `api.go` for your new metric(s).  
3. Add a Functional Option in `options.go` for clean, composable configuration.  
4. Extend `notify.go` if you introduce new events or sensor/logger hooks.  
5. Write or update unit tests in `meter_test.go` to validate your new metric logic.

---

## 📖 Further Reading

- [Root README](../../../README.md) – Electrician’s overall architecture and principles.  
- [Internal README](../README.MD) – How internal packages interact with `types/`.  
- [Examples Directory](../../../example/meter_example/) – Demonstrates real-world Meter usage.

---

## 📝 License

The **Meter package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

Happy monitoring! If you have questions or need support, feel free to open a GitHub issue.
