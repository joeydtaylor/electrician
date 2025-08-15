package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

func envOr(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

// best-effort wait so we don't race the broker start
func waitForBroker(addr string, maxWait time.Duration) {
	deadline := time.Now().Add(maxWait)
	for {
		c, err := net.DialTimeout("tcp", addr, 500*time.Millisecond)
		if err == nil {
			_ = c.Close()
			return
		}
		if time.Now().After(deadline) {
			fmt.Printf("warning: broker %s not reachable yet: %v\n", addr, err)
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Use IPv4 loopback to avoid ::1 issues on some setups
	broker := envOr("BROKER", "127.0.0.1:19092")
	topic := envOr("TOPIC", "feedback-demo")
	waitForBroker(broker, 5*time.Second)

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// --- sensors (writer path) ---
	s := builder.NewSensor[Feedback](
		builder.SensorWithOnKafkaWriterStartFunc[Feedback](func(_ builder.ComponentMetadata, t, f string) {
			fmt.Printf("[sensor] writer.start topic=%s format=%s\n", t, f)
		}),
		builder.SensorWithOnKafkaKeyRenderedFunc[Feedback](func(_ builder.ComponentMetadata, key []byte) {
			fmt.Printf("[sensor] key.rendered len=%d\n", len(key))
		}),
		builder.SensorWithOnKafkaHeadersRenderedFunc[Feedback](func(_ builder.ComponentMetadata, hdrs []struct{ Key, Value string }) {
			fmt.Printf("[sensor] headers.rendered n=%d\n", len(hdrs))
		}),
		builder.SensorWithOnKafkaBatchFlushFunc[Feedback](func(_ builder.ComponentMetadata, t string, recs, bytes int, comp string) {
			fmt.Printf("[sensor] batch.flush topic=%s records=%d bytes=%d compression=%s\n", t, recs, bytes, comp)
		}),
		builder.SensorWithOnKafkaProduceAttemptFunc[Feedback](func(_ builder.ComponentMetadata, t string, p, keyB, valB int) {
			fmt.Printf("[sensor] produce.attempt topic=%s p=%d keyB=%d valB=%d\n", t, p, keyB, valB)
		}),
		builder.SensorWithOnKafkaProduceSuccessFunc[Feedback](func(_ builder.ComponentMetadata, _ string, _ int, _ int64, _ time.Duration) {
			fmt.Printf("[sensor] produce.success\n")
		}),
		builder.SensorWithOnKafkaWriterStopFunc[Feedback](func(_ builder.ComponentMetadata) {
			fmt.Println("[sensor] writer.stop")
		}),
	)

	// segmentio writer (topic set here; adapter won’t double-set)
	kw := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}

	w := builder.NewKafkaClientAdapter[Feedback](
		ctx,
		builder.KafkaClientAdapterWithKafkaGoWriter[Feedback](kw),
		builder.KafkaClientAdapterWithWriterFormat[Feedback]("ndjson", ""),
		builder.KafkaClientAdapterWithWriterBatchSettings[Feedback](2, 1<<20, 500*time.Millisecond),
		builder.KafkaClientAdapterWithWriterKeyTemplate[Feedback]("{customerId}"),
		builder.KafkaClientAdapterWithWriterHeaderTemplates[Feedback](map[string]string{
			"source": "demo-writer",
		}),
		builder.KafkaClientAdapterWithSensor[Feedback](s),
		builder.KafkaClientAdapterWithLogger[Feedback](log),
	)

	// send 2 demo records
	in := make(chan Feedback, 2)
	in <- Feedback{CustomerID: "C-01", Content: "I love this", Tags: []string{"demo"}}
	in <- Feedback{CustomerID: "C-02", Content: "Not great", IsNegative: true, Category: "ux"}
	close(in)

	if err := w.ServeWriter(ctx, in); err != nil {
		fmt.Printf("writer error: %v\n", err)
	}
}
