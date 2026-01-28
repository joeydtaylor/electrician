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
		a.NotifyLoggers(
			types.InfoLevel,
			"Kafka reader created",
			"component", a.componentMetadata,
			"event", "ReaderCreated",
			"group", a.rGroupID,
			"topics", a.rTopics,
		)
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
	a.NotifyLoggers(
		types.InfoLevel,
		"Kafka consumer started",
		"component", a.componentMetadata,
		"event", "ConsumerStart",
		"result", "SUCCESS",
		"group", a.rGroupID,
		"topics", a.rTopics,
		"start_at", a.rStartAt,
	)
	defer func() {
		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnKafkaConsumerStop(a.componentMetadata)
		}
		a.NotifyLoggers(
			types.InfoLevel,
			"Kafka consumer stopped",
			"component", a.componentMetadata,
			"event", "ConsumerStop",
			"result", "SUCCESS",
		)
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
				a.NotifyLoggers(
					types.WarnLevel,
					"FetchMessage warning",
					"component", a.componentMetadata,
					"event", "FetchMessage",
					"error", err,
				)
				break inner
			}

			for _, sensor := range a.snapshotSensors() {
				if sensor == nil {
					continue
				}
				sensor.InvokeOnKafkaMessage(a.componentMetadata, msg.Topic, msg.Partition, msg.Offset, len(msg.Key), len(msg.Value))
			}
			a.NotifyLoggers(
				types.InfoLevel,
				"Kafka message",
				"component", a.componentMetadata,
				"event", "Message",
				"topic", msg.Topic,
				"partition", msg.Partition,
				"offset", msg.Offset,
				"key_bytes", len(msg.Key),
				"val_bytes", len(msg.Value),
			)

			var v T
			switch strings.ToLower(strings.TrimSpace(a.rFormat)) {
			case "ndjson", "", "json":
				if err := json.Unmarshal(msg.Value, &v); err != nil {
					a.NotifyLoggers(
						types.ErrorLevel,
						"Kafka decode failed",
						"component", a.componentMetadata,
						"event", "Decode",
						"topic", msg.Topic,
						"error", err,
					)
					continue
				}
			default:
				if err := json.Unmarshal(msg.Value, &v); err != nil {
					a.NotifyLoggers(
						types.ErrorLevel,
						"Kafka decode failed",
						"component", a.componentMetadata,
						"event", "Decode",
						"topic", msg.Topic,
						"error", err,
					)
					continue
				}
			}
			decodedByTopic[msg.Topic]++

			if err := submit(ctx, v); err != nil {
				a.NotifyLoggers(
					types.ErrorLevel,
					"Kafka submit failed",
					"component", a.componentMetadata,
					"event", "Submit",
					"error", err,
				)
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
						a.NotifyLoggers(
							types.ErrorLevel,
							"Kafka commit failed",
							"component", a.componentMetadata,
							"event", "CommitAfterEach",
							"result", "FAILURE",
							"group", a.rGroupID,
							"topic", msg.Topic,
							"partition", msg.Partition,
							"offset", msg.Offset,
							"error", err,
						)
					} else {
						key := fmt.Sprintf("%s:%d", msg.Topic, msg.Partition)
						for _, sensor := range a.snapshotSensors() {
							if sensor == nil {
								continue
							}
							sensor.InvokeOnKafkaCommitSuccess(a.componentMetadata, a.rGroupID, map[string]int64{key: msg.Offset + 1})
						}
						a.NotifyLoggers(
							types.InfoLevel,
							"Kafka commit success",
							"component", a.componentMetadata,
							"event", "CommitAfterEach",
							"result", "SUCCESS",
							"group", a.rGroupID,
							"offsets", map[string]int64{key: msg.Offset + 1},
						)
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
				a.NotifyLoggers(
					types.InfoLevel,
					"Kafka decode",
					"component", a.componentMetadata,
					"event", "Decode",
					"result", "SUCCESS",
					"topic", t,
					"rows", n,
					"format", a.rFormat,
				)
			}
		}

		if mode == "manual" && policy == "after-batch" && len(latestByTP) > 0 {
			a.NotifyLoggers(
				types.InfoLevel,
				"Kafka commit after batch",
				"component", a.componentMetadata,
				"event", "CommitAfterBatch",
				"result", "SUCCESS",
				"group", a.rGroupID,
				"partitions", len(latestByTP),
			)
			a.commitLatest(ctx, r, latestByTP, a.rGroupID)
			latestByTP = make(map[string]kafka.Message)
		}

		if commitTicker != nil {
			select {
			case <-commitTicker.C:
				if len(latestByTP) > 0 {
					a.NotifyLoggers(
						types.InfoLevel,
						"Kafka commit on interval",
						"component", a.componentMetadata,
						"event", "CommitInterval",
						"result", "SUCCESS",
						"group", a.rGroupID,
						"partitions", len(latestByTP),
						"every", a.rCommitEvery,
					)
					a.commitLatest(ctx, r, latestByTP, a.rGroupID)
					latestByTP = make(map[string]kafka.Message)
				}
			default:
			}
		}

		select {
		case <-ctx.Done():
			if mode == "manual" && len(latestByTP) > 0 {
				a.NotifyLoggers(
					types.InfoLevel,
					"Kafka commit on shutdown",
					"component", a.componentMetadata,
					"event", "CommitFinal",
					"result", "SUCCESS",
					"group", a.rGroupID,
					"partitions", len(latestByTP),
				)
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
