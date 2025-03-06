package wire_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/codec"
	"github.com/joeydtaylor/electrician/pkg/internal/wire"
)

// TestWire_Start verifies that a Wire instance can be started successfully.
// It creates a Wire with an immediately canceled context and then checks that the component metadata is properly set.
func TestWire_Start(t *testing.T) {
	// Create a context that is immediately canceled.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Create a new Wire instance.
	w := wire.NewWire[int](ctx)

	// Start the Wire.
	w.Start(ctx)

	// Wait until the context is done.
	<-ctx.Done()

	// Validate that the Wire's component metadata has a non-empty ID.
	if len(w.GetComponentMetadata().ID) == 0 {
		t.Error("Wire should have started successfully; component metadata ID is empty")
	}
}

// TestWire_Submit verifies the submission and processing of elements in the Wire.
// It sets up a Wire with a LineEncoder and a transformer that simulates an error for input 42.
// Then, it submits elements and checks that a valid element is processed despite the simulated error.
func TestWire_Submit(t *testing.T) {
	// Create a context with a timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Create a new LineEncoder for integer values.
	encoder := codec.NewLineEncoder[int]() // Using integer type for simplicity.

	// Create a new Wire instance with the provided encoder.
	w := wire.NewWire[int](ctx, wire.WithEncoder[int](encoder))
	// Attach a transformer that simulates an error when the input equals 42.
	w.ConnectTransformer(func(input int) (int, error) {
		if input == 42 {
			return 0, fmt.Errorf("simulated error")
		}
		return input, nil
	})

	// Start the Wire.
	w.Start(ctx)

	// Submit an element that causes an error in the transformer.
	w.Submit(ctx, 42)

	// Submit a valid element.
	w.Submit(ctx, 24)

	// Wait for the context to complete.
	<-ctx.Done()

	// Stop the Wire.
	w.Stop()

	// Retrieve the output buffer as a string.
	outputString := w.GetOutputBuffer().String()

	// Verify that the valid element ("24") was processed.
	expectedOutput := "24"
	if !strings.Contains(outputString, expectedOutput) {
		t.Errorf("Expected output to contain '%s', got '%s'", expectedOutput, outputString)
	}
}

// TestWire_Terminate verifies that a Wire instance terminates correctly.
// It starts the Wire and then stops it, ensuring that the output channel is closed.
func TestWire_Terminate(t *testing.T) {
	// Create a new Wire instance using a background context.
	w := wire.NewWire[int](context.Background())

	// Start the Wire.
	w.Start(context.Background())

	// Stop the Wire.
	w.Stop()

	// Check if the output channel is closed.
	_, ok := <-w.GetOutputChannel()
	if ok {
		t.Error("Expected output channel to be closed after Wire termination")
	}
}
