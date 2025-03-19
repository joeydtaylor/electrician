package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math/rand"
	"strings"
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

// errorSimulator simulates a processing error if the feedback content contains the word "error".
func errorSimulator(feedback Feedback) (Feedback, error) {
	if strings.Contains(strings.ToLower(feedback.Content), "error") {
		return Feedback{}, errors.New("simulated processing error")
	}
	return feedback, nil
}

// generator periodically generates and submits feedback, simulating input from an external source.
func plugFunc(ctx context.Context, submitFunc func(ctx context.Context, feedback Feedback) error) {
	ticker := time.NewTicker(1000 * time.Millisecond) // Increase rate to 10 messages per second
	defer ticker.Stop()

	count := 0
	for {
		select {
		case <-ctx.Done():
			// If the context is done, stop generating feedback.
			return
		case <-ticker.C:
			// Randomly decide to introduce an error in the feedback content
			errorTrigger := ""
			if rand.Intn(10) == 0 { // Approximately 10% chance to trigger an error
				errorTrigger = " error "
			}

			feedbackContent := fmt.Sprintf("This is feedback item number %d%s", count, errorTrigger)
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
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
	)

	// Initialize components of the processing pipeline: circuit breaker, transformers, and groundWire.
	groundWire := builder.NewWire[Feedback](ctx)
	circuitBreaker := builder.NewCircuitBreaker(
		ctx,
		1, 10*time.Second,
		builder.CircuitBreakerWithNeutralWire(groundWire),
		builder.CircuitBreakerWithLogger[Feedback](logger),
	)
	generatorWire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Feedback](logger),
		builder.WireWithConcurrencyControl[Feedback](1000, 100),
		builder.WireWithCircuitBreaker(circuitBreaker),
		builder.WireWithTransformer(errorSimulator),
		builder.WireWithGenerator(generator),
	)

	// Enforce TLS 1.3
	tlsConfig := builder.NewTlsClientConfig(
		true,                // UseTLS should be true to use TLS
		"../tls/client.crt", // Path to the client's certificate
		"../tls/client.key", // Path to the client's private key
		"../tls/ca.crt",     // Path to the CA certificate
		tls.VersionTLS13,    // MinVersion: Only allow TLS 1.3
		tls.VersionTLS13,    // MaxVersion: Only allow TLS 1.3
	)

	forwardRelay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithLogger[Feedback](logger),
		builder.ForwardRelayWithTarget[Feedback]("localhost:50051", "localhost:50052"),
		builder.ForwardRelayWithInput(groundWire),
		builder.ForwardRelayWithTLSConfig[Feedback](tlsConfig),
	)

	generatorWire.Start(ctx)
	forwardRelay.Start(ctx)

	// Wait for processing to complete or timeout.
	<-ctx.Done()

	// Clean up and terminate processing.
	generatorWire.Stop()
	forwardRelay.Stop()

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Processing timeout.")
	} else {
		fmt.Println("Processing finished within the allotted time.")
	}
}
