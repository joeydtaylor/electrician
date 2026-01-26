# internallogger

The internallogger package provides the structured logging backend used by internal components. It wraps a concrete logger implementation and exposes a consistent interface defined in types/logger.go.

## Responsibilities

- Provide structured logging with levels and sinks.
- Emit a standardized JSON log format suitable for aggregation.
- Allow components to emit consistent log events.

## Key types and functions

- Logger: main implementation.
- Sink configuration (stdout, file, relay).
- LogLevel constants (debug, info, warn, error, panic, fatal).

## Configuration

Loggers are configured through builder options:

- Level and development mode
- Optional default fields applied to every log line
- Sinks and sink configuration

## Behavior

Internal components call NotifyLoggers with a log level. The logger decides whether to emit based on its configured level.

## Usage

```go
logger := builder.NewLogger(
    builder.LoggerWithLevel("info"),
    builder.LoggerWithDevelopment(true),
    builder.LoggerWithFields(map[string]interface{}{
        "service": map[string]string{
            "name":        "payments",
            "environment": "prod",
            "version":     "1.2.3",
            "instance":    "payments-1",
        },
    }),
)
```

## Standard log schema (electrician.log.v1)

Logs are emitted as JSON objects with these reserved keys:

- `log_schema`: schema identifier (`electrician.log.v1`)
- `ts`: RFC3339Nano timestamp in UTC
- `level`: lowercase level string (debug/info/warn/error/dpanic/panic/fatal)
- `msg`: log message
- `logger`: logger name (optional)
- `caller`: file:line (optional)
- `stack`: stacktrace (optional)

Additional fields are emitted at the top level. Recommended keys:

- `component`: component metadata (type/id/name)
- `event`: event name
- `result`: result status
- `error`: error message
- `trace_id`, `span_id`: tracing IDs

Example:

```json
{"log_schema":"electrician.log.v1","ts":"2025-01-27T04:15:22.531Z","level":"info","msg":"Wire started","component":{"id":"abc123","type":"WIRE","name":""},"event":"Start","result":"SUCCESS","service":{"name":"payments","environment":"prod","version":"1.2.3","instance":"payments-1"}}
```

## Relay sink (gRPC)

The relay sink ships JSON log entries to a relay receiver over gRPC. Configure it with a sink config:

```go
relaySink := builder.SinkConfig{
    Type: string(builder.RelaySink),
    Config: map[string]interface{}{
        "targets": []string{"localhost:50090"},
        "queue_size": 2048,
        "submit_timeout": "2s",
        "drop_on_full": true,
        "static_headers": map[string]string{"x-tenant": "prod"},
        "tls": map[string]interface{}{
            "cert": "client.crt",
            "key": "client.key",
            "ca": "ca.crt",
            "server_name": "localhost",
            "min_version": "1.2",
            "max_version": "1.3",
        },
        "bearer_token_env": "LOG_RELAY_TOKEN",
        "auth_required": true,
    },
}
if err := logger.AddSink("relay", relaySink); err != nil {
    // handle error
}
```

Typed helper (optional):

```go
relaySink := builder.RelaySinkConfig{
    Targets:       []string{"localhost:50090"},
    QueueSize:     2048,
    SubmitTimeout: 2 * time.Second,
    DropOnFull:    builder.BoolPtr(true),
    TLS:           builder.NewTlsServerConfig(true, "client.crt", "client.key", "ca.crt", "localhost", 0, 0),
}.ToSinkConfig()
```

Relay sink config keys:

- `targets` (string or []string, required)
- `queue_size` (int, default 2048)
- `submit_timeout` (duration string or milliseconds, default 2s)
- `flush_timeout` (duration string or milliseconds, default 2s)
- `drop_on_full` (bool, default true)
- `static_headers` (map[string]string)
- `tls` (map: cert/key/ca/server_name/min_version/max_version)
- `bearer_token` or `bearer_token_env` (optional)
- `auth_required` (bool, default true)

## References

- builder: pkg/builder/logger.go
- internal contracts: pkg/internal/types/logger.go
