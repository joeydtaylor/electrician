package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
	broker       = "127.0.0.1:19092" // TLS+SASL listener from compose
	topic        = "feedback-demo"
	serverName   = "localhost"
	saslUser     = "app"
	saslPass     = "app-secret"
	saslMech     = "SCRAM-SHA-256"
	clientID     = "electrician-producer"
	batchTimeout = 400 * time.Millisecond
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// ---- security (only builder helpers you exported) ----
	tlsCfg, err := builder.TLSFromCAFilesStrict([]string{
		"./tls/ca.crt",
		"../tls/ca.crt",
		"../../tls/ca.crt",
	}, serverName)
	if err != nil {
		panic(err)
	}
	mech, err := builder.SASLSCRAM(saslUser, saslPass, saslMech)
	if err != nil {
		panic(err)
	}
	sec := builder.NewKafkaSecurity(
		builder.WithTLS(tlsCfg),
		builder.WithSASL(mech),
		builder.WithClientID(clientID),
		// Dialer defaults (10s, DualStack=true) are fine for a writer Transport
	)

	// ---- kafka-go writer (secured via NewKafkaGoWriterWithSecurity) ----
	kw := builder.NewKafkaGoWriterWithSecurity(
		[]string{broker},
		topic,
		sec,
		builder.KafkaGoWriterWithLeastBytes(),
		builder.KafkaGoWriterWithBatchTimeout(batchTimeout),
	)
	// (RequiredAcks defaults to All in NewKafkaGoWriter)

	// ---- sensors/loggers (unchanged from your example) ----
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

	// ---- adapter wiring ----
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

	// demo records
	in := make(chan Feedback, 2)
	in <- Feedback{CustomerID: "C-01", Content: "I love this", Tags: []string{"demo"}}
	in <- Feedback{CustomerID: "C-02", Content: "Not great", IsNegative: true, Category: "ux"}
	close(in)

	if err := w.ServeWriter(ctx, in); err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
	}
}
