# plug

The plug package defines a pluggable input source. A plug wraps one or more adapters (or adapter functions) and exposes a uniform interface for generators to pull or receive data.

## Responsibilities

- Encapsulate adapter(s) that produce items.
- Provide a consistent API for generators.
- Emit telemetry through sensors and loggers.

## Key types and functions

- Plug[T]: main type.
- Serve(ctx, submit): invoke adapters and submit items.

## Configuration

Plugs are configured with:

- Adapter implementations or adapter functions
- Optional sensor and logger

Configuration must be complete before Start/Serve.

## Lifecycle

Plugs are typically driven by a generator. They do not manage long-lived goroutines unless an adapter does so internally.

## Observability

Plugs emit start/stop and submission events to sensors, enabling meters and loggers to track ingestion.

## Usage

```go
plug := builder.NewPlug(
    ctx,
    builder.PlugWithAdapterFunc(myAdapter),
    builder.PlugWithSensor(sensor),
)
```

## References

- examples: example/plug_example/
- builder: pkg/builder/plug.go
- internal contracts: pkg/internal/types/plug.go
