package types

import (
	"context"
	"crypto/tls"
	"time"

	"github.com/segmentio/kafka-go/sasl"
)

// KafkaSecurity bundles TLS + SASL + ClientID for kafka-go.
type KafkaSecurity struct {
	SASL      sasl.Mechanism // nil => no SASL
	TLS       *tls.Config    // nil => PLAINTEXT
	ClientID  string         // optional
	DialerTO  time.Duration  // optional (defaults 10s)
	DualStack bool           // optional (defaults true)
}

////////////////////////////////////////////////////////////////////////////////
// Common Kafka client wiring (no envs)
////////////////////////////////////////////////////////////////////////////////

// pkg/internal/types/types.go  (section: KafkaClientDeps)

// pkg/internal/types/kafka.go (or wherever your types live)

type KafkaClientDeps struct {
	// Connection
	Brokers []string // e.g., []{"broker-1:9092","broker-2:9092"}

	// Driver-specific handles (optional; one or both may be set depending on mode)
	Producer any
	Consumer any

	// Optional dead-letter topic for writer-side DLQ emission.
	DLQTopic string

	// Optional security bundle (TLS + SASL + client id, dialer prefs)
	Security *KafkaSecurity // <--- add this line
}

////////////////////////////////////////////////////////////////////////////////
// Writer config (format + batching + routing)
////////////////////////////////////////////////////////////////////////////////

type KafkaWriterConfig struct {
	// Destination
	Topic string // required

	// Pluggable value format (encode) — examples: "ndjson", "json", "parquet", "raw"
	// - record-oriented encoders -> ServeWriter(ctx, <-chan T)
	// - raw bytes passthrough     -> ServeWriterRaw(ctx, <-chan []byte)
	Format        string
	FormatOptions map[string]string // e.g. {"gzip":"true"} or {"parquet_compression":"zstd"}

	// Legacy compression hint (producer-level), e.g. "gzip","snappy","zstd",""
	Compression string

	// Routing / record adornments
	// Key/Header templates are evaluated per record (simple template expansion recommended).
	KeyTemplate     string            // e.g. "{customerId}" (value serialized to bytes by adapter)
	HeaderTemplates map[string]string // e.g. {"source":"{service}", "schema":"v1"}

	// Batching thresholds (prior to Produce/flush)
	BatchMaxRecords int           // default 50_000
	BatchMaxBytes   int           // default 4<<20 (driver may impose tighter limits)
	BatchMaxAge     time.Duration // default 1s..5s typical; we’ll default 1s in impl

	// Producer delivery semantics
	Acks              string        // "", "0", "1", "all"
	RequestTimeout    time.Duration // network/request timeout
	PartitionStrategy string        // "", "hash", "round_robin", "manual"
	ManualPartition   *int          // used when PartitionStrategy=="manual"

	// DLQ behavior (optional)
	EnableDLQ bool // if true and DLQTopic in deps is set, failed messages are sent there
}

////////////////////////////////////////////////////////////////////////////////
// Reader config (subscription + decode + commits)
////////////////////////////////////////////////////////////////////////////////

type KafkaReaderConfig struct {
	// Subscription
	GroupID string   // required for consumer groups
	Topics  []string // one or more topics

	// Start position
	// StartAt: "latest" | "earliest" | "timestamp"
	StartAt     string
	StartAtTime time.Time // used when StartAt == "timestamp"

	// Fetch/poll knobs
	PollInterval   time.Duration // optional; if zero, impl picks a sane default
	MaxPollRecords int           // per poll/decode batch
	MaxPollBytes   int           // per fetch limit (driver permitting)

	// Decode
	Format        string            // "ndjson","json","parquet","raw"
	FormatOptions map[string]string // e.g. {"columns":"id,name"}

	// Legacy field; only used for gzip-compressed NDJSON.
	Compression string

	// Commit strategy
	// "auto" => driver-managed; "manual" => adapter commits after submit success
	CommitMode string // "auto" | "manual"
	// When CommitMode=="manual", choose when to commit:
	// "after-each" | "after-batch" | "time" (every CommitInterval)
	CommitPolicy   string
	CommitInterval time.Duration
}

////////////////////////////////////////////////////////////////////////////////
// Writer (record-oriented) adapter
////////////////////////////////////////////////////////////////////////////////

type KafkaWriterAdapter[T any] interface {
	// Dependency/config injection
	SetKafkaClientDeps(KafkaClientDeps)
	SetWriterConfig(KafkaWriterConfig)

	// Wire fan-in
	ConnectInput(...Wire[T])

	// Lifecycle (record-oriented encode → produce)
	Serve(ctx context.Context, in <-chan T) error
	// Start consuming from previously connected wires
	StartWriter(ctx context.Context) error

	// Raw bytes passthrough (one []byte -> one Kafka message)
	ServeWriterRaw(ctx context.Context, in <-chan []byte) error

	Stop()

	// Introspection / hooks
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Reader (record-oriented) adapter
////////////////////////////////////////////////////////////////////////////////

type KafkaReaderAdapter[T any] interface {
	SetKafkaClientDeps(KafkaClientDeps)
	SetReaderConfig(KafkaReaderConfig)

	// Serve polls Kafka -> decodes -> submit(T)
	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Stop()

	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Unified Kafka client adapter (read + write)
////////////////////////////////////////////////////////////////////////////////

type KafkaClientAdapter[T any] interface {
	// deps/config
	SetKafkaClientDeps(KafkaClientDeps)
	SetWriterConfig(KafkaWriterConfig)
	SetReaderConfig(KafkaReaderConfig)

	// --- WRITER ---
	ServeWriter(ctx context.Context, in <-chan T) error
	ServeWriterRaw(ctx context.Context, in <-chan []byte) error
	ConnectInput(...Wire[T])
	StartWriter(ctx context.Context) error

	// --- READER ---
	Serve(ctx context.Context, submit func(context.Context, T) error) error
	Fetch() (HttpResponse[[]T], error) // optional one-shot fetch (impl-defined)

	// lifecycle/introspection
	Stop()
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	Name() string
}

////////////////////////////////////////////////////////////////////////////////
// Options (builder-style)
////////////////////////////////////////////////////////////////////////////////

type KafkaWriterOption[T any] func(KafkaWriterAdapter[T])
type KafkaReaderOption[T any] func(KafkaReaderAdapter[T])
type KafkaClientOption[T any] func(KafkaClientAdapter[T])
