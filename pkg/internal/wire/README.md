# Wire

Package `wire` implements the core in-process pipeline stage used by Electrician. It is an internal package; external consumers should use the builder APIs under `pkg/builder`.

A wire receives elements, applies a sequence of transforms, and emits results. Optional components can enforce circuit breaking, rate limiting, retry, encoding, and telemetry.

## Quick start

```go
ctx := context.Background()

w := wire.NewWire[int](ctx,
	wire.WithConcurrencyControl[int](1024, 4),
	wire.WithTransformer[int](func(v int) (int, error) {
		return v * 2, nil
	}),
)

_ = w.Start(ctx)
_ = w.Submit(ctx, 21)

out := <-w.GetOutputChannel()
_ = w.Stop()
```

## Configuration

Configuration is expected to happen before `Start`. `Connect*` and `Set*` methods panic if called after the wire has started.

You can configure a wire using options at construction time:

```go
w := wire.NewWire[Event](ctx,
	wire.WithComponentMetadata[Event]("ingest", "wire-1"),
	wire.WithConcurrencyControl[Event](65536, 12),
	wire.WithInsulator[Event](retryFn, 3, 50*time.Millisecond),
	wire.WithSurgeProtector[Event](protector),
	wire.WithCircuitBreaker[Event](breaker),
	wire.WithLogger[Event](logger),
	wire.WithSensor[Event](sensor),
)
```

Direct connect/set methods are also available (`Connect*`, `Set*`) but should be treated as configuration-time operations.

## Lifecycle

- `Start(ctx)` starts workers, resister processing, and generators.
- `Stop()` cancels the wire context, waits for workers, and closes output/error channels.
- `Restart(ctx)` stops the wire, reinitializes channels, and starts again.

## Fast path

A fast path is used when the wire has no optional components and exactly one transform source:

- no circuit breaker
- no surge protector
- no insulator
- no encoder
- no loggers/sensors
- one transform (explicit or factory)

This avoids per-element overhead in the hot loop. User transforms can still allocate.

## Output

Processed elements are sent to the output channel (`GetOutputChannel`). If an encoder is configured, elements are also written to `OutputBuffer` under a mutex.

`Load()` stops the wire and returns a copy of the output buffer. `LoadAsJSONArray()` stops the wire and drains available output into a JSON array (non-blocking).

## Package layout

- `wire.go`: `Wire` type and constructor.
- `accessors.go`: getters and state accessors.
- `breaker.go`: circuit breaker polling loop.
- `config.go`: setters for configuration-time state.
- `connect.go`: component attachment helpers.
- `error.go`: error routing.
- `internal.go`: worker loops and fast-path processing.
- `lifecycle.go`: start/stop/restart.
- `load.go`: load helpers.
- `options.go`: functional options.
- `resister.go`: resister queue processing.
- `routing.go`: diversion and output routing.
- `submit.go`: submit paths.
- `telemetry*.go`: logging and sensor hooks.
- `transform.go`: transform and retry helpers.
- `wire_test.go`: tests.

## Notes

- The wire is designed for predictable lifecycle behavior. Mutating configuration while running is unsupported.
- Telemetry is best-effort and should not backpressure the pipeline.
- If you add a new capability, ensure it is configuration-time, add tests, and keep hot-path allocations in check.
