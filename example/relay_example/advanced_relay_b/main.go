package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// Feedback is the same struct used by the forward relay
type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

// sentimentAnalyzer is just an example transformer for demonstration.
func sentimentAnalyzer(f Feedback) (Feedback, error) {
	positiveWords := []string{"love", "great", "happy"}
	for _, word := range positiveWords {
		if strings.Contains(strings.ToLower(f.Content), word) {
			f.Tags = append(f.Tags, "Positive Sentiment")
			return f, nil
		}
	}
	f.Tags = append(f.Tags, "Needs Attention")
	return f, nil
}

// For example/demo only; using the same key as the forward relay. In practice,
// you’d store or fetch this from a secure location (KMS, environment, etc.).
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
	// Set up a context that we can cancel on SIGINT or SIGTERM for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		fmt.Println("Received termination signal, shutting down...")
		cancel()
	}()

	// A simple logger
	logger := builder.NewLogger()

	// Create a wire that applies a transformer to the unwrapped data
	sentimentWire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(sentimentAnalyzer),
	)
	// Build a conduit that can convert final data to JSON if needed
	outputConduit := builder.NewConduit(
		ctx,
		builder.ConduitWithWire(sentimentWire),
	)

	// TLS config for receiving – only TLS 1.3
	tlsConfig := builder.NewTlsServerConfig(
		true,                // UseTLS
		"../tls/server.crt", // Path to server certificate
		"../tls/server.key", // Path to server private key
		"../tls/ca.crt",     // Path to CA certificate
		"localhost",         // SubjectAlternativeName
		tls.VersionTLS13,    // MinTLSVersion
		tls.VersionTLS13,    // MaxTLSVersion
	)

	// Convert the hex key to raw bytes -> if your code expects a string, cast it to string
	decryptionKeyBytes := loadExampleAESKey()
	// Our code expects a string of 32 bytes => AES-256
	decryptionKey := string(decryptionKeyBytes)

	// Build the first receiving relay that listens on localhost:50051
	receivingRelay := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50051"),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(outputConduit),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsConfig),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decryptionKey),
	)

	// (Optional) A second receiving relay on localhost:50052,
	// also decrypting with the same key and passing data to a different wire or output
	secondaryWire := builder.NewWire[Feedback](ctx)
	secondReceivingRelay := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50052"),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(secondaryWire),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsConfig),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decryptionKey),
	)

	// Start both
	receivingRelay.Start(ctx)
	secondReceivingRelay.Start(ctx)

	fmt.Println("Receiving relays started. Press Ctrl+C to stop...")

	// Block until canceled
	<-ctx.Done()

	// Attempt to gather results from outputConduit
	// The data on secondaryWire is also available if you want to do something with it
	outputJSON, err := outputConduit.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("Feedback Analysis Summary:")
	fmt.Println(string(outputJSON))
	fmt.Println("Shutting down gracefully.")
}
