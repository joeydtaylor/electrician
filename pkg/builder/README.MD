# 🏗️ Builder Package

The `builder` package is Electrician’s **public construction layer**.

It provides ergonomic constructors (`NewWire`, `NewConduit`, `NewRelay`, etc.) plus a consistent set of **functional options** (`WireWith…`, `ConduitWith…`, …) so you can assemble pipelines without reaching into implementation details.

Under the hood, `builder` wraps the engine in `pkg/internal/*`.

---

## 🔒 Import boundary

Electrician uses a “public surface + private engine” layout:

* `pkg/builder` is intended for consumers of the module.
* `pkg/internal/*` is private to `pkg/*` (enforced by Go’s `internal/` rule).

Practically:

* ✅ packages under `pkg/` can import `pkg/internal/...`
* ❌ external consumers cannot import `pkg/internal/...`

So `builder` is where the stable, user-facing API lives.

---

## 📦 What builder does

* 🧱 **Constructs components** with sane defaults.
* 🧩 **Applies options** consistently (configuration-time wiring).
* 🔁 **Keeps internal packages swappable** (internal implementations can evolve without forcing API churn).

`builder` is intentionally thin: it should not add runtime overhead to the hot path.

---

## ⚙️ Functional options

Electrician’s public API uses the functional options pattern to keep constructors clean and composable.

Example (Wire):

```go
w := builder.NewWire[string](
    ctx,
    builder.WireWithConcurrencyControl[string](1024, 8),
    builder.WireWithTransformer[string](func(v string) (string, error) {
        return strings.ToUpper(v), nil
    }),
    builder.WireWithCircuitBreaker[string](cb),
    builder.WireWithSurgeProtector[string](sp),
    builder.WireWithSensor[string](sensor),
)
```

Notes:

* Options are meant to be applied **before** `Start()`.
* Configuration mutation while running is not supported.

---

## 🧠 How builder maps to internal

The pattern is consistent across components:

1. `builder.NewX(...)` delegates to `pkg/internal/x.NewX(...)`.
2. `builder.XWithY(...)` options call the underlying component’s connect/set method.
3. The returned value is a ready-to-start component implementing the shared `types` contract.

This gives you:

* stable constructors
* predictable option naming
* internal flexibility without exposing internals

---

## 📂 Package layout

Builder exposes one file per component family (names may vary as the repo evolves, but the idea holds):

| File                | Responsibility                                |
| ------------------- | --------------------------------------------- |
| `wire.go`           | Wire constructor + Wire options (`WireWith…`) |
| `conduit.go`        | Conduit constructor + options                 |
| `circuitbreaker.go` | Circuit breaker constructor + options         |
| `surgeprotector.go` | Surge protector constructor + options         |
| `resister.go`       | Resister/queue wiring + options               |
| `sensor.go`         | Sensor constructor + options                  |
| `meter.go`          | Meter constructor + options                   |
| `logger.go`         | Logger constructors/adapters + options        |
| `codec.go`          | Encoder/decoder helpers + options             |
| `adapter.go`        | Integration adapters (HTTP, auth, etc.)       |
| `forwardrelay.go`   | Outbound relay constructors/options           |
| `receivingrelay.go` | Inbound relay constructors/options            |
| `plug.go`           | Plug constructors/options                     |
| `generator.go`      | Generator constructors/options                |

---

## 🚄 Performance stance

Builder is configuration-time glue. The performance characteristics of a running pipeline are dominated by the internal engine (especially `wire`).

If you’re optimizing throughput:

* keep the hot path minimal (transform + emit)
* keep telemetry and encoding opt-in
* prefer worker-local state (factory transforms / scratch buffers) over shared pooling unless measured

---

## 🔧 Extending builder

When you add a capability, keep the layering clean:

1. Add/extend the contract in `pkg/internal/types` (if it’s cross-component).
2. Implement it in the relevant `pkg/internal/*` component.
3. Expose it through `pkg/builder` via:

   * a constructor update (if needed), and/or
   * a new `XWith…` option.

The goal is for consumers to never need to touch `pkg/internal/*`.

---

## 📖 Further Reading

- **[Root README](../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../internal/README.md)** – In-depth explanation of how the `internal/` directory works.
- **[Examples Directory](../../example/)** – Real-world use cases demonstrating the builder API in action.

---

## 📝 License

The **builder package** is part of Electrician and is released under the [Apache 2.0 License](../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚙️🚀** If you have questions or need support, feel free to open a GitHub issue.
