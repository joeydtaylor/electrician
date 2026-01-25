# wire

The wire package is the core pipeline primitive in Electrician. A wire accepts inputs, applies one or more transforms, and emits outputs with bounded concurrency. It is designed for predictable lifecycle behavior and a low-allocation hot path.

## Responsibilities

- Accept submissions and route them to workers.
- Apply transform functions in order.
- Emit outputs and collect errors.
- Integrate optional reliability components (circuit breaker, insulator, surge protector, resister).
- Emit telemetry through sensors and loggers.

## Key types and functions

- Wire[T]: the main pipeline type.
- Start(ctx): starts worker goroutines and begins processing.
- Stop(): stops processing and closes outputs.
- Submit(ctx, item): submit an item into the wire.
- GetOutputChannel(): read processed output.
- LoadAsJSONArray(): convenience helper for debugging and tests.

## Configuration

Wires are configured via functional options before Start(). Common options include:

- Transformer functions or transformer factories
- Concurrency and buffering
- Circuit breaker or insulator (retry)
- Surge protector / resister
- Sensor (telemetry)
- Logger

Configuration mutation after Start() is not supported and will panic to avoid races.

## Lifecycle

1) Construct with options
2) Start(ctx)
3) Submit items
4) Stop() when done

The wire manages its own worker lifecycle and respects the provided context for cancellation.

## Concurrency and performance

- Workers process items concurrently with bounded queues.
- The fast path avoids per-element allocations when optional components are disabled.
- Transform factories and scratch buffers allow worker-local state without shared pools.

## Error handling

- Transform errors can be counted, reported, and routed to an insulator if configured.
- Circuit breakers can trip and divert or drop items based on policy.
- Surge protectors and resisters handle rate limiting and queueing.

## Observability

Wire integrates with Sensor and Logger. Sensors emit events (submit, processed, error), and meters aggregate counts and rates.

## Usage

```go
w := builder.NewWire[Item](
    ctx,
    builder.WireWithTransformer(process),
    builder.WireWithConcurrencyControl[Item](1024, 8),
    builder.WireWithSensor(sensor),
)

_ = w.Start(ctx)
_ = w.Submit(ctx, item)
_ = w.Stop()
```

## References

- examples: example/wire_example/
- builder: pkg/builder/wire.go
- internal contracts: pkg/internal/types/wire.go
