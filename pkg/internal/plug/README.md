# 🔌 Plug Package

A **Plug** is Electrician’s **ingestion bridge**.

It sits between one or more **Adapters** (things that fetch/produce data from “somewhere”) and downstream components like **Generators** and **Wires**.

Think of it as the place where you:

* connect an adapter implementation (HTTP, Kafka, file reader, custom source, …)
* optionally wrap it with adapter functions for shaping/validation
* attach telemetry hooks (sensor/logger) around ingestion

A plug is not a pipeline stage by itself. It’s the input side of a pipeline.

---

## ✅ What it does

| Capability           | Meaning                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------ |
| 🔌 Adapter wiring    | Attach one or more adapters that can produce `T`.                                          |
| 🧩 Adapter functions | Optional functional wrappers to normalize or post-process adapter output.                  |
| 📡 Telemetry hooks   | Emit events around ingestion (success/failure, timing, etc.) depending on implementation.  |
| 🧱 Stable contract   | Gives generators a single surface to pull from without caring which adapter is underneath. |

---

## 🧠 How it fits in a pipeline

A common composition is:

**Adapter → Plug → Generator → Wire**

* **Adapter** talks to the external system.
* **Plug** owns the adapter wiring and any “ingestion shaping” logic.
* **Generator** decides when/how often to pull from the plug.
* **Wire** handles concurrent transformation and output.

This separation keeps ingestion concerns (network calls, polling, auth, decoding) out of the hot processing path.

---

## 📂 Package structure

| File          | Purpose                              |
| ------------- | ------------------------------------ |
| `plug.go`     | Type definition + constructor        |
| `api.go`      | Public methods / component wiring    |
| `internal.go` | Adapter execution / ingestion logic  |
| `options.go`  | Functional options for configuration |
| `*_test.go`   | Tests                                |

If there is no `notify.go`, that’s intentional — telemetry hooks live next to the call sites.

---

## 🔧 Core responsibilities

### 1) Adapter management

Plugs typically support:

* connecting adapter instances (implementing the `types.Adapter[...]` contract)
* enumerating connected adapters (introspection/debugging)

### 2) Adapter function chaining

Adapter functions are an escape hatch for “do a small thing around ingestion” without building a whole new adapter type.

Typical uses:

* normalize payload shape
* inject defaults
* validate required fields
* translate adapter errors into structured errors

Keep adapter funcs small. If it grows teeth, make a real adapter.

### 3) Telemetry

Plugs can invoke sensors/loggers around ingestion events.

Important: sensors are typically invoked inline by components. If you attach a sensor that does heavy work, offload inside the sensor.

---

## ⚙️ Configuration contract

Plugs are configuration-time wiring:

* attach adapters and functions before the pipeline starts
* mutating plug configuration while running is not supported

This is the same contract as the rest of Electrician: predictable concurrency and minimal hot-path overhead.

---

## 🔧 Extending the Plug package

When you add capability, decide whether it’s a contract change or an implementation change:

* **Cross-component contract change** → update `types/plug.go` first.
* **Plug-only behavior** → implement inside `pkg/internal/plug`.
* **User-facing knob** → expose via `pkg/builder` (`PlugWith…`).

Add tests that cover:

* adapter execution behavior
* error propagation
* cancellation semantics
* (if applicable) telemetry hooks

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
