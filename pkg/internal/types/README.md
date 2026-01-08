# ⚙️ Types Package

The `types` package is Electrician’s **contract layer**.

It defines the shared **interfaces**, **option types**, and core **data structures** that internal components implement and the public `pkg/builder` API composes.

If two packages need to agree on behavior or shape (Wire ↔ SurgeProtector, Relay ↔ Receiver, Logger ↔ Sensors, etc.), the contract lives here.

---

## ✅ What belongs in `types/`

* Component interfaces (Wire, CircuitBreaker, SurgeProtector, Sensor, Generator, Relay, Adapter, …)
* Cross-component structs (metadata, envelope types, element wrappers, error types)
* Shared enums/constants used across packages (log levels, component type strings, etc.)
* Functional option primitives (e.g., `Option[T]`) used by `builder` and internal constructors

## ❌ What doesn’t belong in `types/`

* Implementation logic
* Package-specific helpers that aren’t used outside that package
* Anything that’s experimental or not meant to be part of a shared contract

“Types is the source of truth” does **not** mean every change starts here.

* If a feature is internal to a single component, implement it there.
* If a feature changes how multiple components interact (or should be exposed through `builder`), define/extend the contract here.

---

## 📦 Package overview (by file)

This is the conceptual map. File names may evolve, but the responsibilities stay the same.

| File                    | What it defines                                         |
| ----------------------- | ------------------------------------------------------- |
| `adapter.go`            | Adapter interfaces for external data sources            |
| `circuitbreaker.go`     | Circuit breaker contract                                |
| `codec.go`              | Encoder/decoder interfaces                              |
| `common.go`             | Shared types/constants used across components           |
| `conduit.go`            | Conduit interface for multi-stage pipelines             |
| `element.go`            | Element wrappers used for routing/queues/error channels |
| `forwardrelay.go`       | Forward relay interfaces (outbound)                     |
| `generator.go`          | Generator interfaces                                    |
| `httpclient_adapter.go` | HTTP adapter-facing interface(s)                        |
| `logger.go`             | Logging interfaces + levels                             |
| `meter.go`              | Meter interfaces for measurement/monitoring             |
| `plug.go`               | Plug interfaces for pluggable inputs                    |
| `receiver.go`           | Receiver interface(s)                                   |
| `receivingrelay.go`     | Receiving relay interfaces (inbound)                    |
| `resister.go`           | Resister/queue contracts                                |
| `sensor.go`             | Sensor interfaces for hooks/telemetry                   |
| `submitter.go`          | Submission contracts for components that accept items   |
| `surgeprotector.go`     | Surge protector contract                                |
| `wave.go`               | Wave/advanced encoding contract (specialized)           |
| `wire.go`               | Wire interface (core pipeline primitive)                |

---

## 🔗 Relationship to `pkg/internal/*` and `pkg/builder`

* `pkg/internal/*` packages implement these interfaces and use these shared structs.
* `pkg/builder` is the public construction layer that:

  * returns values that satisfy `types` interfaces,
  * exposes functional options that ultimately call internal connect/set methods.

The point of `types/` is to keep components interoperable without tight coupling between implementations.

---

## 🔧 Extending Electrician (the accurate workflow)

When you add capability, decide first: **is this a shared contract change or an internal implementation change?**

### 1) If it changes cross-component behavior

1. Update/add the interface or shared struct in `types/`.
2. Implement it in the relevant `pkg/internal/*` package.
3. Expose it via `pkg/builder` (constructor arg and/or `XWith…` option).
4. Add tests (and allocation/benchmark tests if it touches hot paths).

### 2) If it’s internal to one component

1. Implement it inside that `pkg/internal/*` package.
2. Optionally expose it through `builder` if it’s a user-facing configuration knob.
3. Only touch `types/` if other packages must depend on the new contract.

---

## 📚 Dependency stance

`types/` is intentionally **standard-library only**.

Contracts should be easy to audit and stable to import; third-party libraries belong in implementations (internal packages / adapters), not in the contract surface.

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/)** – Demonstrates how these interfaces power Electrician's features.

---

## 📝 License

The **Types package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
