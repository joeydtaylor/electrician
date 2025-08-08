package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
	"google.golang.org/grpc/metadata"
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

// ----- OAuth2 helpers -----

type introspectResp struct {
	Active bool   `json:"active"`
	Scope  string `json:"scope"`
}

// streamAuthValidator returns a validator that runs for each request/stream.
// It reads the Bearer token from metadata, calls Aegis introspection with the
// provided client credentials, and enforces required scopes.
func streamAuthValidator(
	introspectionURL string,
	clientID string,
	clientSecret string,
	requiredScopes []string,
) func(ctx context.Context, md map[string]string) error {

	// Scope set check
	hasAllScopes := func(granted string) bool {
		if len(requiredScopes) == 0 {
			return true
		}
		parts := strings.Fields(granted) // space-separated per RFC
		set := make(map[string]struct{}, len(parts))
		for _, s := range parts {
			set[s] = struct{}{}
		}
		for _, need := range requiredScopes {
			if _, ok := set[need]; !ok {
				return false
			}
		}
		return true
	}

	// Dev HTTP client: TLS1.3 + accept self-signed localhost cert
	hc := &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				MaxVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true, // dev only
			},
		},
	}

	return func(ctx context.Context, userMD map[string]string) error {
		// Combine metadata from context + provided map (so either source works)
		var token string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if vals := md.Get("authorization"); len(vals) > 0 {
				// Prefer actual Authorization header if present
				if strings.HasPrefix(strings.ToLower(vals[0]), "bearer ") {
					token = strings.TrimSpace(vals[0][len("Bearer "):])
				}
			}
		}
		if token == "" {
			// Fallback: builder may map headers into userMD
			if v, ok := userMD["authorization"]; ok && strings.HasPrefix(strings.ToLower(v), "bearer ") {
				token = strings.TrimSpace(v[len("bearer "):])
			}
		}
		if token == "" {
			return fmt.Errorf("unauthenticated: missing bearer token")
		}

		// RFC 7662 introspection
		form := url.Values{}
		form.Set("token", token)

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectionURL, strings.NewReader(form.Encode()))
		if err != nil {
			return fmt.Errorf("introspection req: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth(clientID, clientSecret) // will fail if secret is wrong

		resp, err := hc.Do(req)
		if err != nil {
			return fmt.Errorf("introspection http: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return fmt.Errorf("introspection status %d", resp.StatusCode)
		}

		var ir introspectResp
		if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
			return fmt.Errorf("introspection decode: %w", err)
		}
		if !ir.Active {
			return fmt.Errorf("unauthenticated: token inactive")
		}
		if !hasAllScopes(ir.Scope) {
			return fmt.Errorf("permission denied: insufficient scope")
		}
		return nil
	}
}

// -------------------------------------------------------------------

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

	// ---- OAuth2 (Aegis) resource server config via builder helpers ----
	introspect := builder.NewReceivingRelayOAuth2IntrospectionOptions(
		"https://localhost:3000/api/auth/oauth/token/introspect",
		"basic",            // client auth type for introspection
		"steeze-local-cli", // seeded confidential client_id
		"local-secret",     // <-- wrong on purpose to prove enforcement
		"",                 // bearer token if using "bearer" auth type (not used here)
		60,                 // cache successful introspection responses (seconds)
	)
	authOpts := builder.NewReceivingRelayAuthenticationOptionsOAuth2(introspect)

	// Dynamic validator that **actually enforces** on streaming RPCs
	requiredScopes := []string{"write:data"} // scopes required for this relay
	validator := streamAuthValidator(
		"https://localhost:3000/api/auth/oauth/introspect",
		"steeze-local-cli",
		"local-secret", // wrong on purpose; should trigger 401/403 behavior
		requiredScopes,
	)

	// Build the first receiving relay that listens on localhost:50051
	receivingRelay := builder.NewReceivingRelay(
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50051"),
		builder.ReceivingRelayWithBufferSize[Feedback](1000), // Buffer size for data channel
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(outputConduit),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsConfig),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decryptionKey),

		// Hints (for visibility) + **actual enforcement** via dynamic validator
		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ReceivingRelayWithDynamicAuthValidator[Feedback](validator),
	)

	// (Optional) A second receiving relay on localhost:50052,
	// also decrypting with the same key and passing data to a different wire or output
	secondaryWire := builder.NewWire[Feedback](ctx)
	secondReceivingRelay := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback]("localhost:50052"),
		builder.ReceivingRelayWithBufferSize[Feedback](1000), // Example buffer size
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(secondaryWire),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsConfig),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decryptionKey),

		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ReceivingRelayWithDynamicAuthValidator[Feedback](validator),
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
