package main

import (
	"context"
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

	recv := builder.NewWebSocketReceivingRelay[Feedback](
		ctx,
		builder.WebSocketReceivingRelayWithAddress[Feedback]("localhost:8084"),
		builder.WebSocketReceivingRelayWithPath[Feedback]("/relay"),
		builder.WebSocketReceivingRelayWithLogger[Feedback](logger),
		builder.WebSocketReceivingRelayWithOutput[Feedback](wire),
	)

	if err := recv.Start(ctx); err != nil {
		panic(err)
	}

	fmt.Println("WebSocket receiver up at ws://localhost:8084/relay. Ctrl+C to stop.")
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
