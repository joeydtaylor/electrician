package main

import (
	"context"
	"crypto/tls"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// Domain struct
type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

// A sample transform
func errorSimulator(feedback Feedback) (Feedback, error) {
	if strings.Contains(strings.ToLower(feedback.Content), "error") {
		return Feedback{}, fmt.Errorf("simulated processing error")
	}
	return feedback, nil
}

// Example plug generator
func plugFunc(ctx context.Context, submitFunc func(context.Context, Feedback) error) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	count := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fb := Feedback{
				CustomerID: fmt.Sprintf("Feedback%d", count),
				Content:    fmt.Sprintf("This is feedback item %d", count),
				IsNegative: count%10 == 0,
			}
			if err := submitFunc(ctx, fb); err != nil {
				fmt.Printf("Error submitting feedback: %v\n", err)
			}
			count++
		}
	}
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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigch
		fmt.Println("Got termination signal, shutting down.")
		cancel()
	}()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))
	gob.Register(Feedback{}) // register domain type for GOB if needed

	///////////////////////////////////////////////////////////////////////
	// (A) Build a generator producing domain Feedback
	///////////////////////////////////////////////////////////////////////
	plug := builder.NewPlug(ctx, builder.PlugWithAdapterFunc(plugFunc))
	generator := builder.NewGenerator(ctx, builder.GeneratorWithPlug(plug))

	///////////////////////////////////////////////////////////////////////
	// (B) A wire typed as Feedback => domain transforms
	///////////////////////////////////////////////////////////////////////
	domainWire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Feedback](logger),
		builder.WireWithGenerator(generator),
		builder.WireWithTransformer(errorSimulator),
	)

	///////////////////////////////////////////////////////////////////////
	// (C) A second wire typed as string => 1st encryption layer
	///////////////////////////////////////////////////////////////////////
	encryptionWire := builder.NewWire(ctx,
		builder.WireWithLogger[string](logger),
	)

	// Performance + security for the wire's encryption
	perfOptions := builder.NewPerformanceOptions(true, builder.COMPRESS_SNAPPY)
	wireSecOpts := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)
	wireAESKey := loadKey(WireAES256KeyHex)

	// The wire uses the wireAESKey for encryption
	builder.WireWithEncryptOptions[string](wireSecOpts, wireAESKey)(encryptionWire)
	builder.WireWithConcurrencyControl[string](1000, 100)(encryptionWire)

	// bridging domainWire => encryptionWire
	go func() {
		for fb := range domainWire.GetOutputChannel() {
			// Convert domain object => JSON => encryptionWire
			data := feedbackToJSON(fb)
			encryptionWire.Submit(ctx, data)
		}
	}()

	// Start domain wire + encryption wire
	domainWire.Start(ctx)
	encryptionWire.Start(ctx)

	///////////////////////////////////////////////////////////////////////
	// (D) The forward relay typed as string => 2nd encryption layer
	///////////////////////////////////////////////////////////////////////
	tlsConfig := builder.NewTlsClientConfig(
		true,
		"../tls/client.crt",
		"../tls/client.key",
		"../tls/ca.crt",
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	// The forward relay uses a second, distinct key
	relaySecOpts := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)
	relayAESKey := loadKey(RelayAES256KeyHex)

	forwardRelay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithLogger[string](logger),
		builder.ForwardRelayWithTarget[string]("localhost:50052"),
		builder.ForwardRelayWithPerformanceOptions[string](perfOptions),
		// The second encryption layer => onion
		builder.ForwardRelayWithSecurityOptions[string](relaySecOpts, relayAESKey),
		builder.ForwardRelayWithInput(encryptionWire),
		builder.ForwardRelayWithTLSConfig[string](tlsConfig),
	)

	forwardRelay.Start(ctx)

	fmt.Println("All wires & forward relay started. Press Ctrl+C to stop.")
	<-ctx.Done()

	// Stop everything
	forwardRelay.Stop()
	encryptionWire.Stop()
	domainWire.Stop()

	fmt.Println("Shutdown complete.")
}

func feedbackToJSON(f Feedback) string {
	return fmt.Sprintf(`{"CustomerID":%q,"Content":%q,"IsNegative":%v}`,
		f.CustomerID, f.Content, f.IsNegative)
}
