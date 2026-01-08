package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync/atomic"
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

const (
	// Relay auth/crypto bits
	AES256KeyHex     = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
	assertJWTEnvName = "ASSERT_JWT"
	bearerJWTEnvName = "BEARER_JWT"
	orgIDEnvName     = "ORG_ID"
	defaultOrgID     = "4d948fa0-084e-490b-aad5-cfd01eeab79a"

	kTLSServerName        = "localhost" // must match server cert SAN; use "redpanda" if that’s what you issued
	kCACandidates         = "../tls/ca.crt"
	kClientCertCandidates = "../tls/client.crt"
	kClientKeyCandidates  = "../tls/client.key"

	// Kafka (matches your compose)
	kafkaBroker = "127.0.0.1:19092"
	kafkaTopic  = "feedback-demo"
	kafkaUser   = "app"
	kafkaPass   = "app-secret"
	kafkaMech   = "SCRAM-SHA-256"
	kafkaCA1    = "./tls/ca.crt"
	kafkaCA2    = "../tls/ca.crt"
	kafkaCA3    = "../../tls/ca.crt"
	kafkaSNI    = "localhost"
	clientID    = "electrician-writer"
	kSASLUser   = "app"
	kSASLPass   = "app-secret"
	kSASLMech   = "SCRAM-SHA-256"
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

func trueish(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes" || s == "y"
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDurMs(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Millisecond
		}
	}
	return def
}

func mustAES() string {
	raw, err := hex.DecodeString(AES256KeyHex)
	if err != nil || len(raw) != 32 {
		log.Fatalf("bad AES key: %v", err)
	}
	return string(raw)
}

func base64URLDecode(s string) ([]byte, error) {
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}

func orgFromJWT(tok string) (string, bool) {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return "", false
	}
	payloadB, err := base64URLDecode(parts[1])
	if err != nil {
		return "", false
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadB, &claims); err != nil {
		return "", false
	}
	for _, k := range []string{"org", "org_id", "orgId", "organization"} {
		if v, ok := claims[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s, true
			}
		}
	}
	return "", false
}

func resolveOrgID() string {
	if s := os.Getenv(orgIDEnvName); s != "" {
		return s
	}
	if s := os.Getenv(assertJWTEnvName); s != "" {
		if org, ok := orgFromJWT(s); ok {
			return org
		}
	}
	if s := os.Getenv(bearerJWTEnvName); s != "" {
		if org, ok := orgFromJWT(s); ok {
			return org
		}
	}
	return defaultOrgID
}

