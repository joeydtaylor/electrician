# adapter/kafkaclient

The kafkaclient adapter integrates with Kafka using kafka-go. It provides reader and writer adapters with structured configuration and telemetry.

## Responsibilities

- Configure Kafka readers and writers.
- Serialize and deserialize payloads (JSON, NDJSON, Parquet when applicable).
- Emit telemetry for connection, batch, and produce/consume events.

## Key types and functions

- KafkaClientAdapter[T]: main adapter type.
- Serve(): consume messages and submit into the pipeline.
- Write(): publish messages to Kafka topics.

## Configuration

Common options include:

- Brokers, topics, group IDs
- Start position (earliest/latest)
- Commit policy and intervals
- Batch size and flush behavior
- TLS and SASL configuration
- Sensor and logger

## Observability

Sensors emit metrics for consume/produce success, failures, and batch flushes. Loggers capture connection and runtime events.

## Usage

```go
reader := builder.NewKafkaClientAdapter[Item](
    ctx,
    builder.KafkaClientAdapterWithReaderTopics[Item]("topic"),
    builder.KafkaClientAdapterWithReaderGroup[Item]("group"),
    builder.KafkaClientAdapterWithSensor[Item](sensor),
)
```

## References

- examples: example/adapter/kafka_adapter/
- builder: pkg/builder/kafkaclient_adapter.go
- internal contracts: pkg/internal/types/kafka_adapter.go
