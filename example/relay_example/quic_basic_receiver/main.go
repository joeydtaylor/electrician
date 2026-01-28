package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	wire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithLogger[Feedback](logger),
	)
	if err := wire.Start(ctx); err != nil {
		log.Fatalf("wire start: %v", err)
	}

	tlsCfg := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS13, tls.VersionTLS13,
	)

	recv := builder.NewQuicReceivingRelay[Feedback](
		ctx,
		builder.QuicReceivingRelayWithAddress[Feedback](envOr("RX_ADDR", "localhost:50071")),
		builder.QuicReceivingRelayWithLogger[Feedback](logger),
		builder.QuicReceivingRelayWithOutput[Feedback](wire),
		builder.QuicReceivingRelayWithTLSConfig[Feedback](tlsCfg),
	)

	if err := recv.Start(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("QUIC receiver up. Ctrl+C to stop.")
	<-ctx.Done()
	fmt.Println("Shutting down...")

	recv.Stop()

	if b, err := wire.LoadAsJSONArray(); err == nil {
		fmt.Println("---- Receiver ----")
		fmt.Println(string(b))
	}
}
