# ⚡ Wire Package

The `wire` package is Electrician’s **in-process pipeline primitive**: a generic, concurrent stage that ingests items, transforms them, and emits results.

It’s built for **high-throughput** workflows with a **configuration-first** contract:

* ✅ Configure once (connect components / set options)
* ✅ `Start()` to begin processing
* ✅ `Stop()` / `Restart()` to end or rehydrate
* ❌ Mutating configuration while running is **not supported** (race risk by design)

---

## 🧠 Mental Model

A **Wire** is a bounded, concurrent pipeline:

1. **Submit** pushes items into an internal input channel.
2. A worker pool pulls from the input channel.
3. Each item runs through **transform(s)** (or a per-worker factory transformer).
4. Results go to:

   * the **output channel**, and optionally
   * an **output buffer** via an **encoder**.

Optional components can change behavior:

* 🧯 **Circuit breaker**: reject/divert items when open
* 🛡️ **Surge protector**: rate limit + queue (“resister”) behavior
* 🧪 **Insulator**: retry/recovery policy on transform errors
* 📡 **Sensors + loggers**: structured telemetry hooks

---

## 📦 What You Get

| Capability             | What it does                                                     |
| ---------------------- | ---------------------------------------------------------------- |
| ⚙️ Concurrency control | Configurable buffer size + worker count                          |
| 🔁 Transform chain     | Ordered transforms, or a per-worker transformer factory          |
| 🧯 Circuit breaker     | Trip-aware submit and optional diversion to neutral/ground wires |
| 🛡️ Surge protector    | Rate limiting + queued retry path (resister processing loop)     |
| 🧪 Insulator           | Retry-on-error with threshold + interval, cancellation-aware     |
| 📤 Output channels     | Stream processed items to `OutputChan`                           |
| 🧾 Optional encoding   | Encode processed items into an output `bytes.Buffer`             |
| 📡 Telemetry           | Logger + sensor hooks for lifecycle + processing events          |

---

## 🗂️ Package Layout

This package is intentionally small and “stdlib-heavy”. telemetry hooks live alongside internal processing.

| File           | Purpose                                                                 |
| -------------- | ----------------------------------------------------------------------- |
| `wire.go`      | Core `Wire[T]` implementation + constructor                             |
| `api.go`       | Public API: connect components, lifecycle, submit/load helpers          |
| `internal.go`  | Processing engine, fast-path, retry/insulator, telemetry, resister loop |
| `options.go`   | Functional options (`With*`) for ergonomic construction                 |
| `wire_test.go` | Correctness + allocation-regression tests                               |

---

## 🚀 Quick Start

```go
ctx := context.Background()

w := wire.NewWire[int](
    ctx,
    wire.WithConcurrencyControl[int](1024, 4),
    wire.WithTransformer[int](func(v int) (int, error) {
        return v * 2, nil
    }),
)

_ = w.Start(ctx)
_ = w.Submit(ctx, 21)

out := <-w.GetOutputChannel() // 42
_ = w.Stop()
```

---

## 🔌 Configuration (Connect / Options)

### ✅ Preferred: functional options at construction

```go
w := wire.NewWire[Event](
    ctx,
    wire.WithComponentMetadata[Event]("ingest-wire", "wire-1"),
    wire.WithConcurrencyControl[Event](65536, 12),
    wire.WithInsulator[Event](retryFn, 3, 50*time.Millisecond),
    wire.WithSurgeProtector[Event](protector),
    wire.WithBreaker[Event](breaker),
    wire.WithLogger[Event](logger),
    wire.WithSensor[Event](sensor),
)
```

### ✅ Also supported: direct connect methods

* `ConnectCircuitBreaker(cb)`
* `ConnectSurgeProtector(sp)`
* `ConnectGenerator(g...)`
* `ConnectTransformer(t...)`
* `ConnectLogger(l...)`
* `ConnectSensor(s...)`

**Contract:** treat these as **configuration-time** calls. If you mutate while running, you’re outside the supported model.

---

## 🔁 Transformers

### 1) Ordered transform chain

```go
w.ConnectTransformer(t1, t2, t3)
```

* Order matters.
* `ConnectTransformer` panics if called **after** `Start()`.

### 2) Per-worker transformer factory

Use this when a transformer needs **worker-local state** without sync/pools.

```go
wire.WithTransformerFactory[T](func() types.Transformer[T] {
    // allocated once per worker
    scratch := make([]byte, 4096)
    return func(v T) (T, error) {
        // use scratch
        return v, nil
    }
})
```

