package kafkaclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/segmentio/kafka-go"
)

func (a *KafkaClient[T]) forEachSensor(fn func(types.Sensor[T])) {
	a.sensorLock.Lock()
	local := make([]types.Sensor[T], len(a.sensors))
	copy(local, a.sensors)
	a.sensorLock.Unlock()
	for _, s := range local {
		if s != nil {
			fn(s)
		}
	}
}

// small helper to merge wire outputs
func (a *KafkaClient[T]) fanIn(ctx context.Context, dst chan<- T, src <-chan T) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-src:
			if !ok {
				return
			}
			dst <- v
		}
	}
}

func (a *KafkaClient[T]) produce(
	ctx context.Context,
	msgTopic string, // "" means "use writer.Topic if available"
	manualPartition *int,
	key []byte,
	val []byte,
	headers []struct{ Key, Value string },
) (partition int, offset int64, err error) {
	switch w := a.producer.(type) {
	case *kafka.Writer:
		msg := kafka.Message{
			Key:   key,
			Value: val,
		}
		// only set Topic on the message if the writer doesn't already carry one
		if strings.TrimSpace(w.Topic) == "" && strings.TrimSpace(msgTopic) != "" {
			msg.Topic = strings.TrimSpace(msgTopic)
		}
		if manualPartition != nil {
			msg.Partition = *manualPartition
		}
		if n := len(headers); n > 0 {
			msg.Headers = make([]kafka.Header, 0, n)
			for _, h := range headers {
				msg.Headers = append(msg.Headers, kafka.Header{Key: h.Key, Value: []byte(h.Value)})
			}
		}

		if err := w.WriteMessages(ctx, msg); err != nil {
			return -1, -1, err
		}
		// kafka-go Writer doesn’t return offset; we return -1.
		return int(msg.Partition), -1, nil

	default:
		return -1, -1, fmt.Errorf("kafkaclient: unsupported producer type %T (expected *kafka.Writer)", a.producer)
	}
}

// effectiveWriterTopic returns a usable topic if either the adapter has one
// configured or the injected *kafka.Writer carries one. If neither is set,
// ok == false.
func (a *KafkaClient[T]) effectiveWriterTopic() (topic string, ok bool) {
	if t := strings.TrimSpace(a.wTopic); t != "" {
		return t, true
	}
	if w, is := a.producer.(*kafka.Writer); is {
		if t := strings.TrimSpace(w.Topic); t != "" {
			// Writer already has a topic; in this case we must NOT also set Message.Topic
			return t, true
		}
	}
	return "", false
}

// very small "{field}" renderer for structs marshalled to map[string]any
func renderFieldFromValue[T any](v T, placeholder string) (string, bool) {
	ph := strings.TrimSpace(placeholder)
	if !strings.HasPrefix(ph, "{") || !strings.HasSuffix(ph, "}") {
		return "", false
	}
	field := strings.TrimSpace(ph[1 : len(ph)-1])
	if field == "" {
		return "", false
	}
	// JSON round-trip to map for simplicity
	b, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return "", false
	}
	if raw, ok := m[field]; ok && raw != nil {
		return fmt.Sprint(raw), true
	}
	return "", false
}

func renderKeyFromTemplate[T any](tmpl string, v T) []byte {
	t := strings.TrimSpace(tmpl)
	if t == "" {
		return nil
	}
	if val, ok := renderFieldFromValue(v, t); ok {
		return []byte(val)
	}
	// non-placeholder templates could be supported here; for now return literal
	return []byte(t)
}

func renderHeadersFromTemplates[T any](tmpls map[string]string, v T) []struct{ Key, Value string } {
	if len(tmpls) == 0 {
		return nil
	}
	out := make([]struct{ Key, Value string }, 0, len(tmpls))
	for k, t := range tmpls {
		val := t
		if vv, ok := renderFieldFromValue(v, t); ok {
			val = vv
		}
		out = append(out, struct{ Key, Value string }{Key: k, Value: val})
	}
	return out
}

