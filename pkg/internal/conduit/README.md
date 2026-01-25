# conduit

The conduit package composes multiple wires into a multi-stage pipeline. It provides a single lifecycle to start and stop the connected stages and helps coordinate backpressure between them.

## Responsibilities

- Connect multiple wires into a processing chain.
- Manage start/stop for the chain.
- Provide a single output surface for downstream consumption.

## Key types and functions

- Conduit[T]: main composition type.
- Start(ctx) / Stop(): lifecycle management.
- LoadAsJSONArray(): convenience output helper for debugging and tests.

## Configuration

Conduits are configured with:

- One or more wires (ordered stages)
- Optional concurrency controls at the conduit level
- Optional sensor for aggregated telemetry

Configuration must be finalized before Start().

## Lifecycle

1) Construct and attach wires
2) Start(ctx)
3) Allow the chain to process
4) Stop() to drain and terminate

## Concurrency

Each wire retains its own concurrency settings. The conduit focuses on orchestration, not per-stage execution details.

## Observability

Conduits can emit telemetry through sensors and meters attached to their wires or at the conduit level.

## Usage

```go
conduit := builder.NewConduit(
    ctx,
    builder.ConduitWithWire(wireA),
    builder.ConduitWithWire(wireB),
)

conduit.Start(ctx)
conduit.Stop()
```

## References

- examples: example/conduit_example/
- builder: pkg/builder/conduit.go
- internal contracts: pkg/internal/types/conduit.go