* Factory and explicit transformers are **mutually exclusive**.
* Factory is evaluated once per worker.

### 🧰 `WithScratchBytes`

Convenience wrapper for “allocate once per worker” patterns:

```go
w := wire.NewWire[Event](
    ctx,
    wire.WithScratchBytes[Event](1024, func(buf []byte, e Event) (Event, error) {
        // buf reused per worker
        return e, nil
    }),
)
```

---

## ⚡ Fast Path (low overhead)

Wire has a **fast submit/processing path** when all “extras” are off:

* no circuit breaker
* no surge protector
* no insulator
* no encoder
* no loggers/sensors
* and there is exactly **one** transform source:

  * either `len(transformations)==1`, or
  * a `TransformerFactory` providing one transformer per worker

This is how you get the tightest inner loop and best shot at “effectively zero alloc” processing.

> ⚠️ Reality check: your transformer code can still allocate. Wire won’t save you from that.

---

## 🧯 Circuit Breaker

When attached, `Submit` checks `cb.Allow()`.

* If disallowed, the element is **diverted** to the breaker’s **neutral/ground wires** (if any).
* If no neutral wires exist, the element is dropped (best-effort) with telemetry.

There’s also a small ticker loop that can publish breaker state to an internal control channel for polling/observers.

---

## 🛡️ Surge Protector (rate limit + queue)

When a surge protector is attached:

* Submissions may be **rate limited**.
* If rate limited or tripped, elements can be queued (resister queue).
* A background loop drains the queue on a cadence based on the protector’s refill interval.

Design intent: **never stall the hot pipeline** because of telemetry or a congested queue.

---

## 🧪 Insulator (retry / recovery)

When configured, transform errors can trigger retries:

* threshold: max attempts
* interval: optional wait between attempts

The retry loop is:

* cancellation-aware (wire ctx)
* timer-based (avoids per-attempt `time.After` allocations)

If retries exhaust, the circuit breaker (if present) can record an error.

---

## 📤 Output + Loading

### Output channel

* Processed items are sent to `OutputChan`.
* `Stop()` closes `OutputChan` (best-effort, once).

### Optional encoded buffer

If an encoder is set, the processed element is encoded into `OutputBuffer` under a mutex.

### Load helpers

* `Load()` stops the wire (best-effort) and returns a **copy** of the current output buffer.
* `LoadAsJSONArray()` stops the wire (best-effort) and **drains whatever is currently available** from `OutputChan` into a JSON array.

  * It does **not** block waiting for future output.
  * Intended for debug/testing, not high-frequency hot paths.

---

## 🧬 Lifecycle

* `Start(ctx)`

  * starts error handling
  * starts resister processing (if applicable)
  * injects an **identity transform** if no transforms are configured (pass-through wire)
  * starts workers
  * starts generators (if connected)

* `Stop()`

  * cancels the wire context
  * waits for goroutines
  * closes output/error channels (best-effort)

* `Restart(ctx)`

  * `Stop()` + rehydrate context, channels, buffer, termination guards
  * restarts breaker ticker if breaker exists

---

## 📏 Performance Notes

* ✅ No per-item allocations in the wire core under normal conditions.
* ✅ In-place compaction is used when connecting components (drops nils without allocating).
* ✅ `WithScratchBytes` supports allocation-free worker-local scratch.
* ✅ Telemetry paths are best-effort and try hard not to backpressure the pipeline.

If you want to keep allocations low:

* Avoid capturing large structs in closures.
* Prefer worker-local scratch (factory / `WithScratchBytes`) over `sync.Pool` unless you have a measured reason.
* Keep telemetry off in hot paths unless you need it.

---

## 🧩 Extending Wire

1. Update interfaces in `pkg/internal/types` as needed.
2. Implement the behavior on `Wire[T]`.
3. Add a `WithX` functional option if the feature is configuration-time.
4. Add/extend tests (including allocation regression tests where relevant).

---

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Per-Package README]** – Each internal sub-package (`wire`, `circuitbreaker`, `httpserver`, etc.) contains additional details.
- **[Examples Directory](../../../example/)** – Demonstrates these internal components in real-world use cases.

---

## 📝 License

All internal packages fall under Electrician’s [Apache 2.0 License](../../../LICENSE).  
You are free to use, modify, and distribute them under these terms, but note that **they are not designed for direct external use**.

---

## ⚡ Happy wiring

Build pipelines that are boring under load.
