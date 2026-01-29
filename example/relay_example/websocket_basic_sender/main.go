package main

import (
	"context"
	"fmt"
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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	fr := builder.NewWebSocketForwardRelay[Feedback](
		ctx,
		builder.WebSocketForwardRelayWithTarget[Feedback]("ws://localhost:8084/relay"),
		builder.WebSocketForwardRelayWithLogger[Feedback](logger),
		builder.WebSocketForwardRelayWithPayloadFormat[Feedback]("json"),
	)

	for i := 0; i < 5; i++ {
		fb := Feedback{
			CustomerID: fmt.Sprintf("cust-%02d", i),
			Content:    "WebSocket relay payload",
			Category:   "feedback",
			IsNegative: i%2 == 0,
			Tags:       []string{"ws"},
		}
		if err := fr.Submit(ctx, fb); err != nil {
			logger.Error("Submit failed", "error", err, "seq", i)
		}
		time.Sleep(100 * time.Millisecond)
	}
	logger.Info("Send complete")
}
