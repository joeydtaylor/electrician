# Plug

Package `plug` implements Electrician's ingestion binding layer. External consumers should use the builder APIs under `pkg/builder`.

A plug owns adapter wiring and adapter functions. It provides generators with a single surface to pull from without coupling to specific adapter implementations.

## Configuration

Configuration is expected to happen before use. Once a plug is frozen by a generator, `Connect*`, `AddAdapterFunc`, and `SetComponentMetadata` panic if called.

Typical construction:

```go
ctx := context.Background()

p := plug.NewPlug[Item](ctx,
	plug.WithAdapter[Item](httpAdapter),
	plug.WithAdapterFunc[Item](adapterFunc),
)
```

## Execution model

Plugs do not execute work directly. Generators call `GetAdapterFuncs` and `GetConnectors` and invoke adapters according to their own scheduling and cancellation rules.

## Telemetry

Plugs store sensors and loggers for ingestion components. Loggers receive best-effort events via `NotifyLoggers`.

## Package layout

- `plug.go`: `Plug` type and constructor.
- `accessors.go`: getters for adapter funcs, connectors, and metadata.
- `config.go`: configuration setters.
- `connect.go`: adapter and sensor wiring helpers.
- `immutability.go`: runtime guard against mutation after Freeze.
- `options.go`: functional options.
- `telemetry.go`: logger notifications.
- `*_test.go`: tests.

## Notes

- Plug configuration should be completed before a generator starts.
- Adapter functions should remain small; use dedicated adapters for complex ingestion logic.