// -------- reader helpers (kafka-go) --------

func (a *KafkaClient[T]) getOrCreateReaderForServe() (*kafka.Reader, bool, error) {
	if r, ok := a.consumer.(*kafka.Reader); ok && r != nil {
		return r, false, nil
	}

	brokers := append([]string(nil), a.brokers...)
	if len(brokers) == 0 {
		// friendly default to match writer behavior
		brokers = []string{"127.0.0.1:19092"}
	}

	cfg := kafka.ReaderConfig{
		Brokers: brokers,
	}

	// Subscription: group + topics (multi) or single topic without group
	if a.rGroupID != "" {
		cfg.GroupID = a.rGroupID
		if len(a.rTopics) == 1 {
			cfg.Topic = a.rTopics[0]
		} else if len(a.rTopics) > 1 {
			cfg.GroupTopics = append([]string(nil), a.rTopics...)
		} else {
			return nil, false, fmt.Errorf("kafkaclient: reader requires at least one topic")
		}
	} else {
		// no group => single topic reader
		if len(a.rTopics) != 1 {
			return nil, false, fmt.Errorf("kafkaclient: reader without GroupID must specify exactly one topic")
		}
		cfg.Topic = a.rTopics[0]
	}

	// Start position (fallback for partitions with no committed offsets)
	switch strings.ToLower(a.rStartAt) {
	case "earliest":
		cfg.StartOffset = kafka.FirstOffset
	case "latest", "":
		cfg.StartOffset = kafka.LastOffset
	case "timestamp":
		// handled best-effort below (after create); keep default for now
	default:
		cfg.StartOffset = kafka.LastOffset
	}

	// Polling / sizing
	if a.rPollInterval > 0 {
		cfg.MaxWait = a.rPollInterval
	}
	if a.rMaxPollBytes > 0 {
		cfg.MaxBytes = a.rMaxPollBytes
	} else {
		cfg.MaxBytes = 1_000_000 // 1 MiB sane default
	}
	// Let kafka-go choose MinBytes; QueueCapacity left default.

	// Commit strategy
	mode := strings.ToLower(strings.TrimSpace(a.rCommitMode))
	if mode == "" || mode == "auto" {
		// auto-commit at interval
		if a.rCommitEvery > 0 {
			cfg.CommitInterval = a.rCommitEvery
		} else {
			cfg.CommitInterval = 3 * time.Second
		}
	} else {
		// manual commit (we call CommitMessages)
		cfg.CommitInterval = 0
	}

	r := kafka.NewReader(cfg)

	// Optional timestamp start (best-effort; works when not using consumer group or when no offsets exist)
	if strings.EqualFold(a.rStartAt, "timestamp") && !a.rStartAtTime.IsZero() {
		// SetOffsetAt applies to the reader's assigned partitions; for consumer groups, assignment may occur lazily.
		// We'll try once; if it fails, we just continue (committed offsets / latest will apply).
		_ = r.SetOffsetAt(context.Background(), a.rStartAtTime)
	}

	return r, true, nil
}

func (a *KafkaClient[T]) commitLatest(
	ctx context.Context,
	r *kafka.Reader,
	latest map[string]kafka.Message, // key: "topic:partition"
	groupID string,
) {
	if len(latest) == 0 {
		return
	}
	msgs := make([]kafka.Message, 0, len(latest))
	offsets := make(map[string]int64, len(latest))
	for k, m := range latest {
		msgs = append(msgs, m)
		// CommitMessages commits the *next* offset; we report that in the sensor for clarity.
		offsets[k] = m.Offset + 1
	}
	if err := r.CommitMessages(ctx, msgs...); err != nil {
		a.forEachSensor(func(s types.Sensor[T]) { s.InvokeOnKafkaCommitError(a.componentMetadata, groupID, err) })
		a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Commit, err: %v", a.componentMetadata, err)
		return
	}
	a.forEachSensor(func(s types.Sensor[T]) { s.InvokeOnKafkaCommitSuccess(a.componentMetadata, groupID, offsets) })
}
