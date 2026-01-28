package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
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
	AES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envOrInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envOrDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
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

func main() {
	count := envOrInt("RELAY_COUNT", 100)
	interval := envOrDuration("RELAY_INTERVAL", 50*time.Millisecond)
	addr := envOr("RELAY_ADDR", "localhost:50072")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(count)*interval+2*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))
	defer func() {
		_ = logger.Flush()
	}()

	tlsCfg := builder.NewTlsClientConfig(
		true,
		envOr("TLS_CERT", "../tls/client.crt"),
		envOr("TLS_KEY", "../tls/client.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	sec := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)

	token := envOr("RELAY_BEARER_TOKEN", "")
	tokenEnv := envOr("RELAY_BEARER_TOKEN_ENV", "")

	fr := builder.NewQuicForwardRelay[Feedback](
		ctx,
		builder.QuicForwardRelayWithTarget[Feedback](addr),
		builder.QuicForwardRelayWithLogger[Feedback](logger),
		builder.QuicForwardRelayWithTLSConfig[Feedback](tlsCfg),
		builder.QuicForwardRelayWithSecurityOptions[Feedback](sec, mustAES()),
		builder.QuicForwardRelayWithStaticHeaders[Feedback](map[string]string{"x-tenant": "local"}),
		builder.QuicForwardRelayWithAuthRequired[Feedback](true),
	)
	if token != "" {
		fr.SetOAuth2(builder.NewQuicForwardRelayStaticBearerTokenSource(token))
	} else if tokenEnv != "" {
		fr.SetOAuth2(builder.NewQuicForwardRelayEnvBearerTokenSource(tokenEnv))
	}

	start := time.Now()
	for i := 0; i < count; i++ {
		fb := Feedback{
			CustomerID: fmt.Sprintf("cust-%03d", i%100),
			Content:    "Secure QUIC payload",
			Category:   "feedback",
			IsNegative: i%10 == 0,
			Tags:       []string{"quic", "secure"},
		}
		if err := fr.Submit(ctx, fb); err != nil {
			log.Printf("submit error: %v", err)
		}
		time.Sleep(interval)
	}

	log.Printf("sent %d in %s", count, time.Since(start))
}
