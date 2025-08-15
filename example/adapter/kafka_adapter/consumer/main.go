package main

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
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

func newAgg() *agg { return &agg{byCategory: map[string]int{}, tagFreq: map[string]int{}} }
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

// --- prod-ish config (matches your compose) ----
const (
	kBrokersCSV    = "127.0.0.1:19092" // external TLS+SASL listener
	kTLS           = true
	kTLSServerName = "localhost"
	kCACandidates  = "./tls/ca.crt,../tls/ca.crt,../../tls/ca.crt"

	kTopic = "feedback-demo"
	kGroup = "feedback-demo-reader"

	kSASLUser = "app"
	kSASLPass = "app-secret"
	kSASLMech = "SCRAM-SHA-256" // or SCRAM-SHA-512
)

func main() {
	// Consume until Ctrl+C (SIGINT/SIGTERM), then print summary
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Trap Ctrl+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("\n[signal] interrupt received, shutting down...")
		cancel()
	}()

	// --- Security (SASL + TLS) via builder helpers ---
	mech, err := builder.SASLSCRAM(kSASLUser, kSASLPass, kSASLMech)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

	var tlsCfg *tls.Config // created only if kTLS
	if kTLS {
		tlsCfg, err = builder.TLSFromCAPathCSV(kCACandidates, kTLSServerName)
		if err != nil {
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
			return
		}
	}

	// kafka-go Reader with secure dialer (TLS/SASL)
	kr := builder.NewKafkaGoReaderSecure(
		strings.Split(kBrokersCSV, ","),
		kGroup,
		[]string{kTopic},
		tlsCfg,
		mech,
		10*time.Second, // dialer timeout
		true,           // dual stack
		// ReaderConfig-like knobs:
		builder.KafkaGoReaderWithLatestStart(),
		builder.KafkaGoReaderWithMinBytes(1<<10),
		builder.KafkaGoReaderWithMaxBytes(10<<20),
		builder.KafkaGoReaderWithCommitInterval(0), // disable driver auto-commit
	)

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))
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
		builder.SensorWithOnKafkaMessageFunc[Feedback](func(cm builder.ComponentMetadata, t string, partition int, offset int64, keyBytes, valueBytes int) {
			fmt.Printf("[sensor] message topic=%s p=%d off=%d keyB=%d valB=%d\n", t, partition, offset, keyBytes, valueBytes)
		}),
		builder.SensorWithOnKafkaCommitSuccessFunc[Feedback](func(cm builder.ComponentMetadata, group string, offsets map[string]int64) {
			fmt.Printf("[sensor] commit.success group=%s offsets=%v\n", group, offsets)
		}),
		builder.SensorWithOnKafkaCommitErrorFunc[Feedback](func(cm builder.ComponentMetadata, group string, err error) {
			fmt.Printf("[sensor] commit.error group=%s err=%v\n", group, err)
		}),
	)

	reader := builder.NewKafkaClientAdapter[Feedback](
		ctx,
		builder.KafkaClientAdapterWithKafkaGoReader[Feedback](kr),
		// adapter-level intent (sensors/behavior); underlying reader already set:
		builder.KafkaClientAdapterWithReaderTopics[Feedback](kTopic),
		builder.KafkaClientAdapterWithReaderGroup[Feedback](kGroup),
		builder.KafkaClientAdapterWithReaderStartAt[Feedback]("latest", time.Time{}),
		builder.KafkaClientAdapterWithReaderPollSettings[Feedback](1*time.Second, 10000, 4<<20),
		builder.KafkaClientAdapterWithReaderFormat[Feedback]("ndjson", ""),
		builder.KafkaClientAdapterWithReaderCommit[Feedback]("manual", "time", 3*time.Second),
		builder.KafkaClientAdapterWithLogger[Feedback](log),
		builder.KafkaClientAdapterWithSensor[Feedback](s),
	)

	a := newAgg()
	submit := func(_ context.Context, f Feedback) error { a.add(f); return nil }

	if err := reader.Serve(ctx, submit); err != nil && !errors.Is(err, context.Canceled) {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

	printSummary(a)
}

func printSummary(a *agg) {
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

	var tags []kv
	for k, v := range a.tagFreq {
		tags = append(tags, kv{k, v})
	}
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].v == tags[j].v {
			return tags[i].k < tags[j].k
		}
		return tags[i].v > tags[j].v
	})

	fmt.Printf("\n=== Kafka summary (since start) ===\n")
	fmt.Printf("total=%d negatives=%d positives=%d\n", a.total, a.negCount, a.total-a.negCount)
	fmt.Println("by_category:")
	for _, c := range cats {
		fmt.Printf("  - %-16s : %d\n", c.k, c.v)
	}
	fmt.Println("top_tags:")
	for i, t := range tags {
		if i >= 15 {
			break
		}
		fmt.Printf("  - %-16s : %d\n", t.k, t.v)
	}
}
