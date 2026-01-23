# Generator

Package `generator` implements Electrician’s internal source component. External consumers should use the builder APIs under `pkg/builder`.

A generator runs plug adapter functions and connectors, producing values of type `T` and submitting them to downstream components (typically a wire). It is a lightweight driver; heavy processing belongs in the wire.

## Configuration

Configuration is expected to happen before `Start`. `Connect*` and `Set*` methods panic if called after the generator has started.

Typical construction:

```go
ctx := context.Background()

plug := plug.NewPlug[Item](ctx,
	plug.WithAdapterFunc[Item](adapterFunc),
)

gen := generator.NewGenerator[Item](ctx,
	generator.WithPlug[Item](plug),
)
```

## Lifecycle

- `Start(ctx)` begins plug execution and enables submissions.
- `Stop()` cancels the generator context and waits for plugs to exit.
- `Restart(ctx)` stops the generator and starts again with the new context.

`Stop()` blocks until plug goroutines exit. Plugs are expected to honor context cancellation.

## Execution model

- Adapter funcs and connectors run in goroutines started by `Start`.
- Submissions are routed to all connected submitters.
- If a submitter implements `FastSubmit`, it is used to avoid unnecessary overhead.
- An optional circuit breaker can gate submissions and connector restarts.

## Telemetry

Sensors receive `OnStart`, `OnStop`, and `OnRestart` events. Loggers receive best-effort log events through `NotifyLoggers`.

## Package layout

- `generator.go`: `Generator` type and constructor.
- `accessors.go`: metadata and state accessors.
- `config.go`: configuration setters.
- `connect.go`: component attachment helpers.
- `immutability.go`: runtime guard against mutation after `Start`.
- `internal.go`: submission path and control loops.
- `lifecycle.go`: start/stop/restart.
- `options.go`: functional options.
- `telemetry.go`: logger/sensor notifications.
- `*_test.go`: tests.

## Notes

- Generator configuration is immutable after `Start`.
- Plug logic should be context-aware to allow clean shutdowns.
- For durable ingestion, place a real broker upstream and treat the generator as a local driver.
