//go:build webtransport

package main

import (
	"context"
	"crypto/tls"
	"fmt"
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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	wire := builder.NewWire[Feedback](ctx, builder.WireWithLogger[Feedback](logger))
	_ = wire.Start(ctx)

	tlsCfg := builder.NewTlsServerConfig(
		true,
		"../tls/server.crt",
		"../tls/server.key",
		"../tls/ca.crt",
		"localhost",
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	recv := builder.NewWebTransportReceivingRelay[Feedback](
		ctx,
		builder.WebTransportReceivingRelayWithAddress[Feedback]("localhost:8443"),
		builder.WebTransportReceivingRelayWithPath[Feedback]("/relay"),
		builder.WebTransportReceivingRelayWithLogger[Feedback](logger),
		builder.WebTransportReceivingRelayWithOutput[Feedback](wire),
		builder.WebTransportReceivingRelayWithTLSConfig[Feedback](tlsCfg),
	)

	if err := recv.Start(ctx); err != nil {
		panic(err)
	}

	fmt.Println("WebTransport receiver up at https://localhost:8443/relay. Ctrl+C to stop.")
	<-ctx.Done()

	recv.Stop()
	wire.Stop()

	fmt.Println("---- Receiver ----")
	if b, err := wire.LoadAsJSONArray(); err == nil {
		fmt.Println(string(b))
	} else {
		fmt.Println("err:", err)
	}
}
