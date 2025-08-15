// pkg/builder/kafkaclient_adapter.go
package builder

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"

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

// --- ADD TO pkg/builder/kafkaclient_adapter.go ---
// Required imports to add alongside existing ones:
//   "crypto/tls"
//   "crypto/x509"
//   "os"
//   "path/filepath"
//   "github.com/segmentio/kafka-go/sasl"
//   "github.com/segmentio/kafka-go/sasl/scram"

// -------------------------------------------------
// Security helpers (TLS + SASL) to reduce boilerplate
// -------------------------------------------------

// TLSFromCAFilesStrict loads a strict TLS config (Min TLS1.2) using the first
// existing file path from candidates. If serverName != "", it is set for SNI
// and hostname verification.
func TLSFromCAFilesStrict(candidates []string, serverName string) (*tls.Config, error) {
	var picked string
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			picked = p
			break
		}
	}
	if picked == "" {
		return nil, fmt.Errorf("no CA file found in candidates: %v", candidates)
	}
	pem, err := os.ReadFile(filepath.Clean(picked))
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("invalid CA PEM at %s", picked)
	}
	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    cp,
	}
	if serverName != "" {
		cfg.ServerName = serverName
	}
	return cfg, nil
}

// TLSFromCAPathCSV convenience wrapper around TLSFromCAFilesStrict.
func TLSFromCAPathCSV(csv, serverName string) (*tls.Config, error) {
	var paths []string
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			paths = append(paths, p)
		}
	}
	return TLSFromCAFilesStrict(paths, serverName)
}

// SASLSCRAM returns a sasl.Mechanism for kafka-go from a common name.
// Supported: "SCRAM-SHA-256" (default), "SCRAM-SHA-512".
func SASLSCRAM(user, pass, mech string) (sasl.Mechanism, error) {
	switch strings.ToUpper(strings.ReplaceAll(mech, "_", "-")) {
	case "", "SCRAM-SHA-256":
		return scram.Mechanism(scram.SHA256, user, pass)
	case "SCRAM-SHA-512":
		return scram.Mechanism(scram.SHA512, user, pass)
	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s", mech)
	}
}

// NewKafkaGoTransport builds a kafka-go Transport with optional TLS/SASL/ClientID.
func NewKafkaGoTransport(tlsCfg *tls.Config, mech sasl.Mechanism, clientID string) *kafka.Transport {
	return &kafka.Transport{
		TLS:      tlsCfg,
		SASL:     mech,
		ClientID: clientID,
	}
}

// NewKafkaGoDialer builds a kafka-go Dialer with optional TLS/SASL.
// If timeout is zero, 10s is used. DualStack is set as provided.
func NewKafkaGoDialer(tlsCfg *tls.Config, mech sasl.Mechanism, timeout time.Duration, dualStack bool) *kafka.Dialer {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &kafka.Dialer{
		Timeout:       timeout,
		DualStack:     dualStack,
		SASLMechanism: mech,
		TLS:           tlsCfg,
	}
}

// Option helpers to attach Transport/Dialer to the convenience constructors below.
func KafkaGoWriterWithTransport(t *kafka.Transport) KafkaGoWriterOption {
	return func(w *kafka.Writer) { w.Transport = t }
}
func KafkaGoReaderWithDialer(d *kafka.Dialer) KafkaGoReaderOption {
	return func(c *kafka.ReaderConfig) { c.Dialer = d }
}

// NewKafkaGoWriterSecure: NewKafkaGoWriter + Transport(TLS/SASL) in one call.
func NewKafkaGoWriterSecure(brokers []string, topic string, tlsCfg *tls.Config, mech sasl.Mechanism, clientID string, opts ...KafkaGoWriterOption) *kafka.Writer {
	transport := NewKafkaGoTransport(tlsCfg, mech, clientID)
	opts = append([]KafkaGoWriterOption{KafkaGoWriterWithTransport(transport)}, opts...)
	return NewKafkaGoWriter(brokers, topic, opts...)
}

