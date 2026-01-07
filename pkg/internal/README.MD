# ⚙️ `pkg/internal/`

Welcome to the **engine room** of Electrician.

Everything under `pkg/internal/` is **implementation**, not public API. Go enforces this with the `internal/` import rule:

- ✅ Any package **under** `pkg/` can import `pkg/internal/...`
- ❌ Packages **outside** `pkg/` (and downstream users of this module) cannot

In this repo, that means:

- `pkg/` is the **public surface area** (what other packages in this module are meant to import)
- `pkg/internal/` is **private implementation** for `pkg/`
- `pkg/builder` is the **public exporter/wrapper**, assembling internal components via options/builders

---

## 🧠 What belongs in `internal/`

- Pipeline primitives (concurrency, backpressure, lifecycle)
- Reliability components (circuit breakers, rate limiting, recovery)
- Telemetry hooks (loggers/sensors/meters)
- Adapters + relays for integration surfaces (HTTP, gRPC, Kafka, S3, codecs, etc.)
- Shared contracts (`types/`) and small utilities (`utils/`)

What does *not* belong here:

- Stable user-facing ergonomics → `pkg/builder`
- Public API commitments → outside `pkg/internal/`

---

## 📦 Packages (high-level map)

| Package | Role |
|---|---|
| `adapter/` | Integration-facing adapters (external systems / protocols / auth) |
| `httpserver/` | Inbound HTTP/HTTPS listener + request handling |
| `circuitbreaker/` | Failure detection + trip/reset behavior |
| `codec/` | Serialization / compression / encoding helpers used by pipelines |
| `conduit/` | Multi-stage wiring (linking components / wires) |
| `forwardrelay/` | Outgoing gRPC relay (send streams/events) |
| `receivingrelay/` | Incoming gRPC relay (receive/ingest streams/events) |
| `relay/` | Protobuf / gRPC contract code (generated + helpers) |
| `generator/` | Autonomous sources (synthetic feeds, pollers, etc.) |
| `internallogger/` | Structured logging backend used across internal components |
| `meter/` | Performance + counters (throughput, errors, timing) |
| `plug/` | Pluggable input source abstraction (adapter-like) |
| `resister/` | Queue-based smoothing / deferred delivery (rate-limit queue) |
| `sensor/` | Telemetry hooks/callbacks for pipeline events |
| `surgeprotector/` | Rate limiting + overload behavior around submission |
| `types/` | **Source of truth**: shared interfaces/contracts across components |
| `utils/` | Small internal helpers (hashes, common glue) |
| `wire/` | Core pipeline primitive: concurrent ingest → transform → emit |

---

## 🧱 File layout conventions

Many internal packages follow a consistent “shape”, but it’s not enforced everywhere:

| Common file | Typical purpose |
|---|---|
| `package.go` | Primary type + constructor |
| `api.go` | Methods intended to be called by other Electrician packages |
| `internal.go` | Private implementation details |
| `options.go` | Functional options (`With*`) for configuration |
| `*_test.go` | Unit/integration tests |

Notes:

- Don’t assume every package uses every file name but if they have need of those features they will follow this convention — follow the code.

---

## 🧩 Design contract (configuration-first)

Electrician is built around one operational rule:

✅ Configure → Start → Run → Stop/Restart

- Components are expected to be connected/configured **before** `Start()`.
- Mutating configuration while a pipeline is running is **not supported** and may race.

This is a performance and correctness choice: predictable concurrency semantics and lean hot paths.

---

## 🚄 Performance philosophy (what we actually mean)

Electrician’s performance work is targeted at the **core pipeline path** (especially `wire`):

- ✅ Avoid **per-element allocations** in core pipeline mechanics under typical configs.
- ✅ Prefer **worker-local state** (transformer factories, `WithScratchBytes`) over shared pools.
- ✅ Keep telemetry **best-effort** so it doesn’t backpressure the pipeline.
- ✅ Catch regressions with tests (e.g. alloc-budget tests in `wire_test.go`).

What we do *not* claim:

- ❌ “Zero allocations everywhere.” Integrations, encoders, and user transforms can allocate.

---

## 🧰 External dependencies

Electrician is **stdlib-first** for the pipeline core, but the repo is not dependency-free because integrations need real libraries for real protocols/file formats.

Two practical truths:

- ✅ Most third-party deps are **integration surfaces** (Kafka, S3, Parquet, compression). If you don’t import those packages, they won’t be linked into your final binary.
- ✅ The hot-path primitives aim for **no per-element allocations** — but that does not magically apply to everything you plug into them.

### ✅ Direct dependencies by purpose (high level)

| Area | Libraries | What they’re used for |
|---|---|---|
| 🧾 Logging | `go.uber.org/zap` | Structured logging backend |
| 📡 RPC / Messaging | `google.golang.org/grpc`, `google.golang.org/protobuf` | gRPC relays + cross-service streaming contracts |
| ☁️ AWS | `github.com/aws/aws-sdk-go-v2/*` | S3/STS and related AWS integrations |
| 🧵 Kafka | `github.com/segmentio/kafka-go` | Kafka producers/consumers for adapters and pipelines |
| 🗜️ Compression / Codecs | `github.com/klauspost/compress`, `github.com/golang/snappy`, `github.com/pierrec/lz4`, `github.com/andybalholm/brotli` | Optional compression paths (often via codecs/Parquet) |
| 🧱 Parquet | `github.com/parquet-go/parquet-go` | Parquet IO (columnar formats) |
| 📈 Numerics / DSP | `gonum.org/v1/gonum`, `github.com/mjibson/go-dsp` | Optional math/DSP utilities where needed |
| 🖥️ Host metrics | `github.com/shirou/gopsutil` | System/host introspection for telemetry |
| 🌐 Network helpers | `golang.org/x/net` | Supplemental networking primitives |

We keep these dependencies because they’re **boring, maintained, and widely deployed**. They exist to support integrations — not to reinvent Kafka, Parquet, AWS, or gRPC.

---

## 🔧 Extending internal packages

If you’re adding capabilities:

1) Update interfaces/contracts in `pkg/internal/types` if you need new cross-component behavior  
2) Implement on the internal component  
3) Add a `WithX` option if it’s configuration-time behavior  
4) Add tests (and alloc tests where it’s a hot path)

---

## 📖 Further Reading

- **[Root README](../../README.md)** – Electrician’s overall architecture and principles.
- **[Per-Package README]** – Each internal sub-package (`wire`, `circuitbreaker`, `httpserver`, etc.) contains additional details.
- **[Examples Directory](../../example/)** – Demonstrates these internal components in real-world use cases.

---

## 📝 License

All internal packages fall under Electrician’s [Apache 2.0 License](../../LICENSE).  
You are free to use, modify, and distribute them under these terms, but note that **they are not designed for direct external use**.
