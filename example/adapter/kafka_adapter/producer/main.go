package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

const (
	brokersCSV   = "127.0.0.1:19092" // TLS+SASL listener from compose
	topic        = "feedback-demo"
	clientID     = "electrician-producer"
	batchTimeout = 400 * time.Millisecond

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
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tlsCfg, err := builder.TLSFromMTLSPathCSV(caCandidates, clientCertCandidates, clientKeyCandidates, tlsServerName)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}
	mech, err := builder.SASLSCRAM(saslUser, saslPass, saslMech)
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}
	sec := builder.NewKafkaSecurity(
		builder.WithTLS(tlsCfg),
		builder.WithSASL(mech),
		builder.WithClientID(clientID),
	)

	kw := builder.NewKafkaGoWriterWithSecurity(
		splitCSV(brokersCSV),
		topic,
		sec,
		builder.KafkaGoWriterWithLeastBytes(),
		builder.KafkaGoWriterWithBatchTimeout(batchTimeout),
	)

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))
	s := builder.NewSensor[Feedback](
		builder.SensorWithOnKafkaWriterStartFunc[Feedback](func(_ builder.ComponentMetadata, t, f string) {
			fmt.Printf("[sensor] writer.start topic=%s format=%s\n", t, f)
		}),
		builder.SensorWithOnKafkaBatchFlushFunc[Feedback](func(_ builder.ComponentMetadata, t string, recs, bytes int, comp string) {
			fmt.Printf("[sensor] batch.flush topic=%s records=%d bytes=%d compression=%s\n", t, recs, bytes, comp)
		}),
		builder.SensorWithOnKafkaProduceSuccessFunc[Feedback](func(_ builder.ComponentMetadata, _ string, _ int, _ int64, _ time.Duration) {
			fmt.Printf("[sensor] produce.success\n")
		}),
		builder.SensorWithOnKafkaWriterStopFunc[Feedback](func(_ builder.ComponentMetadata) {
			fmt.Println("[sensor] writer.stop")
		}),
	)

	w := builder.NewKafkaClientAdapter[Feedback](
		ctx,
		builder.KafkaClientAdapterWithKafkaGoWriter[Feedback](kw),
		builder.KafkaClientAdapterWithWriterFormat[Feedback]("ndjson", ""),
		builder.KafkaClientAdapterWithWriterBatchSettings[Feedback](2, 1<<20, 500*time.Millisecond),
		builder.KafkaClientAdapterWithWriterKeyTemplate[Feedback]("{customerId}"),
		builder.KafkaClientAdapterWithWriterHeaderTemplates[Feedback](map[string]string{"source": "demo-writer"}),
		builder.KafkaClientAdapterWithSensor[Feedback](s),
		builder.KafkaClientAdapterWithLogger[Feedback](log),
	)

	in := make(chan Feedback, 2)
	in <- Feedback{CustomerID: "C-01", Content: "I love this", Tags: []string{"demo"}}
	in <- Feedback{CustomerID: "C-02", Content: "Not great", IsNegative: true, Category: "ux"}
	close(in)

	if err := w.ServeWriter(ctx, in); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
	}
}
