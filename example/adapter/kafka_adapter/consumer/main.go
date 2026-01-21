package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
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

// --- prod-ish config (matches compose) ----
const (
	kBrokersCSV           = "127.0.0.1:19092" // external TLS+mTLS listener
	kTLSServerName        = "localhost"       // must match server cert SAN; use "redpanda" if that’s what you issued
	kCACandidates         = "../tls/ca.crt"
	kClientCertCandidates = "../tls/client.crt"
	kClientKeyCandidates  = "../tls/client.key"

	kTopic = "feedback-demo"
	kGroup = "feedback-demo-reader"

	kSASLUser = "app"
	kSASLPass = "app-secret"
	kSASLMech = "SCRAM-SHA-256"
)

func firstExisting(csv string) (string, error) {
	for _, p := range strings.Split(csv, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("no file found in: %s", csv)
}

func buildTLSConfig() (*tls.Config, error) {
	caPath, err := firstExisting(kCACandidates)
	if err != nil {
		return nil, err
	}
	caPEM, err := os.ReadFile(caPath)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return nil, fmt.Errorf("failed adding CA to pool: %s", caPath)
	}

	certPath, err := firstExisting(kClientCertCandidates)
	if err != nil {
		return nil, err
	}
	keyPath, err := firstExisting(kClientKeyCandidates)
	if err != nil {
		return nil, err
	}
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS12,
		ServerName:   kTLSServerName, // SNI + name verification
		RootCAs:      pool,
		Certificates: []tls.Certificate{cert}, // mTLS: present client cert
	}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() { <-sig; fmt.Println("\n[signal] interrupt received, shutting down..."); cancel() }()

	mech, err := builder.SASLSCRAM(kSASLUser, kSASLPass, kSASLMech)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

	tlsCfg, err := buildTLSConfig()
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

	kr := builder.NewKafkaGoReaderSecure(
		strings.Split(kBrokersCSV, ","),
		kGroup,
		[]string{kTopic},
		tlsCfg,
		mech,
		10*time.Second, // dialer timeout
		true,           // dual stack
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
