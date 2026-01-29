//go:build nats

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

	fr := builder.NewNATSForwardRelay[Feedback](
		ctx,
		builder.NATSForwardRelayWithURL[Feedback]("nats://localhost:4222"),
		builder.NATSForwardRelayWithSubject[Feedback]("relay.feedback"),
		builder.NATSForwardRelayWithPayloadFormat[Feedback]("json"),
	)

	for i := 0; i < 5; i++ {
		fb := Feedback{
			CustomerID: fmt.Sprintf("cust-%02d", i),
			Content:    "NATS relay payload",
			Category:   "feedback",
			IsNegative: i%2 == 0,
			Tags:       []string{"nats"},
		}
		if err := fr.Submit(ctx, fb); err != nil {
			logger.Error("Publish failed", "error", err, "seq", i)
		}
		time.Sleep(100 * time.Millisecond)
	}
	logger.Info("Publish complete")
}