type kafkaStats struct {
	writerStarts    atomic.Int64
	writerStops     atomic.Int64
	batchFlushes    atomic.Int64
	recordsFlushed  atomic.Int64
	bytesFlushed    atomic.Int64
	produceSuccess  atomic.Int64
	lastCompression atomic.Value // string
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigs; fmt.Println("Shutting down..."); cancel() }()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// ------------------------
	// Receiving relays (TLS+JWT)
	// ------------------------
	// TLS (server)
	tlsCfg := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS13, tls.VersionTLS13,
	)

	// Kafka (security)
	// ---- security (only builder helpers you exported) ----
	kTlsCfg, err := buildTLSConfig()
	if err != nil {
		_ = json.NewEncoder(os.Stdout).Encode(map[string]any{"error": err.Error()})
		return
	}
	mech, err := builder.SASLSCRAM(kSASLUser, kSASLPass, kSASLMech)
	if err != nil {
		panic(err)
	}
	sec := builder.NewKafkaSecurity(
		builder.WithTLS(kTlsCfg),
		builder.WithSASL(mech),
		builder.WithClientID(clientID),
		// Dialer defaults (10s, DualStack=true) are fine for a writer Transport
	)

	// OAuth (JWKS-only)
	issuerBase := envOr("OAUTH_ISSUER_BASE", "https://localhost:3000")
	jwksURL := envOr("OAUTH_JWKS_URL", "https://localhost:3000/api/auth/.well-known/jwks.json")
	jwtOpts := builder.NewReceivingRelayOAuth2JWTOptions(
		issuerBase,
		jwksURL,
		[]string{"your-api"},
		[]string{"write:data"},
		300,
	)
	oauth := builder.NewReceivingRelayMergeOAuth2Options(jwtOpts, nil)
	authOpts := builder.NewReceivingRelayAuthenticationOptionsOAuth2(oauth)

	staticHeaders := map[string]string{"x-tenant": "local"}
	decKey := mustAES()

	// Wires (no transforms)
	wireA := builder.NewWire[Feedback](ctx, builder.WireWithLogger[Feedback](logger))
	wireB := builder.NewWire[Feedback](ctx, builder.WireWithLogger[Feedback](logger))
	_ = wireA.Start(ctx)
	_ = wireB.Start(ctx)

	recvA := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback](envOr("RX_A", "localhost:50052")),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(wireA),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsCfg),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decKey),
		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ReceivingRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.ReceivingRelayWithAuthRequired[Feedback](true),
	)
	recvB := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback](envOr("RX_B", "localhost:50051")),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(wireB),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsCfg),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decKey),
		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ReceivingRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.ReceivingRelayWithAuthRequired[Feedback](true),
	)

	// Start receivers
	if err := recvA.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := recvB.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// kafka-go Writer with security
	kw := builder.NewKafkaGoWriterWithSecurity(
		[]string{kafkaBroker},
		kafkaTopic,
		sec,
		builder.KafkaGoWriterWithLeastBytes(),
		builder.KafkaGoWriterWithBatchTimeout(400*time.Millisecond),
	)

	// Kafka sensors + counters
	var kstats kafkaStats
	kstats.lastCompression.Store("")

	sens := builder.NewSensor[Feedback](
		builder.SensorWithLogger[Feedback](logger),
		builder.SensorWithOnKafkaWriterStartFunc[Feedback](func(_ builder.ComponentMetadata, _topic, _fmt string) {
			kstats.writerStarts.Add(1)
		}),
		builder.SensorWithOnKafkaBatchFlushFunc[Feedback](func(_ builder.ComponentMetadata, _t string, recs, bytes int, comp string) {
			kstats.batchFlushes.Add(1)
			kstats.recordsFlushed.Add(int64(recs))
			kstats.bytesFlushed.Add(int64(bytes))
			if comp != "" {
				kstats.lastCompression.Store(comp)
			}
		}),
		builder.SensorWithOnKafkaProduceSuccessFunc[Feedback](func(_ builder.ComponentMetadata, _ string, _ int, _ int64, _ time.Duration) {
			kstats.produceSuccess.Add(1)
		}),
		builder.SensorWithOnKafkaWriterStopFunc[Feedback](func(_ builder.ComponentMetadata) {
			kstats.writerStops.Add(1)
		}),
	)

	// Adapter: connect writer + fan-in wires (NOTE: include WriterTopic)
	kad := builder.NewKafkaClientAdapter[Feedback](
		ctx,
		builder.KafkaClientAdapterWithKafkaGoWriter[Feedback](kw),
		builder.KafkaClientAdapterWithWriterTopic[Feedback](kafkaTopic), // required by StartWriter
		builder.KafkaClientAdapterWithWriterFormat[Feedback]("ndjson", ""),
		builder.KafkaClientAdapterWithWriterBatchSettings[Feedback](
			envInt("BATCH_MAX_RECORDS", 10000),
			envInt("BATCH_MAX_BYTES_MB", 16)*(1<<20),
			envDurMs("BATCH_MAX_AGE_MS", 800*time.Millisecond),
		),
		builder.KafkaClientAdapterWithWriterKeyTemplate[Feedback]("{customerId}"),
		builder.KafkaClientAdapterWithWriterHeaderTemplates[Feedback](map[string]string{"source": "relay-writer"}),
		builder.KafkaClientAdapterWithWire[Feedback](wireA, wireB),
		builder.KafkaClientAdapterWithSensor[Feedback](sens),
		builder.KafkaClientAdapterWithLogger[Feedback](logger),
	)

	// Stream until canceled
	if err := kad.StartWriter(ctx); err != nil {
		log.Fatalf("Kafka StartWriter failed: %v", err)
	}

	fmt.Println("Receivers up; streaming to Kafka. Ctrl+C to stop.")
	<-ctx.Done()

	// Graceful shutdown
	recvA.Stop()
	recvB.Stop()
	wireA.Stop()
	wireB.Stop()
	kad.Stop()

	// Summary
	orgID := resolveOrgID()
	out := map[string]any{
		"orgId":            orgID,
		"topic":            kafkaTopic,
		"broker":           kafkaBroker,
		"writerStarts":     kstats.writerStarts.Load(),
		"writerStops":      kstats.writerStops.Load(),
		"batchFlushes":     kstats.batchFlushes.Load(),
		"recordsFlushed":   kstats.recordsFlushed.Load(),
		"bytesFlushed":     kstats.bytesFlushed.Load(),
		"produceSuccesses": kstats.produceSuccess.Load(),
		"lastCompression":  kstats.lastCompression.Load(),
	}
	js, _ := json.MarshalIndent(out, "", "  ")
	fmt.Println(string(js))
	fmt.Println("Done.")
}
