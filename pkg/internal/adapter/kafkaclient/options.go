// pkg/internal/adapter/kafkaclient/options.go
package kafkaclient

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithKafkaClientDeps injects producer/consumer deps and cluster info.
func WithKafkaClientDeps[T any](deps types.KafkaClientDeps) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetKafkaClientDeps(deps)
	}
}

// WithWriterConfig sets the Kafka writer configuration wholesale.
func WithWriterConfig[T any](cfg types.KafkaWriterConfig) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(cfg)
	}
}

// WithReaderConfig sets the Kafka reader configuration wholesale.
func WithReaderConfig[T any](cfg types.KafkaReaderConfig) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(cfg)
	}
}

// WithSensor attaches sensors to the adapter.
func WithSensor[T any](sensor ...types.Sensor[T]) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.ConnectSensor(sensor...)
	}
}

// WithLogger attaches loggers to the adapter.
func WithLogger[T any](l ...types.Logger) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.ConnectLogger(l...)
	}
}

// WithWire connects one or more Wire[T] as inputs to the Kafka writer.
func WithWire[T any](wires ...types.Wire[T]) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.ConnectInput(wires...)
	}
}

//
// ---------- Writer helpers ----------
//

// WithWriterTopic sets the output topic.
func WithWriterTopic[T any](topic string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{Topic: topic})
	}
}

// WithWriterFormat sets writer format + optional compression hint (driver/codec-specific).
func WithWriterFormat[T any](format, compression string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{
			Format:      format,
			Compression: compression,
		})
	}
}

// WithWriterFormatOptions merges format/encoder options (e.g., parquet knobs, header templating switches).
func WithWriterFormatOptions[T any](opts map[string]string) types.KafkaClientOption[T] {
	cp := make(map[string]string, len(opts))
	for k, v := range opts {
		cp[k] = v
	}
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{FormatOptions: cp})
	}
}

// WithWriterBatchSettings overrides writer batch thresholds.
func WithWriterBatchSettings[T any](maxRecords, maxBytes int, maxAge time.Duration) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{
			BatchMaxRecords: maxRecords,
			BatchMaxBytes:   maxBytes,
			BatchMaxAge:     maxAge,
		})
	}
}

// WithWriterAcks sets producer acks ("0","1","all").
func WithWriterAcks[T any](acks string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{Acks: acks})
	}
}

// WithWriterRequestTimeout sets the producer request timeout.
func WithWriterRequestTimeout[T any](d time.Duration) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{RequestTimeout: d})
	}
}

// WithWriterPartitionStrategy sets partitioner ("hash","round_robin","manual", etc. depending on driver).
func WithWriterPartitionStrategy[T any](strategy string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{PartitionStrategy: strategy})
	}
}

// WithWriterManualPartition pins messages to a specific partition (enables manual strategy).
func WithWriterManualPartition[T any](p int) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		pp := p
		adp.SetWriterConfig(types.KafkaWriterConfig{ManualPartition: &pp})
	}
}

// WithWriterDLQ enables/disables DLQ for failed produce attempts.
func WithWriterDLQ[T any](enable bool) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{EnableDLQ: enable})
	}
}

// WithWriterKeyTemplate sets a templated key (for hashing/partitioning).
func WithWriterKeyTemplate[T any](tmpl string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{KeyTemplate: tmpl})
	}
}

// WithWriterHeaderTemplates sets templated headers (k->template).
func WithWriterHeaderTemplates[T any](hdrs map[string]string) types.KafkaClientOption[T] {
	cp := make(map[string]string, len(hdrs))
	for k, v := range hdrs {
		cp[k] = v
	}
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetWriterConfig(types.KafkaWriterConfig{HeaderTemplates: cp})
	}
}

//
// ---------- Reader helpers ----------
//

// WithReaderGroup sets the consumer group id.
func WithReaderGroup[T any](groupID string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(types.KafkaReaderConfig{GroupID: groupID})
	}
}

// WithReaderTopics sets the subscribed topics.
func WithReaderTopics[T any](topics ...string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(types.KafkaReaderConfig{Topics: topics})
	}
}

// WithReaderStartAt selects the starting position ("latest","earliest","timestamp") and optional time.
func WithReaderStartAt[T any](mode string, ts time.Time) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(types.KafkaReaderConfig{
			StartAt:     mode,
			StartAtTime: ts,
		})
	}
}

// WithReaderPollSettings sets poll cadence and limits.
func WithReaderPollSettings[T any](interval time.Duration, maxRecords, maxBytes int) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(types.KafkaReaderConfig{
			PollInterval:   interval,
			MaxPollRecords: maxRecords,
			MaxPollBytes:   maxBytes,
		})
	}
}

// WithReaderFormat sets the READER format (e.g., "ndjson","parquet") and optional compression hint.
func WithReaderFormat[T any](format, compression string) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(types.KafkaReaderConfig{
			Format:      format,
			Compression: compression,
		})
	}
}

// WithReaderFormatOptions merges reader format-specific knobs.
func WithReaderFormatOptions[T any](opts map[string]string) types.KafkaClientOption[T] {
	cp := make(map[string]string, len(opts))
	for k, v := range opts {
		cp[k] = v
	}
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(types.KafkaReaderConfig{
			FormatOptions: cp,
		})
	}
}

// WithReaderCommit configures commit mode/policy/interval.
func WithReaderCommit[T any](mode, policy string, every time.Duration) types.KafkaClientOption[T] {
	return func(adp types.KafkaClientAdapter[T]) {
		adp.SetReaderConfig(types.KafkaReaderConfig{
			CommitMode:     mode,
			CommitPolicy:   policy,
			CommitInterval: every,
		})
	}
}
