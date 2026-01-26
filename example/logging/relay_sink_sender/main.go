package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
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

func main() {
	count := envOrInt("LOG_RELAY_COUNT", 250)
	interval := envOrDuration("LOG_RELAY_INTERVAL", 20*time.Millisecond)
	if interval <= 0 {
		interval = 20 * time.Millisecond
	}
	runFor := envOrDuration("LOG_RELAY_DURATION", time.Duration(count)*interval+2*time.Second)
	if runFor < 2*time.Second {
		runFor = 2 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), runFor)
	defer cancel()

	logger := builder.NewLogger(
		builder.LoggerWithLevel("info"),
		builder.LoggerWithDevelopment(true),
		builder.LoggerWithFields(map[string]interface{}{
			"service": map[string]string{
				"name":        "log-relay-sender",
				"environment": "local",
			},
		}),
	)
	defer func() {
		_ = logger.Flush()
	}()

	relaySink := builder.SinkConfig{
		Type: string(builder.RelaySink),
		Config: map[string]interface{}{
			"targets":        []string{envOr("LOG_RELAY_ADDR", "localhost:50090")},
			"queue_size":     1024,
			"submit_timeout": "2s",
			"drop_on_full":   true,
			"tls": map[string]interface{}{
				"cert":        envOr("TLS_CERT", "../tls/client.crt"),
				"key":         envOr("TLS_KEY", "../tls/client.key"),
				"ca":          envOr("TLS_CA", "../tls/ca.crt"),
				"server_name": envOr("TLS_SERVER_NAME", "localhost"),
				"min_version": "1.2",
				"max_version": "1.3",
			},
		},
	}
	if err := logger.AddSink("relay", relaySink); err != nil {
		log.Fatalf("add relay sink: %v", err)
	}

	logger.Info("log relay sender up", "event", "startup", "count", count, "interval_ms", interval.Milliseconds())

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	start := time.Now()
	sent := 0
sendLoop:
	for sent < count {
		select {
		case <-ctx.Done():
			break sendLoop
		case t := <-ticker.C:
			sent++
			logger.Info(
				"relay log sample",
				"event", "log_emit",
				"seq", sent,
				"elapsed_ms", t.Sub(start).Milliseconds(),
				"customer_id", fmt.Sprintf("cust-%03d", sent%100),
				"tags", []string{"example", "relay"},
			)
			if sent%50 == 0 {
				logger.Warn("relay log checkpoint", "event", "checkpoint", "seq", sent)
			}
			if sent%100 == 0 {
				logger.Error("relay log sample error", "event", "sample_error", "seq", sent, "error", "simulated")
			}
		}
	}

	logger.Info("log relay sender done", "event", "complete", "sent", sent, "duration_ms", time.Since(start).Milliseconds())
	if err := logger.Flush(); err != nil {
		log.Printf("flush error: %v", err)
	}

	cancel()
	fmt.Println("Done.")
}
