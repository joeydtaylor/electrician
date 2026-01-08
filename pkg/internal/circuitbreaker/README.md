# 🛑 Circuit Breaker Package

A **Circuit Breaker** is Electrician’s **admission gate** for protecting pipelines when failures start stacking.

It answers one question:

* ✅ **Should we accept work right now?** (`Allow()`)

And it exposes one primary feedback signal:

* ❌ **An error happened** (`RecordError()`)

When attached to a `Wire`, the breaker can also provide **neutral/ground wires** as an alternate destination for work when the circuit is open.

---

## ✅ What it does

| Capability           | Meaning                                                                                                |
| -------------------- | ------------------------------------------------------------------------------------------------------ |
| 🚦 Allow/deny        | `Allow()` decides whether the pipeline should accept new submissions.                                  |
| 📉 Error recording   | `RecordError()` feeds the breaker’s trip logic (threshold/window/cooldown are implementation-defined). |
| 🧯 Diversion targets | `GetNeutralWires()` can return alternate wires to receive items when the breaker denies.               |
| 🧵 Concurrency-safe  | `Allow()` and `RecordError()` must be safe under concurrent calls (hot path).                          |

---

## ❌ What it is not

* Not a retry policy (that’s the **insulator**).
* Not durable buffering.
* Not a distributed circuit breaker.
* Not a guarantee of “no overload” — it’s a local control mechanism.

---

## 🧠 How it works with Wire

When a breaker is attached to a `Wire`, the wire uses it in two places:

### 1) Submission gating

`Wire.Submit(...)` checks `cb.Allow()`.

* If allowed: the element proceeds into the wire’s input channel.
* If denied: the wire diverts the element to `cb.GetNeutralWires()` (if any), otherwise the element is dropped (best-effort) and telemetry is emitted.

### 2) Error feedback

When processing fails:

* the wire reports the error on an error channel (best-effort), and
* the breaker can be notified via `RecordError()` (exact call points depend on the failure path).

The breaker is therefore both:

* a **gate** for new work, and
* a **feedback sink** for failures.

---

## 🧯 Neutral / ground wires (diversion)

Neutral wires are simply other wires you provide to the breaker.

When the circuit is open, the wire will attempt to submit the element to each neutral wire:

* if at least one submission succeeds, the element is considered handled
* failures to submit to a neutral wire are ignored (best-effort)
* if no neutral wires exist, the element is dropped

Important: a “neutral wire” is not automatically a durable queue or a backup system. It behaves exactly like whatever wire you connected.

---

## 🧪 Interaction with Insulator (retry/recovery)

Insulator retries exist to recover from transform failures.

The wire’s behavior is intentionally conservative:

* If the breaker is open, the pipeline should be in **protect mode**, so retries may be skipped and work diverted/dropped instead.
* If retries are exhausted, the breaker can record the failure (`RecordError()`), which can trip the circuit depending on the breaker’s policy.

Bottom line: breaker = system safety gate; insulator = per-item recovery.

---

## ⚙️ Configuration contract

Circuit breakers are meant to be configured before `Start()`:

✅ Configure → Start → Run → Stop/Restart

Mutating breaker wiring (especially neutral wires) while the pipeline is running is not supported.

---

## 🔧 Extending the circuit breaker

* Cross-component contract changes start in `pkg/internal/types/circuitbreaker.go`.
* Implement behavior in `pkg/internal/circuitbreaker`.
* Expose user-facing knobs through `pkg/builder` (`CircuitBreakerWith…`).

Tests should cover:

* trip logic (threshold/cooldown)
* `Allow()` behavior under concurrency
* diversion correctness (neutral wires)
* interaction with error recording

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Examples Directory](../../../example/circuit_breaker_example/)** – Demonstrates **Circuit Breakers in action**.

---

## 📝 License

The **Circuit Breaker package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy fault-tolerant processing! ⚙️🛑** If you have questions or need support, feel free to open a GitHub issue.
