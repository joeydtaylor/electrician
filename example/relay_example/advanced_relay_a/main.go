package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
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
	ticker := time.NewTicker(1000 * time.Millisecond)
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
			if rand.Intn(10) == 0 { // ~10% chance to trigger an error
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

// For example/demo only; you'd never embed a real key like this in production code.
const AES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

// loadExampleAESKey decodes the hex-encoded 256-bit key into raw bytes.
func loadExampleAESKey() []byte {
	raw, err := hex.DecodeString(AES256KeyHex)
	if err != nil {
		log.Fatalf("Failed to decode example AES-256 key: %v", err)
	}
	return raw
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// Step 1: Create an adapter (plug) function that periodically generates data.
	plug := builder.NewPlug(
		ctx,
		builder.PlugWithAdapterFunc(plugFunc),
	)

	// Step 2: Create a generator that uses the plug
	generator := builder.NewGenerator(
		ctx,
		builder.GeneratorWithPlug(plug),
	)

	// Step 3: Create a groundWire + circuitBreaker pipeline
	groundWire := builder.NewWire[Feedback](ctx)
	circuitBreaker := builder.NewCircuitBreaker(
		ctx,
		1, 10*time.Second,
		builder.CircuitBreakerWithNeutralWire(groundWire),
		builder.CircuitBreakerWithLogger[Feedback](logger),
	)
	// The main wire that processes data + passes it to the circuit breaker
	generatorWire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Feedback](logger),
		builder.WireWithConcurrencyControl[Feedback](1000, 100),
		builder.WireWithCircuitBreaker(circuitBreaker),
		builder.WireWithTransformer(errorSimulator),
		builder.WireWithGenerator(generator),
	)

	// Step 4: Configure TLS for only TLS1.3
	tlsConfig := builder.NewTlsClientConfig(
		true,                // useTLS
		"../tls/client.crt", // path to client's certificate
		"../tls/client.key", // path to client's private key
		"../tls/ca.crt",     // path to CA certificate
		tls.VersionTLS13,    // min version
		tls.VersionTLS13,    // max version
	)

	// Step 5: Create performance + security (AES-GCM) options
	perfOptions := builder.NewPerformanceOptions(true, builder.COMPRESS_SNAPPY)
	secOptions := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM) // Enable AES-GCM

	// Convert the hex key to a raw string (16, 24, or 32 bytes) for AES-GCM usage
	encryptionKeyBytes := loadExampleAESKey()
	// If your encryption code expects a string of length 32, do:
	encryptionKey := string(encryptionKeyBytes) // This is 32 bytes => AES-256

	// Step 6: Create a forward relay with compression + encryption
	forwardRelay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithLogger[Feedback](logger),
		builder.ForwardRelayWithTarget[Feedback]("localhost:50051", "localhost:50052"),
		builder.ForwardRelayWithPerformanceOptions[Feedback](perfOptions),

		// The new security config
		builder.ForwardRelayWithSecurityOptions[Feedback](secOptions, encryptionKey),

		builder.ForwardRelayWithInput(groundWire),
		builder.ForwardRelayWithTLSConfig[Feedback](tlsConfig),
	)

	// Step 7: Start the pipeline
	generatorWire.Start(ctx)
	forwardRelay.Start(ctx)

	// Step 8: Wait until context times out or is canceled
	<-ctx.Done()

	// Clean up
	generatorWire.Stop()
	forwardRelay.Stop()

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Processing timeout.")
	} else {
		fmt.Println("Processing finished within the allotted time.")
	}
}
