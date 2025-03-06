package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

func sentimentAnalyzer(feedback Feedback) (Feedback, error) {
	positiveWords := []string{"love", "great", "happy"}
	for _, word := range positiveWords {
		if strings.Contains(strings.ToLower(feedback.Content), word) {
			feedback.Tags = append(feedback.Tags, "Positive Sentiment")
			return feedback, nil
		}
	}
	feedback.Tags = append(feedback.Tags, "Needs Attention")
	return feedback, nil
}

func main() {
	// Setup cancellation context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle termination signals to gracefully shut down the relay
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		fmt.Println("Received termination signal, shutting down...")
		cancel()
	}()

	logger := builder.NewLogger()

	// Create the TLS configuration
	tlsConfig := builder.NewTlsServerConfig(
		true,                // UseTLS - Set this to 'true'
		"../tls/server.crt", // CertFile - Update paths as needed
		"../tls/server.key", // KeyFile - Update paths as needed
		"../tls/ca.crt",     // CAFile - Update paths as needed
		"localhost",
	)

	sentimentWire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithTransformer[Feedback](sentimentAnalyzer),
	)
	outputConduit := builder.NewConduit[Feedback](
		ctx,
		builder.ConduitWithWire[Feedback](sentimentWire),
	)

	// Receiving Relay that uses the second conduit
	receivingRelay := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50051"),
		builder.ReceivingRelayWithBufferSize[Feedback](10000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput[Feedback](outputConduit),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsConfig),
	)

	// Start the receiving relay
	receivingRelay.Start(ctx)

	// Block until the context is canceled
	<-ctx.Done()

	// Attempt to aggregate results after cancellation
	output, err := outputConduit.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}

	fmt.Println("Feedback Analysis Summary:")
	fmt.Println(string(output))
}
