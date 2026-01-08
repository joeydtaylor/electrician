# 📜 Internal Logger Package

The **internal logger** package provides a Zap-backed implementation of Electrician’s logging interface (`types.Logger`).

Electrician components emit structured events (start/stop/submit/errors/etc.) as key/value pairs. This package is the default “structured logger” implementation used by the framework.

---

## ✅ What it is

* A **Zap-based logger** wired to Electrician’s `types.Logger` contract.
* A **structured logging** sink for pipeline events.
* A place to standardize field shapes (component metadata, event names, results, errors).

## ❌ What it isn’t

* A promise of “zero allocations everywhere.”
* A metrics system (that’s meters + sensors).
* A network transport.

Zap is chosen because it’s widely deployed and designed for **low-overhead structured logging**. Actual allocation behavior depends on how you use it.

---

## 🧠 How it’s used by the pipeline

Core components (e.g. `wire`) call into loggers with the pattern:

* `level` (debug/info/warn/error…)
* `msg`
* `keysAndValues ...interface{}`

To keep hot paths lean, components may skip log calls when there are no connected loggers.

### Level filtering

Some components optionally check for a level gate via:

```go
type levelChecker interface { IsLevelEnabled(types.LogLevel) bool }
```

If your logger implements `IsLevelEnabled`, components can avoid emitting logs at disabled levels (less work and fewer allocations).

---

## 🚄 Performance stance (accurate)

Zap is optimized for structured logging, but “zero alloc” depends on usage.

To keep overhead down:

* Prefer typed fields over string formatting.
* Avoid `fmt.Sprintf` in log paths.
* Avoid logging large structs/maps on hot paths.
* Gate noisy logs behind `Debug` and ensure debug is disabled in production.

If you need hard numbers, benchmark with your actual event volume and field payload sizes.

---

## 📂 Package structure

| File                | Purpose                              |
| ------------------- | ------------------------------------ |
| `internallogger.go` | Type definition + constructor        |
| `api.go`            | Methods that satisfy `types.Logger`  |
| `options.go`        | Functional options for configuration |
| `internal.go`       | Zap configuration helpers            |
| `*_test.go`         | Tests                                |

---

## ⚙️ Configuration model

Configuration is intended to be done at construction time via options (typically exposed through `pkg/builder`).

Common knobs in a Zap-backed logger (exact options depend on the builder wrapper):

* output encoding (JSON vs console)
* minimum log level
* development vs production presets
* caller / stacktrace behavior

---

## 🔧 Extending the internal logger

* If you need new behavior exposed to consumers, add a builder option that configures the underlying Zap logger.
* Only change `types/logger.go` if you need to change the **logging contract** across the framework.
* Keep the logger implementation boring: logging must never become the bottleneck of the pipeline.


## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.MD)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/logging/)** – Demonstrates **real-world logging configurations**.

---

## 📝 License

The **Internal Logger package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy logging! 📝⚡** If you have questions or need support, feel free to open a GitHub issue.
