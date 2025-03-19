package main

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// Feedback struct defines the structure of feedback received from customers.
type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

// generator periodically generates and submits feedback, simulating input from an external source.
// The generator now properly handles errors returned by the submit function.
func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, feedback Feedback) error) {
	ticker := time.NewTicker(100 * time.Millisecond) // Increase rate to 10 messages per second
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			feedbackContent := fmt.Sprintf("This is feedback item number %d", count)
			feedback := Feedback{
				CustomerID: fmt.Sprintf("Feedback%d", count),
				Content:    feedbackContent,
				IsNegative: count%10 == 0,
			}
			if err := submitFunc(ctx, feedback); err != nil {
				fmt.Printf("Error submitting feedback: %v\n", err)
			}
			count++
		}
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	logger := builder.NewLogger()

	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
	)

	generatorWire := builder.NewWire(ctx, builder.WireWithGenerator(generator))
	// Attach a generator to the transformer and start the processing.

	tlsConfig := builder.NewTlsClientConfig(
		true,                // UseTLS should be true to use TLS
		"../tls/client.crt", // Path to the client's certificate
		"../tls/client.key", // Path to the client's private key
		"../tls/ca.crt",     // Path to the CA certificate
	)

	forwardRelay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithTarget[Feedback]("localhost:50051"),
		builder.ForwardRelayWithInput(generatorWire),
		builder.ForwardRelayWithLogger[Feedback](logger),
		builder.ForwardRelayWithTLSConfig[Feedback](tlsConfig),
	)

	forwardRelay.Start(ctx)

	// Wait for processing to complete or timeout.
	<-ctx.Done()

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Processing timeout.")
	} else {
		fmt.Println("Processing finished within the allotted time.")
	}
}
