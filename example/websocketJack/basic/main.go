package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Message struct {
	ID   int    `json:"id"`
	Body string `json:"body"`
}

func toUpper(msg Message) (Message, error) {
	msg.Body = strings.ToUpper(msg.Body)
	return msg, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		fmt.Println("[signal] shutting down...")
		cancel()
	}()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))
	sensor := builder.NewSensor(builder.SensorWithLogger[Message](logger))

	wire := builder.NewWire(
		ctx,
		builder.WireWithSensor(sensor),
		builder.WireWithTransformer(toUpper),
	)
	if err := wire.Start(ctx); err != nil {
		fmt.Printf("wire start error: %v\n", err)
		return
	}

	server := builder.NewWebSocketServer(
		ctx,
		builder.WebSocketServerWithLogger[Message](logger),
		builder.WebSocketServerWithSensor[Message](sensor),
		builder.WebSocketServerWithAddress[Message](":8080"),
		builder.WebSocketServerWithEndpoint[Message]("/ws"),
		builder.WebSocketServerWithMessageFormat[Message]("json"),
		builder.WebSocketServerWithOutputWire[Message](wire),
	)

	go func() {
		if err := server.Serve(ctx, wire.Submit); err != nil && ctx.Err() == nil {
			fmt.Printf("server error: %v\n", err)
		}
	}()

	fmt.Println("WebSocket server listening on ws://localhost:8080/ws")
	fmt.Println("Send JSON like: {\"id\": 1, \"body\": \"hello\"}")

	<-ctx.Done()
	_ = wire.Stop()
}
