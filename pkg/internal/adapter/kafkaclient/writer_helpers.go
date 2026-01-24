package kafkaclient

import (
	"context"
	"fmt"
	"strings"

	"github.com/segmentio/kafka-go"
)

func (a *KafkaClient[T]) produce(
	ctx context.Context,
	msgTopic string,
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
		return int(msg.Partition), -1, nil
	default:
		return -1, -1, fmt.Errorf("kafkaclient: unsupported producer type %T (expected *kafka.Writer)", a.producer)
	}
}

func (a *KafkaClient[T]) effectiveWriterTopic() (string, bool) {
	if t := strings.TrimSpace(a.wTopic); t != "" {
		return t, true
	}
	if w, ok := a.producer.(*kafka.Writer); ok {
		if t := strings.TrimSpace(w.Topic); t != "" {
			return t, true
		}
	}
	return "", false
}
