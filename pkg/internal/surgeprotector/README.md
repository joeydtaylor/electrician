# 🛡️ Surge Protector Package

The **Surge Protector** is Electrician’s **admission control + smoothing layer**.

It sits in front of a hot pipeline (typically a **Wire**) and prevents overload by:

* 🪙 enforcing **rate limits** (token/refill style)
* 🧱 optionally **buffering** excess work in a *resister queue*
* 🧯 providing a **tripped** mode where normal submission is bypassed and the protector takes over handling

It’s not “load balancing” in the cluster sense. It’s **backpressure control** at the edge of a pipeline.

---

## ✅ What it does

| Capability                   | What it means in practice                                                                                           |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| 🪙 Rate limiting             | Limit how quickly items enter the hot path (`TryTake` gating).                                                      |
| 🧱 Resister queue (optional) | When rate limited/tripped, queue items for later processing. Ordering is implementation-defined (FIFO vs priority). |
| 🚦 Trip / reset behavior     | A tripped protector forces submissions into the protector path rather than direct processing.                       |
| 📡 Telemetry hooks           | The wire can emit events when submissions are deferred/tripped (loggers/sensors).                                   |

---

## 🧠 How it works with Wire

When a surge protector is attached to a `Wire`, the submission path changes:

### 1) Normal case (not tripped, tokens available)

* `Wire.Submit` checks `sp.IsBeingRateLimited()` and calls `sp.TryTake()`.
* If a token is available, the element is submitted normally into the wire’s input channel.

### 2) Rate limited (no token available)

* `Wire.Submit` emits a rate-limit notification (best-effort telemetry).
* The element is wrapped and handed to `sp.Submit(...)` for protector-side handling (typically queueing).

### 3) Tripped

* The element is routed directly to `sp.Submit(...)`.

### 4) Resister draining loop

If the protector reports that a resister queue is connected (`IsResisterConnected()`), the wire starts a background drain loop at `Start()`.

That loop:

* wakes on a cadence derived from the protector’s refill interval (or a conservative fallback)
* tries to take a token (`TryTake()`)
* dequeues one item (`Dequeue()`)
* submits it **directly** into the wire (bypassing the surge protector) so queued work doesn’t get re-queued

The loop stops when the wire context is cancelled or the queue is empty.

---

## 🔧 Contract expectations (what Wire uses)

A surge protector implementation is expected to provide methods with semantics equivalent to:

* `IsTripped()`
* `Trip()` / `Reset()` (if supported by the implementation)
* `IsBeingRateLimited()`
* `TryTake()` (non-blocking token acquisition)
* `GetTimeUntilNextRefill()` (used for “next attempt” telemetry)
* `GetRateLimit()` (includes refill interval and retry policy tuning)
* `Submit(ctx, element)` (protector-side handling: queue/drop/redirect depending on implementation)
* `IsResisterConnected()`, `Dequeue()`, `GetResisterQueue()`
* `ConnectComponent(...)` (wires itself into the component graph)

The exact type names live in `pkg/internal/types`.

---

## 🎛️ Design constraints

Surge protection exists to protect the hot path, so the design constraints are strict:

* ✅ submission should not block the pipeline
* ✅ protector handling should be bounded (queue depth + policy)
* ✅ telemetry must be best-effort (never stall processing)

If you need “hard” delivery guarantees, enforce them at the system boundary (durable queues, retry policies, idempotency) and treat the surge protector as a local control mechanism.

---

## 🔧 Extending the package

* If you need new cross-component behavior, update `types/surgeprotector.go` first.
* Then implement it in `pkg/internal/surgeprotector`.
* Add builder exposure via `SurgeProtectorWith…` / `WireWithSurgeProtector` as appropriate.
* Add tests that cover:

  * rate limit behavior
  * queueing/draining
  * trip/reset semantics
  * cancellation behavior

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/surge_protector_example/)** – Demonstrates how Surge Protectors work in real-world use cases.

---

## 📝 License

The **Surge Protector package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
