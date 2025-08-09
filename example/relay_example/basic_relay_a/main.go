package main

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// Feedback defines the structure of feedback received from customers.
type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// Create a wire without a generator
	wire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Feedback](logger),
	)

	// Build the ForwardRelay with the wire as input
	forwardRelay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithTarget[Feedback]("localhost:50052"),
		builder.ForwardRelayWithInput(wire),
		builder.ForwardRelayWithLogger[Feedback](logger),
	)

	// Start the wire and relay
	wire.Start(ctx)
	forwardRelay.Start(ctx)

	// Manually submit feedback to the wire (bypassing Generator/Plug)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Context done, shutting down...")
			forwardRelay.Stop()
			wire.Stop()
			return
		case <-ticker.C:
			fb := Feedback{
				CustomerID: fmt.Sprintf("Feedback%d", count),
				Content:    fmt.Sprintf("This is feedback item number %d", count),
				IsNegative: count%10 == 0,
			}
			if err := wire.Submit(ctx, fb); err != nil {
				fmt.Printf("Error submitting feedback: %v\n", err)
			} else {
				fmt.Printf("Submitted: %+v\n", fb)
			}
			count++
		}
	}
}
