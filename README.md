# ⚡ Electrician

**Electrician** is a Go library for building **typed, concurrent pipelines**.

It’s generics-first and centered around a small set of primitives (Wire, Conduit, Relays) that let you build ingestion → transform → emit flows without hand-rolling goroutine topologies and channel plumbing every time.

---

## 🚀 What it is

* 🧬 **Generics-first**: end-to-end type safety without `interface{}`.
* ⚙️ **Configurable concurrency**: buffer sizes + worker counts are first-class.
* 🧯 **Reliability components**: circuit breakers, retry/recovery (insulators), surge protection / queueing.
* 🧱 **Composable pipelines**: wires + conduits to build multi-stage flows.
* 📡 **Telemetry hooks**: structured logging + sensors/meters for visibility.
* 🌐 **Integration surfaces**: optional adapters/relays for gRPC, Kafka, AWS, Parquet/codecs, etc.

Electrician is designed to be boring under load: predictable lifecycle, explicit wiring, and hot paths that don’t allocate per item unless you make them.

---

## 🎯 Goals & philosophy

1. 🧠 **Clarity over cleverness** — configure, start, run, stop.
2. 🧬 **Type safety** — pipeline contracts are generic all the way through.
3. 🚄 **Performance where it matters** — core pipeline aims to avoid per-element allocations.
4. 🧩 **Modularity** — reliability and integrations are opt-in.

The core operational contract is:

✅ Configure → Start → Submit/Run → Stop/Restart

Mutating pipeline configuration while running is not supported (race risk by design).

---

## 🧩 Project layout

Electrician follows a “public surface + private engine” layout:

* `pkg/` — public packages intended to be imported by other code in this module
* `pkg/internal/` — private implementation for `pkg/` (enforced by Go’s `internal/` rule)
* `pkg/builder` — the public exporter/wrapper that assembles internal components via options/builders

Downstream users import `pkg/builder` (or other `pkg/*` packages), not `pkg/internal/*`.

---

## 🚄 Performance posture 

Electrician does **not** claim “zero allocations everywhere.”

What it does claim:

* ✅ The **pipeline core** (especially `wire`) is written toC to avoid **per-element allocations** under typical configurations.
* ✅ Fast-path processing exists when “extras” are off (no breaker/surge/insulator/encoder/telemetry and exactly one transform source).
* ✅ Worker-local scratch patterns (transformer factories / `WithScratchBytes`) let you stay allocation-free without `sync.Pool`.
* ✅ Allocation regressions are caught with tests where it matters (see `wire_test.go`).

Your code can still allocate:

* user transformers
* encoders/serialization
* integrations/adapters

If you care about the last 20%, benchmark and keep the hot path minimal.

---

## 📦 Dependencies 

Electrician is **stdlib-first** for core pipeline mechanics (channels, `context`, `sync`, `net/http`, etc.).

The repository includes third-party libraries primarily because integrations need them:

* gRPC/protobuf for relay contracts
* Kafka clients
* AWS SDK v2
* Parquet + compression codecs
* structured logging
* optional numerics/DSP and host metrics

These are deliberate choices: widely used, maintained, and boring.

Important detail: if you don’t import a given integration package, it won’t end up in your final binary.

---

## 🔧 Getting started

### 1️⃣ Install

```bash
go get github.com/joeydtaylor/electrician
```

### 2️⃣ Build a tiny pipeline (Wire)

This example:

* receives strings
* uppercases them
* reads results from the output channel

```go
package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	transform := func(input string) (string, error) {
		return strings.ToUpper(input), nil
	}

	w := builder.NewWire(
		ctx,
		builder.WireWithTransformer(transform),
	)

	_ = w.Start(ctx)
	_ = w.Submit(ctx, "hello")
	_ = w.Submit(ctx, "world")

	// Read outputs (or use LoadAsJSONArray for debugging/testing).
	for {
		select {
		case v, ok := <-w.GetOutputChannel():
			if !ok {
				return
			}
			fmt.Println(v)
		case <-ctx.Done():
			_ = w.Stop()
			return
		}
	}
}
```

---

## 🧱 Core concepts

* ⚡ **Wire**: bounded concurrent stage (ingest → transform → emit). Optional breaker/surge/insulator/telemetry.
* 🧵 **Conduit**: composition tool for chaining stages.
* 🧯 **Circuit breaker**: reject/divert when tripped.
* 🧪 **Insulator**: retry/recovery policy on transform failures.
* 🛡️ **Surge protector / resister queue**: rate limiting + deferred processing.
* 📡 **Sensors / meters / loggers**: hooks for observability.
* 🌐 **Relays**: protobuf + gRPC contracts for cross-service streaming.

---

## 📝 License

[Apache 2.0 License](./LICENSE).
---

## 📖 Further Reading
- **[Per-Package README](./pkg/internal/)** – Each internal sub-package (`wire`, `circuitbreaker`, etc.) contains additional details.
- **[Examples Directory](./example/)** – Demonstrates these internal components in real-world use cases.

---

## 📝 License

[Apache 2.0 License](./LICENSE).  

---

## ⚡ Happy wiring! 🚀

If you have any questions or need clarification, feel free to **open a GitHub issue**.
