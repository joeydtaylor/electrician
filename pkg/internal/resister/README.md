# 🛠️ Resister Package

The **Resister** is Electrician’s **in-memory defer queue**.

It exists to absorb bursts when a pipeline is rate limited or temporarily tripped (typically by a **Surge Protector**) and to hand that work back to the hot path when capacity returns.

This package is deliberately boring:

* ✅ thread-safe enqueue/dequeue
* ✅ predictable draining behavior
* ✅ bounded policies (capacity / drop / retry limits) depending on implementation
* ❌ not durable storage
* ❌ not a broker
* ❌ not “guaranteed delivery”

---

## 🧠 Where Resister sits in the system

The typical topology is:

**Submit → SurgeProtector → (Resister queue) → Wire drain loop → Wire input → workers**

When a surge protector is attached to a wire:

* submissions that can’t be admitted immediately are routed into the protector’s handling path
* the protector may enqueue into a resister queue
* the wire runs a drain loop that pulls items back out later and submits them directly into the wire

Resister is the queue in the middle.

---

## ✅ What it does

| Capability           | Meaning                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------ |
| 🧱 Buffering         | Temporarily store items when immediate processing is not allowed.                          |
| 🔁 Deferred re-entry | Items are replayed into the pipeline when tokens/capacity allow.                           |
| 🧵 Concurrency-safe  | Safe for concurrent producers + a consumer drain loop.                                     |
| 🧯 Bounded behavior  | Policies like max depth / drop behavior / retry attempt limits are implementation-defined. |

What it does **not** do (by default):

* priority scheduling
* time-based decay
* persistence

If you need those, implement them explicitly and document the policy.

---

## 🔧 How draining works (paired with Surge Protector + Wire)

Electrician’s wire integration is intentionally simple:

1. A background loop wakes on a cadence (usually tied to the protector’s refill interval).
2. It attempts to acquire capacity (`TryTake`).
3. If capacity is available, it dequeues one item.
4. It submits that item **directly** into the wire (bypassing the surge protector) so queued work doesn’t get re-queued.

This avoids feedback loops and keeps the queue mechanism isolated.

---

## 📦 Contract expectations

The concrete interfaces live in `pkg/internal/types`.

At minimum, the resister/queue side needs to support the semantics the surge protector and wire depend on:

* enqueue a wrapped element
* dequeue one element (or return an error/empty)
* report current depth

If the system tracks retry attempts (e.g. a max attempt budget for queued items), that should be part of the element wrapper or queue policy.

---

## 🎛️ Design constraints

Resister exists to protect the hot path, so constraints are strict:

* ✅ enqueue should be fast and bounded
* ✅ dequeue should be non-blocking or cancellation-aware
* ✅ queue operations should not allocate per-op beyond what the caller already did
* ✅ the drain loop must stop cleanly on context cancellation

If you require durability, ordering guarantees across restarts, or cross-process coordination, you want a real system boundary (Kafka, SQS, NATS, etc.) — not an in-memory resister.

---

## 🔧 Extending the package

When changing queue behavior, be explicit about policy:

* ordering (FIFO/LIFO/priority)
* capacity and drop behavior
* retry attempt tracking (if any)
* cancellation semantics

If a change impacts cross-component behavior, update `types/resister.go` first, then implement it here, then expose it via `builder` if it’s a user-facing knob.

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/surge_protector_example/rate_limit/)** – Demonstrates how Resisters **prioritize, retry, and decay elements** in real-world use cases.

---

## 📝 License

The **Resister package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
