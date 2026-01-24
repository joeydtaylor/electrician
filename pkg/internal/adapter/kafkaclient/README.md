# Kafka Client Adapter

The Kafka Client Adapter produces and consumes Kafka messages using injected driver handles (for example, `kafka-go`).
It is typically wired through a `Plug` and driven by a `Generator`.

## Responsibilities

- Encode records as NDJSON and write them to Kafka.
- Decode Kafka messages into `T` and submit them to downstream handlers.
- Emit telemetry callbacks for writer/reader lifecycle and message events.
- Support batching and commit strategies for readers.

## Pipeline Composition

A common arrangement is:

```
Kafka Client Adapter -> Plug -> Generator -> Wire
```

## Configuration

Writer configuration includes:

- Topic, format, and format options.
- Batching thresholds.
- Optional key and header templates.

Reader configuration includes:

- Group and topic subscription.
- Start offset behavior.
- Polling limits.
- Commit mode and policy.

Driver objects (such as `*kafka.Writer` or `*kafka.Reader`) are injected via options, and security (TLS/SASL) is configured
on those driver objects.

## Telemetry

Sensors can subscribe to:

- Writer start/stop
- Batch flush
- Produce success/error
- Reader start/stop
- Message decode and commit

Loggers receive formatted messages from adapter operations.

## Package Layout

- `kafkaclient.go`: core type and constructor
- `config.go`: dependency and configuration setters
- `connect.go`: logger/sensor/wire wiring
- `helpers.go`: internal helpers
- `lifecycle.go`: stop handling
- `metadata.go`: component metadata accessors
- `notify.go`: logger emission
- `reader.go`: reader serve loop
- `reader_helpers.go`: reader construction and commit helpers
- `writer.go`: writer serve loop
- `writer_raw.go`: raw writer loop
- `writer_helpers.go`: writer helpers
- `templates.go`: key/header templating
- `options.go`: functional options
- `*_test.go`: tests

## References

- Kafka adapter examples: `example/adapter/kafka_adapter/`
