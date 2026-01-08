# 📡 Sensor Package

The **Sensor** package is Electrician’s **telemetry hook surface**.

A sensor is a typed receiver of pipeline events (start/stop/submit/errors/etc.). Components call sensor methods inline as work flows through the system. This lets you attach:

* counters/meters
* structured event emission
* tracing correlation
* lightweight auditing/debug hooks

Sensors are **not** a metrics backend by themselves. They’re the hook points.

---

## ✅ What sensors are (and aren’t)

* ✅ A **contract** (`types.Sensor[T]`) + implementation(s) that react to pipeline events.

* ✅ A way to centralize “what happened” without littering business logic with instrumentation.

* ✅ Typed by the element flowing through the component (`T`).

* ❌ Not a guarantee of non-blocking behavior: sensors are invoked **inline** by components. Keep handlers fast or offload work inside the sensor.

* ❌ Not a promise that every component emits every event. Events are emitted by the components that call them.

---

## 📦 Package structure

File layout is intentionally simple. Names may evolve, but this is the typical split:

| File         | Purpose                                        |
| ------------ | ---------------------------------------------- |
| `sensor.go`  | Core type(s) + constructor                     |
| `api.go`     | Public methods / interface fulfillment         |
| `options.go` | Functional options (`With*`) for configuration |
| `*_test.go`  | Unit tests                                     |

---

## 🧠 Event model

Sensors are invoked by components at key points in the lifecycle and hot path.

Common event families (based on how the pipeline currently calls sensors):

### 🚦 Lifecycle

* `OnStart` / `OnStop`
* `OnComplete` (when a component finishes a cycle)

### 📥 Flow

* `OnSubmit`
* `OnCancel` (context cancellation)
* `OnElementProcessed` (successful processing)

### ❌ Errors

* `OnError` (transform/encode/processing failures)

### 🧪 Recovery (Insulator)

* `OnInsulatorAttempt`
* `OnInsulatorSuccess`
* `OnInsulatorFailure` (retries exhausted)

### 🧯 Circuit breaker

* `OnCircuitBreakerNeutralWireSubmission` (diverted to neutral/ground wire)
* `OnCircuitBreakerDrop` (tripped and no neutral wires)

### 🛡️ Surge protector

* `OnSurgeProtectorSubmit` (item routed to protector handling)
* `OnSurgeProtectorRateLimitExceeded` (rate limited / deferred)

If you add new event types, they should be added to `types/sensor.go` and then invoked by the components that own the behavior.

---

## ⚙️ Configuration contract

Sensors are configuration-time wiring.

* Attach sensors before `Start()`.
* Mutating sensor sets while a component is running is not supported.

This is how Electrician avoids locks/allocations in the hot path.

---

## 📈 Meters and aggregation

A common pattern is:

* a **Meter** collects counts/timings
* a **Sensor** updates that meter when events fire

Example wiring (conceptual):

```go
meter := builder.NewMeter[Event](ctx, builder.MeterWithTotalItems[Event](n))
sensor := builder.NewSensor(builder.SensorWithMeter[Event](meter))

w := builder.NewWire(
    ctx,
    builder.WireWithSensor(sensor),
    builder.WireWithTransformer(fn),
)
```

Whether you track latency, errors, throughput, etc. is a sensor/meter implementation detail.

---

## 🔧 Extending sensors

When adding a new telemetry surface:

1. Add the event method to `pkg/internal/types/sensor.go`.
2. Implement it in the sensor implementation.
3. Invoke it from the owning component at the correct point.
4. Add tests that validate:

   * the event fires exactly when expected
   * it does not break the pipeline if the sensor is nil/missing
   * performance impact stays acceptable (no accidental allocations in hot paths)

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../../internal/README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/sensor_example/)** – Demonstrates how Sensors integrate with Electrician’s event system.

---

## 📝 License

The **Sensor package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

## ⚡ Happy wiring! 🚀

If you have any questions or need support, feel free to **open a GitHub issue**.
