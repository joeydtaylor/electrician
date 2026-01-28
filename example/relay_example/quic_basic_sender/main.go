package main

import (
	"context"
	"crypto/tls"
	"fmt"
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

func main() {
	count := envOrInt("RELAY_COUNT", 100)
	interval := envOrDuration("RELAY_INTERVAL", 50*time.Millisecond)
	addr := envOr("RELAY_ADDR", "localhost:50071")
	logLevel := envOr("LOG_LEVEL", "info")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(count)*interval+2*time.Second)
	defer cancel()

	logger := builder.NewLogger(
		builder.LoggerWithDevelopment(true),
		builder.LoggerWithLevel(logLevel),
	)
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

	fr := builder.NewQuicForwardRelay[Feedback](
		ctx,
		builder.QuicForwardRelayWithTarget[Feedback](addr),
		builder.QuicForwardRelayWithLogger[Feedback](logger),
		builder.QuicForwardRelayWithTLSConfig[Feedback](tlsCfg),
	)

	logger.Info("QUIC sender starting",
		"event", "Start",
		"result", "SUCCESS",
		"target", addr,
		"count", count,
		"interval", interval.String(),
	)

	start := time.Now()
	for i := 0; i < count; i++ {
		fb := Feedback{
			CustomerID: fmt.Sprintf("cust-%03d", i%100),
			Content:    "Relay test payload",
			Category:   "feedback",
			IsNegative: i%10 == 0,
			Tags:       []string{"quic", "basic"},
		}
		if err := fr.Submit(ctx, fb); err != nil {
			logger.Error("Submit failed",
				"event", "Submit",
				"result", "FAILURE",
				"error", err,
				"seq", i,
			)
		}
		time.Sleep(interval)
	}

	logger.Info("Send complete",
		"event", "SendComplete",
		"result", "SUCCESS",
		"count", count,
		"elapsed", time.Since(start).String(),
	)
}
