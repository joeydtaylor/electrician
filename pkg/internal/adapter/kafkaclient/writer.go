package kafkaclient

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/segmentio/kafka-go"
)

// StartWriter fans in connected wires and serves records to Kafka.
func (a *KafkaClient[T]) StartWriter(ctx context.Context) error {
	if a.wTopic == "" {
		return fmt.Errorf("kafkaclient: StartWriter requires Topic; call SetWriterConfig(...)")
	}
	if len(a.inputWires) == 0 {
		return fmt.Errorf("kafkaclient: StartWriter requires at least one connected wire; call ConnectInput(...)")
	}

	if a.mergedIn == nil {
		size := a.wBatchMaxRecords
		if size <= 0 {
			size = 1024
		}
		a.mergedIn = make(chan T, size)
	}

	for _, w := range a.inputWires {
		if w == nil {
			continue
		}
		out := w.GetOutputChannel()
		go a.fanIn(ctx, a.mergedIn, out)
	}

	go func() {
		_ = a.ServeWriter(ctx, a.mergedIn)
	}()
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: StartWriter, topic: %s, format: %s, wires: %d",
		a.componentMetadata, a.wTopic, a.wFormat, len(a.inputWires))

	return nil
}

// ServeWriter reads typed records from in and produces NDJSON messages to Kafka.
func (a *KafkaClient[T]) ServeWriter(ctx context.Context, in <-chan T) error {
	effTopic, ok := a.effectiveWriterTopic()
	if !ok {
		return fmt.Errorf("kafkaclient: ServeWriter requires topic (set KafkaWriterConfig.Topic or use a kafka-go Writer with Topic)")
	}

	format := strings.ToLower(strings.TrimSpace(a.wFormat))
	if format == "" {
		format = "ndjson"
	}
	if format != "ndjson" {
		return fmt.Errorf("kafkaclient: ServeWriter only supports ndjson (got %q)", format)
	}

	if !atomic.CompareAndSwapInt32(&a.isServingWriter, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServingWriter, 0)

	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnKafkaWriterStart(a.componentMetadata, effTopic, format)
	}
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: WriterStart, topic: %s, format: %s",
		a.componentMetadata, effTopic, format)
	defer func() {
		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnKafkaWriterStop(a.componentMetadata)
		}
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: WriterStop", a.componentMetadata)
	}()

	maxRecs := a.wBatchMaxRecords
	if maxRecs <= 0 {
		maxRecs = 1000
	}
	maxBytes := a.wBatchMaxBytes
	if maxBytes <= 0 {
		maxBytes = 1 << 20
	}
	maxAge := a.wBatchMaxAge
	if maxAge <= 0 {
		maxAge = time.Second
	}

	type msg struct {
		key     []byte
		val     []byte
		headers []struct{ Key, Value string }
	}
	pending := make([]msg, 0, maxRecs)
	var byteTally int

	tick := time.NewTicker(maxAge)
	defer tick.Stop()

	flush := func() error {
		if len(pending) == 0 {
			return nil
		}
		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnKafkaBatchFlush(a.componentMetadata, effTopic, len(pending), byteTally, a.wCompression)
		}
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: BatchFlush, topic: %s, records: %d, bytes: %d, compression: %s",
			a.componentMetadata, effTopic, len(pending), byteTally, a.wCompression)

		partForHook := -1
		if a.wManualPartition != nil {
			partForHook = *a.wManualPartition
		}

		for _, m := range pending {
			a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: ProduceAttempt, topic: %s, partition: %d, key_bytes: %d, val_bytes: %d",
				a.componentMetadata, effTopic, partForHook, len(m.key), len(m.val))

			topicForMessage := ""
			if _, has := a.producer.(*kafka.Writer); !has {
				topicForMessage = effTopic
			}

			partition, offset, err := a.produce(ctx, topicForMessage, a.wManualPartition, m.key, m.val, m.headers)
			if err != nil {
				for _, sensor := range a.snapshotSensors() {
					if sensor == nil {
						continue
					}
					sensor.InvokeOnKafkaProduceError(a.componentMetadata, effTopic, partForHook, err)
				}
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Produce, topic: %s, err: %v",
					a.componentMetadata, effTopic, err)
				return err
			}

			for _, sensor := range a.snapshotSensors() {
				if sensor == nil {
					continue
				}
				sensor.InvokeOnKafkaProduceSuccess(a.componentMetadata, effTopic, partition, offset, 0)
			}
			offsetStr := "n/a"
			if offset >= 0 {
				offsetStr = fmt.Sprintf("%d", offset)
			}
			a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ProduceSuccess, topic: %s, partition: %d, offset: %s",
				a.componentMetadata, effTopic, partition, offsetStr)
		}

		pending = pending[:0]
		byteTally = 0
		return nil
	}

	encode := func(v T) (key, val []byte, hdrs []struct{ Key, Value string }, err error) {
		val, err = json.Marshal(v)
		if err != nil {
			return nil, nil, nil, err
		}
		val = append(val, '\n')

		key = renderKeyFromTemplate(a.wKeyTemplate, v)
		hdrs = renderHeadersFromTemplates(a.wHdrTemplates, v)

		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			if key != nil {
				sensor.InvokeOnKafkaKeyRendered(a.componentMetadata, key)
			}
			if hdrs != nil {
				sensor.InvokeOnKafkaHeadersRendered(a.componentMetadata, hdrs)
			}
		}
		if key != nil {
			a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: KeyRendered, bytes: %d", a.componentMetadata, len(key))
		}
		if hdrs != nil {
			a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: HeadersRendered, count: %d", a.componentMetadata, len(hdrs))
		}
		return key, val, hdrs, nil
	}

	for {
		select {
		case <-ctx.Done():
			_ = flush()
			return nil

		case m, ok := <-in:
			if !ok {
				_ = flush()
				return nil
			}
			k, v, h, err := encode(m)
			if err != nil {
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Encode, err: %v", a.componentMetadata, err)
				continue
			}
			pending = append(pending, msg{key: k, val: v, headers: h})
			byteTally += len(k) + len(v)

			if len(pending) >= maxRecs || byteTally >= maxBytes {
				if err := flush(); err != nil {
					return err
				}
			}

		case <-tick.C:
			if err := flush(); err != nil {
				return err
			}
		}
	}
}
