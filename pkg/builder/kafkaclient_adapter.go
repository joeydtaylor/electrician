// pkg/builder/kafkaclient_adapter.go
package builder

import (
	"context"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"

	kafkaClientAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/kafkaclient"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

////////////////////////
// Adapter constructor +
////////////////////////

// NewKafkaClientAdapter creates a new Kafka client adapter (read + write capable).
func NewKafkaClientAdapter[T any](ctx context.Context, options ...types.KafkaClientOption[T]) types.KafkaClientAdapter[T] {
	return kafkaClientAdapter.NewKafkaClientAdapter[T](ctx, options...)
}

func KafkaClientAdapterWithKafkaClientDeps[T any](deps types.KafkaClientDeps) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithKafkaClientDeps[T](deps)
}

func KafkaClientAdapterWithWriterConfig[T any](cfg types.KafkaWriterConfig) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterConfig[T](cfg)
}

func KafkaClientAdapterWithReaderConfig[T any](cfg types.KafkaReaderConfig) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderConfig[T](cfg)
}

func KafkaClientAdapterWithSensor[T any](sensor ...types.Sensor[T]) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithSensor[T](sensor...)
}

func KafkaClientAdapterWithLogger[T any](l ...types.Logger) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithLogger[T](l...)
}

func KafkaClientAdapterWithWire[T any](wires ...types.Wire[T]) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWire[T](wires...)
}

////////////////////////////////////
// Writer-side exported options
////////////////////////////////////

func KafkaClientAdapterWithWriterTopic[T any](topic string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterTopic[T](topic)
}

func KafkaClientAdapterWithWriterFormat[T any](format, compression string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterFormat[T](format, compression)
}

func KafkaClientAdapterWithWriterFormatOptions[T any](opts map[string]string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterFormatOptions[T](opts)
}

func KafkaClientAdapterWithWriterBatchSettings[T any](maxRecords, maxBytes int, maxAge time.Duration) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterBatchSettings[T](maxRecords, maxBytes, maxAge)
}

func KafkaClientAdapterWithWriterAcks[T any](acks string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterAcks[T](acks)
}

func KafkaClientAdapterWithWriterRequestTimeout[T any](d time.Duration) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterRequestTimeout[T](d)
}

func KafkaClientAdapterWithWriterPartitionStrategy[T any](strategy string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterPartitionStrategy[T](strategy)
}

func KafkaClientAdapterWithWriterManualPartition[T any](p int) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterManualPartition[T](p)
}

func KafkaClientAdapterWithWriterDLQ[T any](enable bool) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterDLQ[T](enable)
}

func KafkaClientAdapterWithWriterKeyTemplate[T any](tmpl string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterKeyTemplate[T](tmpl)
}

func KafkaClientAdapterWithWriterHeaderTemplates[T any](hdrs map[string]string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithWriterHeaderTemplates[T](hdrs)
}

////////////////////////////////////
// Reader-side exported options
////////////////////////////////////

func KafkaClientAdapterWithReaderGroup[T any](groupID string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderGroup[T](groupID)
}

func KafkaClientAdapterWithReaderTopics[T any](topics ...string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderTopics[T](topics...)
}

func KafkaClientAdapterWithReaderStartAt[T any](mode string, ts time.Time) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderStartAt[T](mode, ts)
}

func KafkaClientAdapterWithReaderPollSettings[T any](interval time.Duration, maxRecords, maxBytes int) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderPollSettings[T](interval, maxRecords, maxBytes)
}

func KafkaClientAdapterWithReaderFormat[T any](format, compression string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderFormat[T](format, compression)
}

func KafkaClientAdapterWithReaderFormatOptions[T any](opts map[string]string) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderFormatOptions[T](opts)
}

func KafkaClientAdapterWithReaderCommit[T any](mode, policy string, every time.Duration) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithReaderCommit[T](mode, policy, every)
}

/////////////////////////////////////////////////////////////
// Driver helpers: kafka-go writer/reader constructors + inject
/////////////////////////////////////////////////////////////

// Inject a kafka-go Writer as the producer.
func KafkaClientAdapterWithKafkaGoWriter[T any](w *kafka.Writer) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithKafkaClientDeps[T](types.KafkaClientDeps{Producer: w})
}

// Inject a kafka-go Reader as the consumer.
func KafkaClientAdapterWithKafkaGoReader[T any](r *kafka.Reader) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithKafkaClientDeps[T](types.KafkaClientDeps{Consumer: r})
}

// ---- kafka-go Writer convenience ----