// NewKafkaGoReaderSecure: NewKafkaGoReader + Dialer(TLS/SASL) in one call.
// timeout=0 => 10s. dualStack=true is typical for local dev.
func NewKafkaGoReaderSecure(brokers []string, group string, topics []string, tlsCfg *tls.Config, mech sasl.Mechanism, timeout time.Duration, dualStack bool, opts ...KafkaGoReaderOption) *kafka.Reader {
	dialer := NewKafkaGoDialer(tlsCfg, mech, timeout, dualStack)
	opts = append([]KafkaGoReaderOption{KafkaGoReaderWithDialer(dialer)}, opts...)
	return NewKafkaGoReader(brokers, group, topics, opts...)
}

// -------------------------------------------------
// Builder for types.KafkaSecurity + adapter option
// -------------------------------------------------

// KafkaSecurityOption mutates a types.KafkaSecurity.
type KafkaSecurityOption func(*types.KafkaSecurity)

// NewKafkaSecurity creates a types.KafkaSecurity with sensible defaults.
// Defaults: DialerTO=10s, DualStack=true. TLS/SASL/ClientID are optional.
func NewKafkaSecurity(opts ...KafkaSecurityOption) *types.KafkaSecurity {
	sec := &types.KafkaSecurity{
		DialerTO:  10 * time.Second,
		DualStack: true,
	}
	for _, o := range opts {
		o(sec)
	}
	return sec
}

func WithTLS(cfg *tls.Config) KafkaSecurityOption {
	return func(s *types.KafkaSecurity) { s.TLS = cfg }
}
func WithSASL(mech sasl.Mechanism) KafkaSecurityOption {
	return func(s *types.KafkaSecurity) { s.SASL = mech }
}
func WithClientID(id string) KafkaSecurityOption {
	return func(s *types.KafkaSecurity) { s.ClientID = id }
}
func WithDialer(timeout time.Duration, dualStack bool) KafkaSecurityOption {
	return func(s *types.KafkaSecurity) {
		if timeout > 0 {
			s.DialerTO = timeout
		}
		s.DualStack = dualStack
	}
}

// NewKafkaGoTransportFromSecurity maps types.KafkaSecurity -> kafka.Transport.
func NewKafkaGoTransportFromSecurity(sec *types.KafkaSecurity) *kafka.Transport {
	if sec == nil {
		return nil
	}
	return NewKafkaGoTransport(sec.TLS, sec.SASL, sec.ClientID)
}

// NewKafkaGoDialerFromSecurity maps types.KafkaSecurity -> kafka.Dialer.
func NewKafkaGoDialerFromSecurity(sec *types.KafkaSecurity) *kafka.Dialer {
	if sec == nil {
		return nil
	}
	timeout := sec.DialerTO
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return NewKafkaGoDialer(sec.TLS, sec.SASL, timeout, sec.DualStack)
}

// NewKafkaGoWriterWithSecurity wires a writer and applies security.Transport.
func NewKafkaGoWriterWithSecurity(brokers []string, topic string, sec *types.KafkaSecurity, opts ...KafkaGoWriterOption) *kafka.Writer {
	if sec != nil {
		opts = append([]KafkaGoWriterOption{KafkaGoWriterWithTransport(NewKafkaGoTransportFromSecurity(sec))}, opts...)
	}
	return NewKafkaGoWriter(brokers, topic, opts...)
}

// NewKafkaGoReaderWithSecurity wires a reader and applies security.Dialer.
func NewKafkaGoReaderWithSecurity(brokers []string, group string, topics []string, sec *types.KafkaSecurity, opts ...KafkaGoReaderOption) *kafka.Reader {
	if sec != nil {
		opts = append([]KafkaGoReaderOption{KafkaGoReaderWithDialer(NewKafkaGoDialerFromSecurity(sec))}, opts...)
	}
	return NewKafkaGoReader(brokers, group, topics, opts...)
}

// KafkaClientAdapterWithConnection sets Brokers + Security on the adapter deps.
// Use this when you want the internal adapter to project TLS/SASL itself.
func KafkaClientAdapterWithConnection[T any](brokers []string, sec *types.KafkaSecurity) types.KafkaClientOption[T] {
	return kafkaClientAdapter.WithKafkaClientDeps[T](types.KafkaClientDeps{
		Brokers:  brokers,
		Security: sec,
	})
}
