# ⚡ Generator Package

A **Generator** is Electrician’s **source driver**.

It produces values of type `T` and submits them into downstream components (most commonly a **Wire**). In practice, generators are the glue between **ingestion** (Plugs/Adapters) and **processing** (Wires).

Generators are optional: you can always call `Wire.Submit(...)` yourself. Use a generator when you want a reusable “pull/produce loop” with consistent wiring and telemetry.

---

## 🧠 Where Generator fits

Common composition:

**Adapter → Plug → Generator → Wire**

* **Adapter** talks to the outside world (HTTP/Kafka/S3/custom).
* **Plug** owns adapter wiring + optional ingestion shaping.
* **Generator** decides *when* to pull/produce and pushes results downstream.
* **Wire** does the concurrent hot-path work (transform → emit).

Generators should do minimal work. Heavy lifting belongs in the Wire.

---

## ✅ What a Generator does

| Capability                           | Meaning                                                                                             |
| ------------------------------------ | --------------------------------------------------------------------------------------------------- |
| 🧵 Produce loop                      | Runs a controlled loop (often tick/poll) that creates or fetches `T`.                               |
| 🔌 Plug integration                  | Pull from a configured Plug/Adapter rather than embedding IO logic in the pipeline.                 |
| 🧯 Circuit breaker gating (optional) | Avoid hammering upstreams when errors occur (policy depends on breaker + generator implementation). |
| 📡 Telemetry hooks                   | Emit events for submits/errors/starts/stops through sensors/loggers.                                |
| 🧬 Downstream wiring                 | Connect to one or more downstream submitters (typically via `ConnectToComponent`).                  |

What it does **not** guarantee:

* durability
* exactly-once delivery
* parallel ingestion by default

If you need durable replay or broker semantics, put a real system boundary upstream (Kafka/SQS/etc.) and treat the generator as a local driver.

---

## ⚙️ Lifecycle model

Generators follow the same operational contract as the rest of Electrician:

✅ Configure → Start → Run → Stop/Restart

* Attach plug/sensor/breaker during configuration.
* `Start()` begins the produce loop.
* `Stop()` cancels and drains/shuts down (best-effort).
* `Restart()` is a stop + rehydrate cycle; it’s not a live reconfiguration mechanism.

Mutating generator configuration while running is not supported.

---

## 🔧 How generation typically works

A common pattern inside a generator is:

1. Wait for a tick / trigger (or run continuously under control).
2. Pull or produce an element `T` (often via a Plug).
3. If a circuit breaker is attached and disallows work, skip or delay (implementation-defined).
4. Submit into the downstream component.
5. Emit telemetry (best-effort).

Backpressure is usually handled by the downstream submitter (e.g., a wire’s input channel may block when full). A generator should respect contexts and avoid unbounded goroutine growth.

---

## 📂 Package structure

| File           | Purpose                              |
| -------------- | ------------------------------------ |
| `generator.go` | Type definition + constructor        |
| `api.go`       | Public methods / wiring              |
| `internal.go`  | Produce loop implementation          |
| `options.go`   | Functional options for configuration |
| `*_test.go`    | Tests                                |

---

## 🔧 Extending Generator

When adding capability:

* Cross-component contract change → update `types/generator.go` first.
* Generator-only behavior → implement in `pkg/internal/generator`.
* User-facing knob → expose via `pkg/builder` (`GeneratorWith…`).

Tests should cover:

* start/stop behavior
* context cancellation
* breaker gating behavior (if applicable)
* correct submission into downstream components
* no accidental hot-path allocations/regressions

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.md)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/generator_example/)** – Demonstrates real-world **Generator usage**.

---

## 📝 License

The **Generator package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy generating! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