type KafkaGoWriterOption func(*kafka.Writer)

// NewKafkaGoWriter builds a sensible kafka-go Writer for the given brokers/topic.
func NewKafkaGoWriter(brokers []string, topic string, opts ...KafkaGoWriterOption) *kafka.Writer {
	w := &kafka.Writer{
		Addr:                   kafka.TCP(brokers...),
		Topic:                  topic,
		Balancer:               &kafka.Hash{},
		BatchTimeout:           200 * time.Millisecond,
		BatchBytes:             int64(1 << 20), // kafka-go uses int64 for BatchBytes
		BatchSize:              1000,
		RequiredAcks:           kafka.RequireAll,
		Async:                  false,
		AllowAutoTopicCreation: true,
	}
	for _, o := range opts {
		o(w)
	}
	return w
}

func KafkaGoWriterWithRoundRobin() KafkaGoWriterOption {
	return func(w *kafka.Writer) { w.Balancer = &kafka.RoundRobin{} }
}
func KafkaGoWriterWithHash() KafkaGoWriterOption {
	return func(w *kafka.Writer) { w.Balancer = &kafka.Hash{} }
}
func KafkaGoWriterWithLeastBytes() KafkaGoWriterOption {
	return func(w *kafka.Writer) { w.Balancer = &kafka.LeastBytes{} }
}
func KafkaGoWriterWithBatchTimeout(d time.Duration) KafkaGoWriterOption {
	return func(w *kafka.Writer) { w.BatchTimeout = d }
}
func KafkaGoWriterWithBatchBytes(n int64) KafkaGoWriterOption { // int64 to match kafka-go
	return func(w *kafka.Writer) { w.BatchBytes = n }
}
func KafkaGoWriterWithBatchSize(n int) KafkaGoWriterOption {
	return func(w *kafka.Writer) { w.BatchSize = n }
}
func KafkaGoWriterWithAsync(async bool) KafkaGoWriterOption {
	return func(w *kafka.Writer) { w.Async = async }
}
func KafkaGoWriterWithRequiredAcks(mode string) KafkaGoWriterOption {
	return func(w *kafka.Writer) {
		switch strings.ToLower(mode) {
		case "0", "none":
			w.RequiredAcks = kafka.RequireNone
		case "1", "leader":
			w.RequiredAcks = kafka.RequireOne
		default: // "all", "-1"
			w.RequiredAcks = kafka.RequireAll
		}
	}
}

// ---- kafka-go Reader convenience ----

type KafkaGoReaderOption func(*kafka.ReaderConfig)

// NewKafkaGoReader builds a kafka-go Reader for a consumer group over one or more topics.
func NewKafkaGoReader(brokers []string, group string, topics []string, opts ...KafkaGoReaderOption) *kafka.Reader {
	cfg := kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        group,
		StartOffset:    kafka.LastOffset, // "latest"
		MinBytes:       1 << 10,          // 1 KiB
		MaxBytes:       10 << 20,         // 10 MiB
		MaxWait:        500 * time.Millisecond,
		CommitInterval: 1 * time.Second, // driver auto-commit cadence
	}
	// Single vs multi-topic wiring (kafka-go uses Topic OR GroupTopics)
	if len(topics) == 1 {
		cfg.Topic = topics[0]
	} else if len(topics) > 1 {
		cfg.GroupTopics = topics
	}
	for _, o := range opts {
		o(&cfg)
	}
	return kafka.NewReader(cfg)
}

func KafkaGoReaderWithEarliestStart() KafkaGoReaderOption {
	return func(c *kafka.ReaderConfig) { c.StartOffset = kafka.FirstOffset }
}
func KafkaGoReaderWithLatestStart() KafkaGoReaderOption {
	return func(c *kafka.ReaderConfig) { c.StartOffset = kafka.LastOffset }
}
func KafkaGoReaderWithMinBytes(n int) KafkaGoReaderOption {
	return func(c *kafka.ReaderConfig) { c.MinBytes = n }
}
func KafkaGoReaderWithMaxBytes(n int) KafkaGoReaderOption {
	return func(c *kafka.ReaderConfig) { c.MaxBytes = n }
}
func KafkaGoReaderWithMaxWait(d time.Duration) KafkaGoReaderOption {
	return func(c *kafka.ReaderConfig) { c.MaxWait = d }
}
func KafkaGoReaderWithCommitInterval(d time.Duration) KafkaGoReaderOption {
	return func(c *kafka.ReaderConfig) { c.CommitInterval = d }
}
