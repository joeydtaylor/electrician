package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	kafka "github.com/segmentio/kafka-go"
	"github.com/segmentio/kafka-go/sasl"       // <- add this
	"github.com/segmentio/kafka-go/sasl/scram" // and keep this

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

// --- hardcoded dev "prod-like" settings ---
const (
	broker = "127.0.0.1:19092" // TLS+SASL listener from compose
	topic  = "feedback-demo"

	saslUser = "app"
	saslPass = "app-secret"
	saslMech = "SCRAM-SHA-256"

	caFileRel1 = "../tls/ca.crt" // try both so you can run from repo root or subdir
	caFileRel2 = "./tls/ca.crt"
)

// strict TLS (no InsecureSkipVerify)
func loadTLS() (*tls.Config, error) {
	var caPath string
	if _, err := os.Stat(caFileRel1); err == nil {
		caPath = caFileRel1
	} else if _, err := os.Stat(caFileRel2); err == nil {
		caPath = caFileRel2
	} else {
		return nil, fmt.Errorf("CA file not found (%s or %s)", caFileRel1, caFileRel2)
	}

	pem, err := os.ReadFile(filepath.Clean(caPath))
	if err != nil {
		return nil, fmt.Errorf("read CA: %w", err)
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(pem) {
		return nil, fmt.Errorf("invalid CA PEM at %s", caPath)
	}
	return &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: cp}, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// --- SASL + TLS ---
	var mech sasl.Mechanism // <- fix the type
	switch saslMech {
	case "SCRAM-SHA-256":
		m, err := scram.Mechanism(scram.SHA256, saslUser, saslPass)
		if err != nil {
			panic(err)
		}
		mech = m
	case "SCRAM-SHA-512":
		m, err := scram.Mechanism(scram.SHA512, saslUser, saslPass)
		if err != nil {
			panic(err)
		}
		mech = m
	default:
		panic("unsupported SASL mechanism")
	}

	tlsCfg, err := loadTLS()
	if err != nil {
		panic(err)
	}

	transport := &kafka.Transport{
		SASL:     mech,
		TLS:      tlsCfg,
		ClientID: "electrician-producer",
	}

	// --- sensors ---
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

	// kafka-go Writer with SASL/TLS transport
	kw := &kafka.Writer{
		Addr:                   kafka.TCP(broker),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		Transport:              transport,
		RequiredAcks:           kafka.RequireAll,
		AllowAutoTopicCreation: false,
		Async:                  false,
		BatchTimeout:           400 * time.Millisecond,
	}

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

	// send 2 demo records
	in := make(chan Feedback, 2)
	in <- Feedback{CustomerID: "C-01", Content: "I love this", Tags: []string{"demo"}}
	in <- Feedback{CustomerID: "C-02", Content: "Not great", IsNegative: true, Category: "ux"}
	close(in)

	if err := w.ServeWriter(ctx, in); err != nil {
		fmt.Printf("writer error: %v\n", err)
	}
}
