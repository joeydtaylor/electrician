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
	"net/url"
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

// ----- Demo payload -----

type Feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

func errorSimulator(feedback Feedback) (Feedback, error) {
	if strings.Contains(strings.ToLower(feedback.Content), "error") {
		return Feedback{}, errors.New("simulated processing error")
	}
	return feedback, nil
}

func plugFunc(ctx context.Context, submit func(ctx context.Context, fb Feedback) error) {
	t := time.NewTicker(1000 * time.Millisecond)
	defer t.Stop()
	var n int
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			errTrig := ""
			if rand.Intn(10) == 0 {
				errTrig = " error "
			}
			_ = submit(ctx, Feedback{
				CustomerID: fmt.Sprintf("Feedback%d", n),
				Content:    fmt.Sprintf("This is feedback item number %d%s", n, errTrig),
				IsNegative: n%10 == 0,
			})
			n++
		}
	}
}

// ----- Local demo-only content encryption key -----

const AES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

func mustHexKey() string {
	raw, err := hex.DecodeString(AES256KeyHex)
	if err != nil {
		log.Fatalf("bad hex key: %v", err)
	}
	return string(raw) // 32 bytes -> AES-256
}

// ----- OAuth2: client_credentials against aegis-auth -----

type tokenResp struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
	Scope       string `json:"scope"`
}

func fetchClientCredentialsToken(ctx context.Context, baseURL, clientID, clientSecret string, scopes []string) (string, error) {
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	if len(scopes) > 0 {
		form.Set("scope", strings.Join(scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/api/auth/oauth/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(clientID, clientSecret)

	// Dev: accept self-signed cert from local aegis-auth
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS13,
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true, // dev only
		},
	}
	hc := &http.Client{Transport: tr, Timeout: 10 * time.Second}

	resp, err := hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("token endpoint status %d", resp.StatusCode)
	}

	var trsp tokenResp
	if err := json.NewDecoder(resp.Body).Decode(&trsp); err != nil {
		return "", err
	}
	if trsp.AccessToken == "" {
		return "", errors.New("empty access_token")
	}
	return trsp.AccessToken, nil
}

// ----- JWT debug helpers -----

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

	if err := json.Unmarshal(dec(parts[0]), &hdr); err != nil {
		log.Printf("JWT header decode error: %v", err)
	}
	if err := json.Unmarshal(dec(parts[1]), &claims); err != nil {
		log.Printf("JWT claims decode error: %v", err)
	}

	log.Printf("JWT HEADER:\n%s", pretty(hdr))
	log.Printf("JWT CLAIMS:\n%s", pretty(claims))

	// Surface common scope fields
	if s, ok := claims["scope"].(string); ok && s != "" {
		log.Printf("scope (string): %q", s)
	}
	if arr, ok := claims["scp"].([]any); ok && len(arr) > 0 {
		var out []string
		for _, v := range arr {
			if x, ok := v.(string); ok {
				out = append(out, x)
			}
		}
		log.Printf("scp (array): %v", out)
	}
	if arr, ok := claims["permissions"].([]any); ok && len(arr) > 0 {
		var out []string
		for _, v := range arr {
			if x, ok := v.(string); ok {
				out = append(out, x)
			}
		}
		log.Printf("permissions (array): %v", out)
	}
}

// ----- Main -----

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	plug := builder.NewPlug(ctx, builder.PlugWithAdapterFunc(plugFunc))
	gen := builder.NewGenerator(ctx, builder.GeneratorWithPlug(plug))

	ground := builder.NewWire[Feedback](ctx)
	cb := builder.NewCircuitBreaker(ctx, 1, 10*time.Second,
		builder.CircuitBreakerWithNeutralWire(ground),
		builder.CircuitBreakerWithLogger[Feedback](logger),
	)
	wire := builder.NewWire(
		ctx,
		builder.WireWithLogger[Feedback](logger),
		builder.WireWithConcurrencyControl[Feedback](1000, 100),
		builder.WireWithCircuitBreaker(cb),
		builder.WireWithTransformer(errorSimulator),
		builder.WireWithGenerator(gen),
	)

	// TLS 1.3 client config for gRPC to receiving relays
	tlsCfg := builder.NewTlsClientConfig(
		true,
		"../tls/client.crt",
		"../tls/client.key",
		"../tls/ca.crt",
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	// Perf + content encryption (AES-GCM over the payload)
	perf := builder.NewPerformanceOptions(true, builder.COMPRESS_SNAPPY)
	sec := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)
	key := mustHexKey()

	// OAuth2 resource-server hints (receivers validate JWT via JWKS)
	const issuerBase = "https://localhost:3000"
	const jwksURL = "https://localhost:3000/api/auth/.well-known/jwks.json"
	aud := []string{"your-api"}
	scp := []string{"write:data"}

	oauthHints := builder.NewForwardRelayOAuth2JWTOptions(issuerBase, jwksURL, aud, scp, 300)
	authOpts := builder.NewForwardRelayAuthenticationOptionsOAuth2(oauthHints)

	// Acquire token dynamically via client_credentials for seeded confidential client:
	// client_id: steeze-local-cli ; client_secret: local-secret
	token, err := fetchClientCredentialsToken(ctx, issuerBase, "steeze-local-cli", "local-secret", scp)
	if err != nil {
		log.Fatalf("token fetch failed: %v", err)
	}
	dumpJWT(token) // print header/claims/scopes
	ts := builder.NewForwardRelayStaticBearerTokenSource(token)

	staticHeaders := map[string]string{"x-tenant": "local"}

	relay := builder.NewForwardRelay(
		ctx,
		builder.ForwardRelayWithLogger[Feedback](logger),
		builder.ForwardRelayWithTarget[Feedback]("localhost:50051", "localhost:50052"),
		builder.ForwardRelayWithPerformanceOptions[Feedback](perf),
		builder.ForwardRelayWithSecurityOptions[Feedback](sec, key),
		builder.ForwardRelayWithInput(ground),
		builder.ForwardRelayWithTLSConfig[Feedback](tlsCfg),

		// OAuth2
		builder.ForwardRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ForwardRelayWithOAuthBearer[Feedback](ts),
		builder.ForwardRelayWithStaticHeaders[Feedback](staticHeaders),
	)

	wire.Start(ctx)
	relay.Start(ctx)

	<-ctx.Done()

	wire.Stop()
	relay.Stop()

	if ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Processing timeout.")
	} else {
		fmt.Println("Processing finished within the allotted time.")
	}
}
