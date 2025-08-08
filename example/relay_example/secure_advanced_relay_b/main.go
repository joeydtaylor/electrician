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

func sentimentAnalyzer(f Feedback) (Feedback, error) {
	positiveWords := []string{"love", "great", "happy"}
	for _, w := range positiveWords {
		if strings.Contains(strings.ToLower(f.Content), w) {
			f.Tags = append(f.Tags, "Positive Sentiment")
			return f, nil
		}
	}
	f.Tags = append(f.Tags, "Needs Attention")
	return f, nil
}

const AES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

func loadExampleAESKey() []byte {
	raw, err := hex.DecodeString(AES256KeyHex)
	if err != nil {
		log.Fatalf("Failed to decode example AES-256 key: %v", err)
	}
	return raw
}

func main() {
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

	logger := builder.NewLogger()

	// Processing pipeline
	sentimentWire := builder.NewWire(
		ctx,
		builder.WireWithTransformer(sentimentAnalyzer),
	)
	outputConduit := builder.NewConduit(
		ctx,
		builder.ConduitWithWire(sentimentWire),
	)

	// TLS 1.3
	tlsConfig := builder.NewTlsServerConfig(
		true,
		"../tls/server.crt",
		"../tls/server.key",
		"../tls/ca.crt",
		"localhost",
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	decryptionKey := string(loadExampleAESKey())

	// ---- Built-in OAuth2 resource-server config (NO custom validator) ----
	const issuerBase = "https://localhost:3000"
	const jwksURL = "https://localhost:3000/api/auth/.well-known/jwks.json"
	const introspectURL = "https://localhost:3000/api/auth/oauth/introspect"

	requiredAud := []string{"your-api"}
	requiredScopes := []string{"write:data"}

	jwtOpts := builder.NewReceivingRelayOAuth2JWTOptions(
		issuerBase,
		jwksURL,
		requiredAud,
		requiredScopes,
		/*jwksCacheSeconds*/ 300,
	)

	introspectOpts := builder.NewReceivingRelayOAuth2IntrospectionOptions(
		introspectURL,
		"basic",
		"steeze-local-cli", // your client that’s allowed to introspect
		"local-secret",
		"", // no bearer needed for "basic"
		/*cacheSeconds*/ 300, // cache successful results
	)

	// Merge: accept JWT (preferred) and allow introspection as fallback — both with required scopes/aud
	oauth := builder.NewReceivingRelayMergeOAuth2Options(jwtOpts, introspectOpts)
	authOpts := builder.NewReceivingRelayAuthenticationOptionsOAuth2(oauth)

	// First receiving relay
	receivingRelay := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50051"),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(outputConduit),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsConfig),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decryptionKey),

		// Built-in auth (no custom validator!)
		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
	)

	// Optional second receiving relay
	secondaryWire := builder.NewWire[Feedback](ctx)
	secondReceivingRelay := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50052"),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(secondaryWire),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsConfig),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decryptionKey),

		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
	)

	// Start both
	receivingRelay.Start(ctx)
	secondReceivingRelay.Start(ctx)

	fmt.Println("Receiving relays started. Press Ctrl+C to stop...")

	<-ctx.Done()

	// Drain and print summary
	outputJSON, err := outputConduit.LoadAsJSONArray()
	if err != nil {
		fmt.Printf("Error converting output to JSON: %v\n", err)
		return
	}
	fmt.Println("Feedback Analysis Summary:")
	fmt.Println(string(outputJSON))
	fmt.Println("Shutting down gracefully.")
}
