package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

/* ---------------- Local transform (demo) ---------------- */
func errorSimulator(feedback Feedback) (Feedback, error) {
	if strings.Contains(strings.ToLower(feedback.Content), "error") {
		return Feedback{}, errors.New("simulated processing error")
	}
	return feedback, nil
}

/* ---------------- Helpers ---------------- */
func mustHexKey(hexKey string) string {
	raw, err := hex.DecodeString(hexKey)
	if err != nil {
		log.Fatalf("bad hex key: %v", err)
	}
	if len(raw) != 32 {
		log.Fatalf("AES key must be 32 bytes (got %d)", len(raw))
	}
	return string(raw)
}

func pretty(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}

func dumpJWT(token string) {
	log.Printf("RAW ACCESS TOKEN:\n%s\n", token)

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		log.Printf("Not a JWT (parts=%d)", len(parts))
		return
	}
	dec := func(s string) []byte {
		b, err := base64.RawURLEncoding.DecodeString(s)
		if err != nil {
			return []byte(fmt.Sprintf(`"<decode error: %v>"`, err))
		}
		return b
	}
	var hdr map[string]any
	var claims map[string]any
	_ = json.Unmarshal(dec(parts[0]), &hdr)
	_ = json.Unmarshal(dec(parts[1]), &claims)
	log.Printf("JWT HEADER:\n%s", pretty(hdr))
	log.Printf("JWT CLAIMS:\n%s", pretty(claims))
}

func httpClientTLSInsecure() *http.Client {
	return &http.Client{
		Timeout: 6 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				MaxVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true, // dev only
			},
		},
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

/* ---------------- Random feedback generator ---------------- */

var (
	posPhrases = []string{
		"love this", "great service", "excellent experience", "super fast",
		"fantastic support", "works beautifully", "very satisfied",
	}
	negPhrases = []string{
		"not impressed", "slow and buggy", "bad ux", "confusing flow",
		"kept crashing", "too many errors", "support unhelpful",
	}
	neutralPhrases = []string{
		"okay", "fine", "could be better", "average", "meh",
	}
	categories = []string{"billing", "shipping", "ux", "performance", "support", "feature"}
	tagPool    = []string{"vip", "beta", "mobile", "web", "priority", "returning"}
)

func randomChoice(r *rand.Rand, arr []string) string {
	return arr[r.Intn(len(arr))]
}

func maybeTags(r *rand.Rand) []string {
	n := r.Intn(3) // 0..2 tags
	if n == 0 {
		return nil
	}
	seen := map[string]bool{}
	out := make([]string, 0, n)
	for len(out) < n {
		t := randomChoice(r, tagPool)
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out
}

func randomFeedback(r *rand.Rand) Feedback {
	now := time.Now().UTC().UnixNano()
	errProb := r.Intn(100) < 7 // ~7% deliberately include "error" to hit errorSimulator

	var content string
	isNeg := false
	switch r.Intn(3) { // pos/neg/neutral
	case 0:
		content = randomChoice(r, posPhrases)
	case 1:
		content = randomChoice(r, negPhrases)
		isNeg = true
	default:
		content = randomChoice(r, neutralPhrases)
	}

	if errProb {
		// inject the word "error" somewhere to exercise the error path
		content = content + " with error spike"
	}

	return Feedback{
		CustomerID: fmt.Sprintf("C-%d-%04d", now, r.Intn(10000)),
		Content:    content,
		Category:   randomChoice(r, categories),
		IsNegative: isNeg,
		Tags:       maybeTags(r),
	}
}

/* ---------------- Main (gRPC only, runs forever) ---------------- */
func main() {
	rand.Seed(time.Now().UnixNano())

	// ---- config bits ----
	rxTargets := []string{
		envOr("RX_1", "localhost:50051"),
		envOr("RX_2", "localhost:50052"),
	}
	aesHex := envOr("AES256_KEY_HEX", "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec")

	// ---- lifecycle ----
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigs; cancel() }()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// ---- local wire to forward relay ----
	ground := builder.NewWire[Feedback](ctx)
	wire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Feedback](logger),
		builder.WireWithTransformer(errorSimulator),
	)

	tlsCfg := builder.NewTlsClientConfig(
		true,
		envOr("TLS_CLIENT_CRT", "../tls/client.crt"),
		envOr("TLS_CLIENT_KEY", "../tls/client.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	perf := builder.NewPerformanceOptions(true, builder.COMPRESS_SNAPPY)
	sec := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)
	key := mustHexKey(aesHex)

	// OAuth2 hints (receivers validate JWT)
	authBaseURL := envOr("OAUTH_BASE_URL", "https://localhost:3000")
	issuerBase := envOr("OAUTH_ISSUER_BASE", "auth-service")
	jwksURL := envOr("OAUTH_JWKS_URL", strings.TrimRight(authBaseURL, "/")+"/api/auth/oauth/jwks.json")
	aud := []string{"your-api"}
	scp := []string{"write:data"}

	oauthHints := builder.NewForwardRelayOAuth2JWTOptions(issuerBase, jwksURL, aud, scp, 300)
	authOpts := builder.NewForwardRelayAuthenticationOptionsOAuth2(oauthHints)
	authHTTP := httpClientTLSInsecure()
	ts := builder.NewForwardRelayRefreshingClientCredentialsSource(
		authBaseURL,
		envOr("OAUTH_CLIENT_ID", "steeze-local-cli"),
		envOr("OAUTH_CLIENT_SECRET", "local-secret"), // set real secret in env
		scp,
		20*time.Second,
		authHTTP,
	)
	if firstTok, err := ts.AccessToken(ctx); err == nil {
		dumpJWT(firstTok)
	} else {
		log.Printf("warning: token prefetch failed: %v", err)
	}

	staticHeaders := map[string]string{"x-tenant": "local"}
	dynamicHeaders := func(ctx context.Context) map[string]string {
		return map[string]string{"x-trace-id": fmt.Sprintf("t-%d", time.Now().UnixNano())}
	}

	relay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithLogger[Feedback](logger),
		builder.ForwardRelayWithTarget[Feedback](rxTargets...),
		builder.ForwardRelayWithPerformanceOptions[Feedback](perf),
		builder.ForwardRelayWithSecurityOptions[Feedback](sec, key),
		builder.ForwardRelayWithInput(ground),
		builder.ForwardRelayWithTLSConfig[Feedback](tlsCfg),

		builder.ForwardRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ForwardRelayWithOAuthBearer[Feedback](ts),
		builder.ForwardRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.ForwardRelayWithDynamicHeaders[Feedback](dynamicHeaders),
		builder.ForwardRelayWithAuthRequired[Feedback](true),
	)

	if err := wire.Start(ctx); err != nil {
		log.Fatalf("wire start: %v", err)
	}
	if err := relay.Start(ctx); err != nil {
		log.Fatalf("relay start: %v", err)
	}

	log.Printf("=== streaming random feedback via gRPC forever ===")

	// base interval + small jitter to avoid burstiness
	base := 500 * time.Millisecond
	jitter := 400 * time.Millisecond

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		select {
		case <-ctx.Done():
			wire.Stop()
			relay.Stop()
			log.Println("Shutdown complete.")
			return
		default:
			fb := randomFeedback(r)
			if err := ground.Submit(ctx, fb); err != nil {
				log.Printf("[gRPC] submit error: %v", err)
			} else {
				log.Printf("[gRPC] sent: %s", pretty(fb))
			}
			sleep := base + time.Duration(r.Int63n(int64(jitter)))
			time.Sleep(sleep)
		}
	}
}
