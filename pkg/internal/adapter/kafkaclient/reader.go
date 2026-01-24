package kafkaclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/segmentio/kafka-go"
)

// Serve consumes Kafka records, decodes them into T, and dispatches to submit.
func (a *KafkaClient[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	if len(a.rTopics) == 0 {
		return fmt.Errorf("kafkaclient: Serve requires at least one Topic; call SetReaderConfig(...)")
	}
	if !atomic.CompareAndSwapInt32(&a.isServingReader, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServingReader, 0)

	r, created, err := a.getOrCreateReaderForServe()
	if err != nil {
		return err
	}
	if created {
		defer func() { _ = r.Close() }()
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ReaderCreated, group: %s, topics: %v",
			a.componentMetadata, a.rGroupID, a.rTopics)
	}

	policy := strings.ToLower(strings.TrimSpace(a.rCommitPolicy))
	if policy == "interval" {
		policy = "time"
	}
	mode := strings.ToLower(strings.TrimSpace(a.rCommitMode))
	if mode == "" {
		mode = "auto"
	}

	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnKafkaConsumerStart(a.componentMetadata, a.rGroupID, append([]string(nil), a.rTopics...), a.rStartAt)
	}
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConsumerStart, group: %s, topics: %v, start_at: %s",
		a.componentMetadata, a.rGroupID, a.rTopics, a.rStartAt)
	defer func() {
		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnKafkaConsumerStop(a.componentMetadata)
		}
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConsumerStop", a.componentMetadata)
	}()

	maxPollRecs := a.rMaxPollRecs
	if maxPollRecs <= 0 {
		maxPollRecs = 10_000
	}
	maxPollBytes := a.rMaxPollBytes
	if maxPollBytes <= 0 {
		maxPollBytes = 1 << 20
	}
	pollEvery := a.rPollInterval
	if pollEvery <= 0 {
		pollEvery = 1 * time.Second
	}

	latestByTP := make(map[string]kafka.Message)
	var commitTicker *time.Ticker
	if mode == "manual" && policy == "time" && a.rCommitEvery > 0 {
		commitTicker = time.NewTicker(a.rCommitEvery)
		defer commitTicker.Stop()
	}

	for {
		windowCtx, cancel := context.WithTimeout(ctx, pollEvery)

		decodedByTopic := map[string]int{}
		fetched := 0
		byteSum := 0

	inner:
		for fetched < maxPollRecs && byteSum < maxPollBytes {
			msg, err := r.FetchMessage(windowCtx)
			if err != nil {
				if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
					break inner
				}
				a.NotifyLoggers(types.WarnLevel, "%s => level: WARN, event: FetchMessage, err: %v", a.componentMetadata, err)
				break inner
			}

			for _, sensor := range a.snapshotSensors() {
				if sensor == nil {
					continue
				}
				sensor.InvokeOnKafkaMessage(a.componentMetadata, msg.Topic, msg.Partition, msg.Offset, len(msg.Key), len(msg.Value))
			}
			a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Message, topic: %s, partition: %d, offset: %d, key_bytes: %d, val_bytes: %d",
				a.componentMetadata, msg.Topic, msg.Partition, msg.Offset, len(msg.Key), len(msg.Value))

			var v T
			switch strings.ToLower(strings.TrimSpace(a.rFormat)) {
			case "ndjson", "", "json":
				if err := json.Unmarshal(msg.Value, &v); err != nil {
					a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Decode, topic: %s, err: %v", a.componentMetadata, msg.Topic, err)
					continue
				}
			default:
				if err := json.Unmarshal(msg.Value, &v); err != nil {
					a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Decode, topic: %s, err: %v", a.componentMetadata, msg.Topic, err)
					continue
				}
			}
			decodedByTopic[msg.Topic]++

			if err := submit(ctx, v); err != nil {
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Submit, err: %v", a.componentMetadata, err)
				continue
			}

			switch mode {
			case "manual":
				switch policy {
				case "after-each":
					if err := r.CommitMessages(ctx, msg); err != nil {
						for _, sensor := range a.snapshotSensors() {
							if sensor == nil {
								continue
							}
							sensor.InvokeOnKafkaCommitError(a.componentMetadata, a.rGroupID, err)
						}
						a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Commit(after-each), err: %v", a.componentMetadata, err)
					} else {
						key := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
						for _, sensor := range a.snapshotSensors() {
							if sensor == nil {
								continue
							}
							sensor.InvokeOnKafkaCommitSuccess(a.componentMetadata, a.rGroupID, map[string]int64{key: msg.Offset + 1})
						}
						a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: CommitSuccess(after-each), group: %s, offsets: %s=%d",
							a.componentMetadata, a.rGroupID, key, msg.Offset+1)
					}
				case "after-batch", "time", "":
					key := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
					latestByTP[key] = msg
				default:
					key := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
					latestByTP[key] = msg
				}
			case "auto":
			}

			fetched++
			byteSum += len(msg.Key) + len(msg.Value)
		}

		cancel()

		if len(decodedByTopic) > 0 {
			for t, n := range decodedByTopic {
				for _, sensor := range a.snapshotSensors() {
					if sensor == nil {
						continue
					}
					sensor.InvokeOnKafkaDecode(a.componentMetadata, t, n, a.rFormat)
				}
				a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Decode, topic: %s, rows: %d, format: %s",
					a.componentMetadata, t, n, a.rFormat)
			}
		}

		if mode == "manual" && policy == "after-batch" && len(latestByTP) > 0 {
			a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Commit(after-batch), partitions: %d",
				a.componentMetadata, len(latestByTP))
			a.commitLatest(ctx, r, latestByTP, a.rGroupID)
			latestByTP = make(map[string]kafka.Message)
		}

		if commitTicker != nil {
			select {
			case <-commitTicker.C:
				if len(latestByTP) > 0 {
					a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Commit(time), partitions: %d, every: %s",
						a.componentMetadata, len(latestByTP), a.rCommitEvery)
					a.commitLatest(ctx, r, latestByTP, a.rGroupID)
					latestByTP = make(map[string]kafka.Message)
				}
			default:
			}
		}

		select {
		case <-ctx.Done():
			if mode == "manual" && len(latestByTP) > 0 {
				a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Commit(final), partitions: %d",
					a.componentMetadata, len(latestByTP))
				a.commitLatest(context.Background(), r, latestByTP, a.rGroupID)
			}
			return nil
		default:
		}
	}
}

// Fetch is a placeholder for one-shot reads (not implemented).
func (a *KafkaClient[T]) Fetch() (types.HttpResponse[[]T], error) {
	return types.HttpResponse[[]T]{StatusCode: 501, Body: nil}, fmt.Errorf("kafkaclient: Fetch not implemented")
}
