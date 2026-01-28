package kafkaclient

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/segmentio/kafka-go"
)

// ServeWriterRaw sends raw byte payloads to Kafka without encoding.
func (a *KafkaClient[T]) ServeWriterRaw(ctx context.Context, in <-chan []byte) error {
	effTopic, ok := a.effectiveWriterTopic()
	if !ok {
		return fmt.Errorf("kafkaclient: ServeWriterRaw requires topic (set KafkaWriterConfig.Topic or use a kafka-go Writer with Topic)")
	}

	if !atomic.CompareAndSwapInt32(&a.isServingWriter, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServingWriter, 0)

	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnKafkaWriterStart(a.componentMetadata, effTopic, "raw")
	}
	a.NotifyLoggers(
		types.InfoLevel,
		"Kafka writer started",
		"component", a.componentMetadata,
		"event", "WriterStart",
		"result", "SUCCESS",
		"topic", effTopic,
		"format", "raw",
	)
	defer func() {
		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnKafkaWriterStop(a.componentMetadata)
		}
		a.NotifyLoggers(
			types.InfoLevel,
			"Kafka writer stopped",
			"component", a.componentMetadata,
			"event", "WriterStop",
			"result", "SUCCESS",
		)
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

	type msg struct{ val []byte }
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
		a.NotifyLoggers(
			types.InfoLevel,
			"Kafka batch flush",
			"component", a.componentMetadata,
			"event", "BatchFlush",
			"result", "SUCCESS",
			"topic", effTopic,
			"records", len(pending),
			"bytes", byteTally,
			"compression", a.wCompression,
		)

		partForHook := -1
		if a.wManualPartition != nil {
			partForHook = *a.wManualPartition
		}

		for _, m := range pending {
			a.NotifyLoggers(
				types.DebugLevel,
				"Kafka produce attempt",
				"component", a.componentMetadata,
				"event", "ProduceAttempt",
				"topic", effTopic,
				"partition", partForHook,
				"key_bytes", 0,
				"val_bytes", len(m.val),
			)

			topicForMessage := ""
			if _, has := a.producer.(*kafka.Writer); !has {
				topicForMessage = effTopic
			}

			partition, offset, err := a.produce(ctx, topicForMessage, a.wManualPartition, nil, m.val, nil)
			if err != nil {
				for _, sensor := range a.snapshotSensors() {
					if sensor == nil {
						continue
					}
					sensor.InvokeOnKafkaProduceError(a.componentMetadata, effTopic, partForHook, err)
				}
				a.NotifyLoggers(
					types.ErrorLevel,
					"Kafka produce failed",
					"component", a.componentMetadata,
					"event", "Produce",
					"result", "FAILURE",
					"topic", effTopic,
					"error", err,
				)
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
			a.NotifyLoggers(
				types.InfoLevel,
				"Kafka produce success",
				"component", a.componentMetadata,
				"event", "ProduceSuccess",
				"result", "SUCCESS",
				"topic", effTopic,
				"partition", partition,
				"offset", offsetStr,
			)
		}

		pending = pending[:0]
		byteTally = 0
		return nil
	}

	for {
		select {
		case <-ctx.Done():
			_ = flush()
			return nil
		case b, ok := <-in:
			if !ok {
				_ = flush()
				return nil
			}
			if len(b) == 0 {
				continue
			}
			pending = append(pending, msg{val: append([]byte(nil), b...)})
			byteTally += len(b)

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
