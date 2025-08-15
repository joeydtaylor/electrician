package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"
	"github.com/segmentio/kafka-go/sasl/scram"

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

// --- hardcoded prod-ish config ----
const (
	// External TLS+SASL (matches compose external listener)
	kBrokersCSV    = "127.0.0.1:19092"
	kTLS           = true
	kTLSServerName = "localhost"
	kCACandidates  = "./tls/ca.crt,../tls/ca.crt,../../tls/ca.crt"

	kTopic = "feedback-demo"
	kGroup = "feedback-demo-reader"

	kSASLUser = "app"
	kSASLPass = "app-secret"
	kSASLMech = "SCRAM-SHA-256" // or SCRAM-SHA-512
)

func buildSASL(user, pass, mech string) (sasl.Mechanism, error) {
	switch strings.ToUpper(mech) {
	case "SCRAM-SHA-512", "SCRAM_SHA_512":
		return scram.Mechanism(scram.SHA512, user, pass)
	case "SCRAM-SHA-256", "SCRAM_SHA_256", "":
		return scram.Mechanism(scram.SHA256, user, pass)
	default:
		return nil, fmt.Errorf("unsupported SASL mechanism: %s", mech)
	}
}

func firstExisting(pathsCSV string) (string, error) {
	for _, p := range strings.Split(pathsCSV, ",") {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p, nil
		}
	}
	return "", fmt.Errorf("no CA file found in: %s", pathsCSV)
}

func buildTLS(candidates, serverName string) (*tls.Config, error) {
	caPath, err := firstExisting(candidates)
	if err != nil {
		return nil, err
	}
	pem, err := os.ReadFile(filepath.Clean(caPath))
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("invalid CA PEM at %s", caPath)
	}
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    cp,
		ServerName: serverName,
	}, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mech, err := buildSASL(kSASLUser, kSASLPass, kSASLMech)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}
	var tlsCfg *tls.Config
	if kTLS {
		tlsCfg, err = buildTLS(kCACandidates, kTLSServerName)
		if err != nil {
			_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
			return
		}
	}

	dialer := &kafka.Dialer{
		Timeout:       10 * time.Second,
		DualStack:     true,
		SASLMechanism: mech,
		TLS:           tlsCfg, // nil for SASL_PLAINTEXT
	}

	rc := kafka.ReaderConfig{
		Brokers:        strings.Split(kBrokersCSV, ","),
		GroupID:        kGroup,
		GroupTopics:    []string{kTopic},
		MinBytes:       1 << 10,  // 1 KiB
		MaxBytes:       10 << 20, // 10 MiB
		StartOffset:    kafka.LastOffset,
		CommitInterval: 0, // manual/time-based commits via adapter
		Dialer:         dialer,
		// IsolationLevel: kafka.ReadCommitted, // uncomment if producers are transactional
	}
	kr := kafka.NewReader(rc)

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

	if err := reader.Serve(ctx, submit); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}

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
