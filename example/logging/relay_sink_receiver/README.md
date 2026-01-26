# Log Relay Sink (gRPC)

This feature lets Electrician services ship structured JSON logs to a central receiver over gRPC.
It is designed to be easy to implement in other SDKs (Rust, etc.) so teams can share logs across stacks.

## What the Go sink sends

The Go relay sink connects as a gRPC client and streams log entries to `RelayService.StreamReceive`
from `proto/electrician_relay.proto`.

Each log entry is sent as a `WrappedPayload` inside a `RelayEnvelope`:

- `payload` is a single JSON object (one log record).
- `metadata.content_type` is `application/json`.
- `payload_type` is `electrician.log.v1`.
- `payload_encoding` is left unspecified.
- `metadata.headers["source"]` is `electrician-logger`.
- `metadata.headers["log_schema"]` is `electrician.log.v1`.
- `timestamp` is set to now (UTC).
- `id` is an RFC3339Nano string (used as a unique ID).
- No compression or encryption by default (performance/security options are nil).

The sink opens the stream with a `StreamOpen` and then sends log payloads. It expects acknowledgments
but will continue sending even if acks are delayed.

## Log schema (electrician.log.v1)

Logs are JSON objects with these reserved fields:

- `log_schema`: schema identifier (`electrician.log.v1`)
- `ts`: RFC3339Nano timestamp in UTC
- `level`: lowercase level string
- `msg`: log message
- `logger`, `caller`, `stack`: optional zap fields

Recommended fields for cross-team analytics:

- `component`: component metadata (type/id/name)
- `event`: event name
- `result`: result status
- `error`: error string
- `trace_id`, `span_id`: tracing IDs

Example record:

```json
{"log_schema":"electrician.log.v1","ts":"2025-01-27T04:15:22.531Z","level":"info","msg":"relay log sample","event":"log_emit","seq":42,"component":{"type":"WIRE","id":"abc123"}}
```

## Rust (or other SDK) interoperability

To receive logs from Go:

1. Implement the `RelayService` server from `proto/electrician_relay.proto`.
2. Handle `StreamReceive`:
   - Read the initial `StreamOpen` (optional but expected).
   - For each `RelayEnvelope.payload`, parse `payload` as a JSON object.
   - Send `StreamAcknowledgment` responses (per message or in batches).
3. Trust the schema fields at the JSON top level (`log_schema`, `ts`, `level`, `msg`, etc.).

If you want Rust to send logs to Go:

- Create `WrappedPayload` messages with the same metadata above and stream them to a Go receiver.
- Use `StreamReceive` for high throughput or `Receive` for a simpler unary call.

## Running the Go examples

Start the receiver:

```bash
go run example/logging/relay_sink_receiver/main.go
```

Send a burst of logs:

```bash
LOG_RELAY_COUNT=300 LOG_RELAY_INTERVAL=15ms go run example/logging/relay_sink_sender/main.go
```

Capture logs to JSONL:

```bash
LOG_RELAY_OUTPUT=logs/relay.jsonl LOG_RELAY_SAMPLE=10 go run example/logging/relay_sink_receiver/main.go
```

### TLS setup

By default the examples expect client/server certs. You can point them at the sample TLS assets:

Receiver:

```bash
TLS_CERT=example/relay_example/tls/server.crt \
TLS_KEY=example/relay_example/tls/server.key \
TLS_CA=example/relay_example/tls/ca.crt \
go run example/logging/relay_sink_receiver/main.go
```

Sender:

```bash
TLS_CERT=example/relay_example/tls/client.crt \
TLS_KEY=example/relay_example/tls/client.key \
TLS_CA=example/relay_example/tls/ca.crt \
go run example/logging/relay_sink_sender/main.go
```

## Useful environment variables

Receiver:

- `RX_ADDR` (default `localhost:50090`)
- `LOG_RELAY_DURATION` (e.g. `5s` to auto-stop)
- `LOG_RELAY_SAMPLE` (default `5`)
- `LOG_RELAY_DUMP` (`true` to print all)
- `LOG_RELAY_OUTPUT` (path to JSONL file)

Sender:

- `LOG_RELAY_ADDR` (default `localhost:50090`)
- `LOG_RELAY_COUNT` (default `250`)
- `LOG_RELAY_INTERVAL` (default `20ms`)
- `LOG_RELAY_DURATION` (overrides runtime)

