package main

import (
	"context"
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

const (
	brokersCSV = "127.0.0.1:19092"
	topic      = "feedback-demo"
	group      = "feedback-demo-reader"

	tlsServerName        = "localhost"
	caCandidates         = "../tls/ca.crt"
	clientCertCandidates = "../tls/client.crt"
	clientKeyCandidates  = "../tls/client.key"

	saslUser = "app"
	saslPass = "app-secret"
	saslMech = "SCRAM-SHA-256"
)

func splitCSV(csv string) []string {
	var out []string
	for _, part := range strings.Split(csv, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	groupID := fmt.Sprintf("%s-%d", group, time.Now().UnixNano())

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() { <-sig; fmt.Println("\n[signal] interrupt received, shutting down..."); cancel() }()

	mech, err := builder.SASLSCRAM(saslUser, saslPass, saslMech)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

	tlsCfg, err := builder.TLSFromMTLSPathCSV(caCandidates, clientCertCandidates, clientKeyCandidates, tlsServerName)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

	sec := builder.NewKafkaSecurity(
		builder.WithTLS(tlsCfg),
		builder.WithSASL(mech),
	)

	kr := builder.NewKafkaGoReaderWithSecurity(
		splitCSV(brokersCSV),
		groupID,
		[]string{topic},
		sec,
		builder.KafkaGoReaderWithLatestStart(),
		builder.KafkaGoReaderWithMinBytes(1<<10),
		builder.KafkaGoReaderWithMaxBytes(10<<20),
		builder.KafkaGoReaderWithCommitInterval(0),
	)

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))
	s := builder.NewSensor[Feedback](
		builder.SensorWithLogger[Feedback](log),
	)

	reader := builder.NewKafkaClientAdapter[Feedback](
		ctx,
		builder.KafkaClientAdapterWithKafkaGoReader[Feedback](kr),
		builder.KafkaClientAdapterWithReaderTopics[Feedback](topic),
		builder.KafkaClientAdapterWithReaderGroup[Feedback](groupID),
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
