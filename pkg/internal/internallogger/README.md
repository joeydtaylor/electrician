# internallogger

The internallogger package provides the structured logging backend used by internal components. It wraps a concrete logger implementation and exposes a consistent interface defined in types/logger.go.

## Responsibilities

- Provide structured logging with levels and sinks.
- Support development and production output formats.
- Allow components to emit consistent log events.

## Key types and functions

- Logger: main implementation.
- Sink configuration (stdout, file, network).
- LogLevel constants (debug, info, warn, error, panic, fatal).

## Configuration

Loggers are configured through builder options:

- Level and development mode
- Sinks and sink configuration

## Behavior

Internal components call NotifyLoggers with a log level. The logger decides whether to emit based on its configured level.

## Usage

```go
logger := builder.NewLogger(
    builder.LoggerWithLevel("info"),
    builder.LoggerWithDevelopment(true),
)
```

## References

- builder: pkg/builder/logger.go
- internal contracts: pkg/internal/types/logger.go
