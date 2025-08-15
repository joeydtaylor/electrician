// examples/kafka_reader/main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

type agg struct {
	byCategory map[string]int
	tagFreq    map[string]int
	negCount   int
	total      int
}

func newAgg() *agg {
	return &agg{
		byCategory: map[string]int{},
		tagFreq:    map[string]int{},
	}
}
func (a *agg) add(r Feedback) {
	a.total++
	if r.IsNegative {
		a.negCount++
	}
	if r.Category != "" {
		a.byCategory[r.Category]++
	}
	for _, t := range r.Tags {
		if t = strings.TrimSpace(t); t != "" {
			a.tagFreq[t]++
		}
	}
}

func envOr(k, def string) string {
	if v := strings.TrimSpace(os.Getenv(k)); v != "" {
		return v
	}
	return def
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// ---- env ----
	brokersCSV := envOr("KAFKA_BROKERS", "127.0.0.1:19092") // matches your compose
	topic := envOr("KAFKA_TOPIC", "feedback-demo")
	group := envOr("KAFKA_GROUP", "feedback-demo-reader")

	// ---- build a kafka-go Reader and inject it ----
	kr := builder.NewKafkaGoReader(
		strings.Split(brokersCSV, ","),
		group,
		[]string{topic},
		builder.KafkaGoReaderWithLatestStart(),     // start at "latest"
		builder.KafkaGoReaderWithMinBytes(1<<10),   // 1 KiB
		builder.KafkaGoReaderWithMaxBytes(10<<20),  // 10 MiB
		builder.KafkaGoReaderWithCommitInterval(0), // we'll do manual commits via adapter
	)

	// ---- sensor: fire the Kafka consumer hooks ----
	s := builder.NewSensor[Feedback](
		builder.SensorWithLogger[Feedback](log),

		builder.SensorWithOnStartFunc[Feedback](func(cm builder.ComponentMetadata) {
			fmt.Printf("[sensor] start: %s (%s)\n", cm.Name, cm.ID)
		}),
		builder.SensorWithOnStopFunc[Feedback](func(cm builder.ComponentMetadata) {
			fmt.Printf("[sensor] stop: %s (%s)\n", cm.Name, cm.ID)
		}),

		builder.SensorWithOnKafkaConsumerStartFunc[Feedback](func(cm builder.ComponentMetadata, group string, topics []string, startAt string) {
			fmt.Printf("[sensor] consumer.start group=%s topics=%v startAt=%s\n", group, topics, startAt)
		}),
		builder.SensorWithOnKafkaConsumerStopFunc[Feedback](func(cm builder.ComponentMetadata) {
			fmt.Println("[sensor] consumer.stop")
		}),
		builder.SensorWithOnKafkaPartitionAssignedFunc[Feedback](func(cm builder.ComponentMetadata, t string, p int, start, end int64) {
			fmt.Printf("[sensor] partition.assigned topic=%s p=%d start=%d end=%d\n", t, p, start, end)
		}),
		builder.SensorWithOnKafkaPartitionRevokedFunc[Feedback](func(cm builder.ComponentMetadata, t string, p int) {
			fmt.Printf("[sensor] partition.revoked topic=%s p=%d\n", t, p)
		}),
		builder.SensorWithOnKafkaMessageFunc[Feedback](func(cm builder.ComponentMetadata, t string, partition int, offset int64, keyBytes, valueBytes int) {
			fmt.Printf("[sensor] message topic=%s p=%d off=%d keyB=%d valB=%d\n", t, partition, offset, keyBytes, valueBytes)
		}),
		builder.SensorWithOnKafkaDecodeFunc[Feedback](func(cm builder.ComponentMetadata, t string, rows int, format string) {
			fmt.Printf("[sensor] decode topic=%s rows=%d format=%s\n", t, rows, format)
		}),
		builder.SensorWithOnKafkaCommitSuccessFunc[Feedback](func(cm builder.ComponentMetadata, group string, offsets map[string]int64) {
			fmt.Printf("[sensor] commit.success group=%s offsets=%v\n", group, offsets)
		}),
		builder.SensorWithOnKafkaCommitErrorFunc[Feedback](func(cm builder.ComponentMetadata, group string, err error) {
			fmt.Printf("[sensor] commit.error group=%s err=%v\n", group, err)
		}),
	)

	// ---- adapter (reader) ----
	reader := builder.NewKafkaClientAdapter[Feedback](
		ctx,
		builder.KafkaClientAdapterWithKafkaGoReader[Feedback](kr),

		// keep these so the adapter’s metadata & hooks are populated consistently
		builder.KafkaClientAdapterWithReaderTopics[Feedback](topic),
		builder.KafkaClientAdapterWithReaderGroup[Feedback](group),
		builder.KafkaClientAdapterWithReaderStartAt[Feedback]("latest", time.Time{}),

		// poll/format/commit choices (manual time-based commits to trigger commit hooks)
		builder.KafkaClientAdapterWithReaderPollSettings[Feedback](1*time.Second, 10000, 4<<20),
		builder.KafkaClientAdapterWithReaderFormat[Feedback]("ndjson", ""),
		builder.KafkaClientAdapterWithReaderCommit[Feedback]("manual", "time", 3*time.Second),

		builder.KafkaClientAdapterWithLogger[Feedback](log),
		builder.KafkaClientAdapterWithSensor[Feedback](s),
	)

	a := newAgg()
	submit := func(_ context.Context, f Feedback) error {
		a.add(f)
		return nil
	}

	if err := reader.Serve(ctx, submit); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

	// compact summary
	type kv struct {
		k string
		v int
	}
	var cats []kv
	for k, v := range a.byCategory {
		cats = append(cats, kv{k, v})
	}
	sort.Slice(cats, func(i, j int) bool {
		if cats[i].v == cats[j].v {
			return cats[i].k < cats[j].k
		}
		return cats[i].v > cats[j].v
	})

	fmt.Printf("\n=== Kafka summary ===\n")
	fmt.Printf("total=%d negatives=%d positives=%d\n", a.total, a.negCount, a.total-a.negCount)
	fmt.Println("by_category:")
	for _, c := range cats {
		fmt.Printf("  - %-12s : %d\n", c.k, c.v)
	}
}
