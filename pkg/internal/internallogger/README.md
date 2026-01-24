# Internal Logger

Package `internallogger` provides Electrician's default logger implementation using zap. External consumers should use the builder APIs under `pkg/builder`.

The logger exposes the internal `types.Logger` contract, supports multiple sinks, and handles level mapping between Electrician and zap.

## Configuration

Construct a logger with optional settings:

```go
logger := internallogger.NewLogger(
	internallogger.LoggerWithLevel("info"),
	internallogger.LoggerWithDevelopment(true),
)
```

Sinks can be added or removed at runtime:

```go
_ = logger.AddSink("stdout", types.SinkConfig{Type: "stdout"})
_ = logger.AddSink("file", types.SinkConfig{Type: "file", Config: map[string]interface{}{"path": "/tmp/app.log"}})
```

## Behavior

- Log levels are mapped to zap levels with `ConvertLevel`.
- `SetLevel` updates the atomic level used by all sinks.
- `AddSink` rebuilds the logger to include the new destination.
- `RemoveSink` rebuilds the logger to exclude the removed destination.

## Package layout

- `logger.go`: adapter type and constructor.
- `levels.go`: level parsing and conversions.
- `log.go`: logging and level accessors.
- `options.go`: functional options.
- `sinks.go`: sink management.
- `*_test.go`: tests.
