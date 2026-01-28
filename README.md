# Electrician

Electrician is a Go library for building typed, concurrent pipelines. It focuses on clear lifecycle semantics, explicit wiring, and predictable behavior under load. The library provides a small set of primitives that cover the common problems of ingestion, transformation, and delivery without forcing you to hand-roll goroutine topologies or channel plumbing.

## What problems it solves

Electrician targets the hard parts of production pipeline work:

- Coordinating concurrent stages while keeping type safety end-to-end.
- Avoiding data races when configuration is updated at runtime.
- Keeping hot paths allocation-light while still supporting rich integrations.
- Handling failure modes (trip/reset, retries, queueing, rate limits) in a repeatable way.
- Adding observability without coupling metrics/logging to the processing path.

These problems are solvable with bespoke goroutines and channels, but the complexity grows quickly. Electrician makes the lifecycle and wiring explicit so you can focus on the business logic.

## Core concepts

- Wire: bounded concurrent stage (ingest -> transform -> emit).
- Conduit: composition for multi-stage pipelines.
- Plug + Generator: pluggable sources for ingestion.
- Circuit breaker: trip/reset behavior on failure.
- Surge protector + Resister: rate limiting and queueing under pressure.
- Sensor + Meter + Logger: observability hooks and counters.
- Relays: gRPC contracts for cross-service streaming.
- Adapters: HTTP, Kafka, S3, codecs, and more.
- Jack HTTP server: inbound HTTP/HTTPS entrypoint.

## Design contract

Configure -> Start -> Submit/Run -> Stop/Restart

Configuration is expected to be complete before Start(). Mutating a running component is not supported and may race.

## Getting started

Install:

```bash
go get github.com/joeydtaylor/electrician
```

Minimal example:

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

## Performance posture

The core pipeline path (especially wire) is designed to avoid per-element allocations in typical configurations. Integrations, encoders, and user transforms can allocate; keep the hot path minimal when optimizing. Worker-local state (factory transforms or scratch buffers) is preferred over shared pools unless measured.

## Observability

Observability is opt-in and explicit. Sensors emit events, meters aggregate counts and rates, and loggers handle structured output. This keeps the pipeline core fast and makes telemetry easy to reason about.

## Integrations

Electrician includes adapters for HTTP, Kafka, S3, codecs, and relays. These live in separate packages and are not linked into your binary unless imported.

## Project layout

- pkg/builder: public construction layer (stable API).
- pkg/internal: private implementation (enforced by Go internal).
- example/: runnable examples covering most components.

## Documentation

- Internal architecture: `pkg/internal/README.MD`
- Relay examples index: `example/relay_example/README.md`
- Frontend integration (gRPC-Web + QUIC): `example/relay_example/FRONTEND_INTEGRATION_UNIFIED.md`
- QUIC relay docs (including Rust guide): `example/relay_example/quic/README.md`
- Mock OAuth server (dev only): `example/relay_example/mock_oauth_server/README.md`
- Onboarding: `docs/ONBOARDING.md`
- Relay gRPC-Web / Connect: `docs/relay-grpc-connect.md`
- Relay CORS: `docs/relay-grpcweb-cors.md`
- JWT issuer mismatch: `docs/relay-issuer-mismatch.md`
- Steeze UI system (if UI is added here): `docs/steeze-ui-system.md`
- Workbench (local codegen playground): `docs/WORKBENCH.md`
- Legal: `docs/LEGAL.md`
- Trademarks: `docs/TRADEMARKS.md`
- Per-package READMEs under `pkg/internal/*` and `pkg/builder`

## Integration tests (LocalStack + Redpanda)

Some adapter behaviors (S3/Kafka) are verified via integration tests gated by the `integration` build tag. The repo includes a LocalStack + Redpanda compose stack under `local-stack/`.

```bash
docker network create steeze-edge || true
docker compose -f local-stack/docker-compose.yml up -d
go test ./pkg/... -tags=integration -count=1
```
