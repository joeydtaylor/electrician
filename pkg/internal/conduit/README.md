# 🔌 Conduit Package

A **Conduit** is Electrician’s **composition layer**.

It links multiple components (most commonly multiple **Wires**) into a single logical pipeline so you can treat a multi-stage flow as “one thing” for lifecycle and submission.

Conduit is not a new processing engine. The work still happens inside the stages you attach (wires, relays, etc.). Conduit’s job is orchestration and routing.

---

## ✅ What a Conduit does

| Capability                 | Meaning                                                                      |
| -------------------------- | ---------------------------------------------------------------------------- |
| 🔗 Stage composition       | Connect stages into a multi-step pipeline (most commonly wire → wire → …).   |
| 🚦 Lifecycle orchestration | Start/stop/restart the composed stages as a unit.                            |
| 📥 Unified submission      | Provide a single `Submit` surface that feeds the first stage.                |
| 📤 Unified output          | Expose the “tail” output (typically the last stage’s output channel/buffer). |

---

## ❌ What a Conduit does *not* do

* ❌ It does not magically add durability or delivery guarantees.
* ❌ It is not a load balancer.
* ❌ It does not replace circuit breakers/surge protectors/insulators — those are per-stage concerns.

If you want resilience behaviors, attach them to the stages (e.g., breakers or surge protectors on wires) and let the conduit orchestrate those stages.

---

## 🧠 How it fits in a pipeline

A common shape:

**Generator → Wire (ingest/normalize) → Wire (transform/enrich) → Wire (encode/emit)**

A conduit gives you:

* one “thing” to start/stop
* one “thing” to submit into
* one “thing” to read output from

Under the hood, it forwards stage output into the next stage’s input using the contracts defined in `types/`.

---

## ⚙️ Configuration contract

Conduit follows Electrician’s standard model:

✅ Configure → Start → Run → Stop/Restart

* Build the stage chain before `Start()`.
* Don’t mutate the stage list while running.

If you need dynamic routing, model that explicitly (multiple conduits, explicit fan-out/fan-in stages, external brokers, etc.).

---

## 📂 Package structure

| File         | Purpose                       |
| ------------ | ----------------------------- |
| `conduit.go` | Type definition + constructor |
| `api.go`     | Public methods / wiring       |
| `options.go` | Functional options (`With*`)  |
| `*_test.go`  | Tests                         |

---

## 🔧 Extending Conduit

When adding capability:

* Cross-component contract change → update `types/conduit.go` first.
* Conduit-only behavior → implement in `pkg/internal/conduit`.
* User-facing knob → expose via `pkg/builder` (`ConduitWith…`).

Tests should cover:

* correct stage chaining
* start/stop behavior across all stages
* cancellation behavior
* no goroutine leaks when downstream stages stop

## 📖 Further Reading

- **[Root README](../../../README.md)** – Electrician’s overall architecture and principles.
- **[Internal README](../README.md)** – How `internal/` packages interact with `types/`.
- **[Examples Directory](../../../example/conduit_example/)** – Demonstrates **Conduits in a real-world pipeline**.

---

## 📝 License

The **Conduit package** is part of Electrician and is released under the [Apache 2.0 License](../../../LICENSE).  
You’re free to use, modify, and distribute it within these terms.

---

**Happy wiring! ⚡🔗** If you have questions or need support, feel free to open a GitHub issue.
