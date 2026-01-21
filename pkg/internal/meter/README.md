# 📊 Meter Package

A **Meter** is Electrician’s **metrics accumulator + reporter**.

It’s a lightweight component that tracks counts/rates/progress for a running pipeline and (optionally) emits periodic snapshots for observability.

Meters don’t magically “monitor everything.” They track what you **increment** (usually via a `Sensor`) and what they can **sample** (runtime/host stats depending on configuration and implementation).

---

## ✅ What a Meter does

| Capability                   | Meaning                                                                                                   |
| ---------------------------- | --------------------------------------------------------------------------------------------------------- |
| 🔢 Counters                  | Track totals like submitted/processed/errors, retry outcomes, etc.                                        |
| ⏱️ Rates                     | Derive per-second rates from counters over time.                                                          |
| 📈 Progress                  | If you provide an expected total, progress can be computed (e.g., `processed / total`).                   |
| 🧵 Runtime sampling          | Optionally sample Go runtime stats (goroutines, heap, GC) at a cadence.                                   |
| 🖥️ Host sampling (optional) | Some implementations may sample host-level stats (CPU/RAM) using external libs (e.g., `gopsutil`).        |
| 📣 Reporting                 | `Monitor()`/ticker-driven reporting loops (logging, callbacks, or structured events depending on wiring). |

---

## 🧠 How it fits in the system

Meters are usually updated by a **Sensor**:

* components emit events (`OnSubmit`, `OnError`, `OnElementProcessed`, …)
* a sensor receives those events
* the sensor increments meter counters / updates gauges

Typical wiring:

**Wire/CB/SP/etc → Sensor → Meter**

This keeps hot-path code free of “metrics logic” and makes the policy (what you measure) explicit.

---

## 📂 Package structure

| File          | Purpose                                               |
| ------------- | ----------------------------------------------------- |
| `meter.go`    | Type definition + constructor                         |
| `api.go`      | Public methods (increment/read/report controls)       |
| `options.go`  | Functional options (`With*`)                          |
| `internal.go` | Sampling / evaluation helpers (implementation detail) |
| `*_test.go`   | Tests                                                 |

If there is no `notify.go`, that’s intentional — instrumentation should live at the emission points, not as a separate “telemetry file” by default.

---

## ⚙️ Practical semantics (no false promises)

### “Real-time monitoring”

A meter can *report periodically*, but it doesn’t control your pipeline. If you want a meter to trigger cancellation or restarts, you must wire that behavior explicitly (e.g., cancel a context when a threshold is crossed).

### “Idle timeout”

If an idle timeout exists, treat it as **a detection/signal** (“no events for N seconds”), not a guarantee that the pipeline will be stopped. Stopping work is a policy decision (and belongs in the owning component/context).

---

## 📑 Metric keys / names

Some meter implementations expose named metrics via string keys.

Do not hard-code guesses in documentation. Treat the metric keys as part of the implementation surface and look for:

* constants/enums in the meter package, or
* public accessor methods that return supported keys

If you need stable metric names across versions, define them in one place and write tests for them.

---

## 🚄 Performance stance

Meters should be cheap:

* keep increments fast (ideally atomic counters)
* do heavier computation/reporting on a ticker, not on every event
* never let reporting backpressure the pipeline

If your meter does expensive work (string formatting, IO, host sampling), isolate that work in the monitor loop.

---

## 🔧 Extending the Meter package

When adding new metrics:

1. Decide whether it’s a shared contract change (update `types/meter.go`) or implementation-only.
2. Implement the counter/gauge/rate logic.
3. Ensure sensors increment it at the correct event points.
4. Add tests that validate:

   * increments are correct under concurrency
   * reporting logic is stable
   * no accidental allocations/regressions on hot paths

## 📖 Further Reading

- [Root README](../../../README.md) – Electrician’s overall architecture and principles.  
- [Internal README](../README.MD) – How internal packages interact with `types/`.  
- [Examples Directory](../../../example/meter_example/) – Demonstrates real-world Meter usage.

---

## 📝 License

The **Meter package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

Happy monitoring! If you have questions or need support, feel free to open a GitHub issue.
