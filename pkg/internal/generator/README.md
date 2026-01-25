# generator

The generator package drives ingestion by coordinating a plug and optional control components (such as a circuit breaker). Generators act as active sources in a pipeline.

## Responsibilities

- Invoke a plug to produce items.
- Apply circuit breaker policy when configured.
- Emit telemetry events for lifecycle and submissions.

## Key types and functions

- Generator[T]: main type.
- Start(ctx) / Stop(): lifecycle control.

## Configuration

Generators are configured with:

- A plug (required)
- Optional circuit breaker
- Optional sensor and logger

Configuration mutation after Start() is not supported.

## Lifecycle

1) Construct generator and attach plug
2) Start(ctx) to begin ingestion
3) Stop() to end ingestion

## Observability

Sensors receive events for submit, error, and lifecycle transitions. Meters can compute ingestion rates and error counts from these events.

## Usage

```go
generator := builder.NewGenerator(
    ctx,
    builder.GeneratorWithPlug(plug),
    builder.GeneratorWithSensor(sensor),
)

generator.Start(ctx)
```

## References

- examples: example/generator_example/
- builder: pkg/builder/generator.go
- internal contracts: pkg/internal/types/generator.go
