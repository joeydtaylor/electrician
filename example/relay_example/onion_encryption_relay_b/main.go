package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
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

// Two separate AES-256 keys (hex). The wireKey is for the wire encryption.
// The relayKey is for the forward relay’s second layer encryption.
const WireAES256KeyHex = "00112233445566778899aaabbbcccdddeeff00112233445566778899aaabbbcc"
const RelayAES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

// loadKey decodes a hex string into 32 bytes => AES-256
func loadKey(hexStr string) string {
	b, err := hex.DecodeString(hexStr)
	if err != nil {
		log.Fatalf("Failed to decode hex key: %v", err)
	}
	if len(b) != 32 {
		log.Fatalf("Key must be 32 bytes for AES-256, got %d bytes", len(b))
	}
	return string(b)
}

func main() {
	// 1) Set up a context that we can cancel on SIGINT or SIGTERM for graceful shutdown
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

	// 2) A simple logger
	logger := builder.NewLogger()
	sec := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)

	// 3) A wire that will apply a final domain transform to the unwrapped data
	//    For instance, sentiment analysis. T=Feedback
	sentimentWire := builder.NewWire(
		ctx,
		// Possibly we do a final decrypt if there's a second layer – see "WireWithDecryptOptions[Feedback](…)" if so.
		builder.WireWithTransformer(sentimentAnalyzer),
	)

	// 5) TLS config for receiving – only TLS 1.3
	tlsConfig := builder.NewTlsServerConfig(
		true,                // UseTLS
		"../tls/server.crt", // Path to server certificate
		"../tls/server.key", // Path to server private key
		"../tls/ca.crt",     // Path to CA certificate
		"localhost",         // SubjectAlternativeName
		tls.VersionTLS13,    // MinTLSVersion
		tls.VersionTLS13,    // MaxTLSVersion
	)

	// 6) Convert the hex key to raw bytes -> if your code expects a string, cast it to string
	aesWireKeyStr := loadKey(WireAES256KeyHex)
	// 6) Convert the hex key to raw bytes -> if your code expects a string, cast it to string
	aesRelayKeyStr := loadKey(RelayAES256KeyHex)

	// 8) (Optional) A second receiving relay on port 50052 for the same logic
	//    that also outputs to a different wire or the same sentimentWire
	secondaryWire := builder.NewWire(ctx, builder.WireWithDecryptOptions[string](sec, aesWireKeyStr))
	receivingRelay := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[string]("localhost:50052"),
		builder.ReceivingRelayWithBufferSize[string](1000),
		builder.ReceivingRelayWithLogger[string](logger),
		builder.ReceivingRelayWithOutput(secondaryWire),
		builder.ReceivingRelayWithTLSConfig[string](tlsConfig),
		builder.ReceivingRelayWithDecryptionKey[string](aesRelayKeyStr),
	)

	receivingRelay.Start(ctx)
	sentimentWire.Start(ctx)

	fmt.Println("Receiving relays started. Press Ctrl+C to stop...")

	// Block until canceled
	<-ctx.Done()

	// Attempt to gather results from outputConduit.
	// The data on secondaryWire is also available if you want to do something with it.
	outputJSON, err := secondaryWire.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	// Unmarshal the JSON array into a slice of strings
	var feedbackStrings []string
	if err := json.Unmarshal(outputJSON, &feedbackStrings); err != nil {
		fmt.Printf("Error unmarshalling output JSON array: %v\n", err)
		return
	}

	// Now, unmarshal each individual JSON string into a Feedback struct
	var feedbackList []Feedback
	for _, fbStr := range feedbackStrings {
		var fb Feedback
		if err := json.Unmarshal([]byte(fbStr), &fb); err != nil {
			fmt.Printf("Error unmarshalling feedback: %v\n", err)
			return
		}
		feedbackList = append(feedbackList, fb)
	}

	// Output the unmarshalled feedback structs
	fmt.Println("Feedback Analysis Summary:")
	for _, fb := range feedbackList {
		fmt.Printf("%+v\n", fb)
	}

	fmt.Println("Shutting down gracefully.")
}
