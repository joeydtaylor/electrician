package kafkaclient

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/segmentio/kafka-go"
)

func (a *KafkaClient[T]) getOrCreateReaderForServe() (*kafka.Reader, bool, error) {
	if r, ok := a.consumer.(*kafka.Reader); ok && r != nil {
		return r, false, nil
	}

	brokers := append([]string(nil), a.brokers...)
	if len(brokers) == 0 {
		brokers = []string{"127.0.0.1:19092"}
	}

	cfg := kafka.ReaderConfig{Brokers: brokers}

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
		if len(a.rTopics) != 1 {
			return nil, false, fmt.Errorf("kafkaclient: reader without GroupID must specify exactly one topic")
		}
		cfg.Topic = a.rTopics[0]
	}

	switch strings.ToLower(a.rStartAt) {
	case "earliest":
		cfg.StartOffset = kafka.FirstOffset
	case "latest", "":
		cfg.StartOffset = kafka.LastOffset
	case "timestamp":
	default:
		cfg.StartOffset = kafka.LastOffset
	}

	if a.rPollInterval > 0 {
		cfg.MaxWait = a.rPollInterval
	}
	if a.rMaxPollBytes > 0 {
		cfg.MaxBytes = a.rMaxPollBytes
	} else {
		cfg.MaxBytes = 1_000_000
	}

	mode := strings.ToLower(strings.TrimSpace(a.rCommitMode))
	if mode == "" || mode == "auto" {
		if a.rCommitEvery > 0 {
			cfg.CommitInterval = a.rCommitEvery
		} else {
			cfg.CommitInterval = 3 * time.Second
		}
	} else {
		cfg.CommitInterval = 0
	}

	r := kafka.NewReader(cfg)

	if strings.EqualFold(a.rStartAt, "timestamp") && !a.rStartAtTime.IsZero() {
		_ = r.SetOffsetAt(context.Background(), a.rStartAtTime)
	}

	return r, true, nil
}

func (a *KafkaClient[T]) commitLatest(
	ctx context.Context,
	r *kafka.Reader,
	latest map[string]kafka.Message,
	groupID string,
) {
	if len(latest) == 0 {
		return
	}
	msgs := make([]kafka.Message, 0, len(latest))
	offsets := make(map[string]int64, len(latest))
	for k, m := range latest {
		msgs = append(msgs, m)
		offsets[k] = m.Offset + 1
	}
	if err := r.CommitMessages(ctx, msgs...); err != nil {
		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnKafkaCommitError(a.componentMetadata, groupID, err)
		}
		a.NotifyLoggers(
			types.ErrorLevel,
			"Kafka commit failed",
			"component", a.componentMetadata,
			"event", "Commit",
			"result", "FAILURE",
			"group", groupID,
			"error", err,
		)
		return
	}
	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnKafkaCommitSuccess(a.componentMetadata, groupID, offsets)
	}
}
