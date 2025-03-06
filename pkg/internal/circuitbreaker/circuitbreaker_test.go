package circuitbreaker_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// Feedback type for testing
type Feedback struct {
	CustomerID string
	Content    string
}

// Simulates an error if the feedback content contains the word "error".
func errorSimulator(feedback Feedback) (Feedback, error) {
	if strings.Contains(strings.ToLower(feedback.Content), "error") {
		return Feedback{}, errors.New("processing error")
	}
	return feedback, nil
}

func TestCircuitBreakerOpen(t *testing.T) {
	// Setup context with a deadline to prevent the test from running indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize the circuit breaker with a threshold of 1 error and a reset timeout of 2 seconds
	circuitBreaker := builder.NewCircuitBreaker[Feedback](ctx, 1, 2*time.Second)

	// Setup a simple wire that utilizes the circuit breaker
	wire := builder.NewWire[Feedback](ctx)
	wire.ConnectCircuitBreaker(circuitBreaker)

	// Set the wire's processing function to one that simulates an error
	wire.ConnectTransformer(errorSimulator)

	// Start processing
	wire.Start(ctx)

	// Submit a piece of feedback to trigger the circuit breaker
	wire.Submit(ctx, Feedback{CustomerID: "Test1", Content: "Trigger error"})

	// Allow some time for the circuit breaker to respond to the error
	time.Sleep(1 * time.Second)

	// Attempt to submit another piece of feedback. This should be blocked by the circuit breaker.
	wire.Submit(ctx, Feedback{CustomerID: "Test2", Content: "Second submission"})

	// Wait a bit to see if the circuit breaker resets
	time.Sleep(3 * time.Second)

	// Check the status of the circuit breaker
	if !circuitBreaker.Allow() {
		t.Error("Expected the circuit breaker to be open")
	}

	// Cleanup
	wire.Stop()
}

func TestCircuitBreakerClose(t *testing.T) {
	// Setup context with a deadline to prevent the test from running indefinitely
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize the circuit breaker with a threshold of 1 error and a reset timeout of 2 seconds
	circuitBreaker := builder.NewCircuitBreaker[Feedback](ctx, 1, 2*time.Second)

	// Setup a simple wire that utilizes the circuit breaker
	wire := builder.NewWire[Feedback](ctx)
	wire.ConnectCircuitBreaker(circuitBreaker)

	// Set the wire's processing function to one that simulates an error
	wire.ConnectTransformer(errorSimulator)

	// Start processing
	wire.Start(ctx)

	// Submit a piece of feedback to trigger the circuit breaker
	wire.Submit(ctx, Feedback{CustomerID: "Test1", Content: "Trigger error"})

	// Allow some time for the circuit breaker to respond to the error
	time.Sleep(1 * time.Second)

	// Attempt to submit another piece of feedback. This should be blocked by the circuit breaker.
	wire.Submit(ctx, Feedback{CustomerID: "Test2", Content: "Second submission"})

	// Check the status of the circuit breaker
	if circuitBreaker.Allow() {
		t.Error("Expected the circuit breaker to be closed")
	}

	// Cleanup
	wire.Stop()
}
