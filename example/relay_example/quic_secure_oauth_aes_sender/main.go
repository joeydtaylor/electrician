package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
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

const (
	relayAddr          = "localhost:50072"
	sendCount          = 100
	sendInterval       = 50 * time.Millisecond
	logLevel           = "info"
	aes256KeyHex       = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
	authBaseURL        = "https://localhost:3000"
	oauthIssuer        = "auth-service"
	oauthJWKSURL       = "https://localhost:3000/api/auth/oauth/jwks.json"
	oauthClientID      = "steeze-local-cli"
	oauthClientSecret  = "local-secret"
	tlsClientCert      = "../tls/client.crt"
	tlsClientKey       = "../tls/client.key"
	tlsCA              = "../tls/ca.crt"
	oauthTokenLeeway   = 20 * time.Second
	oauthJWKSCacheSecs = 300
	staticTenantHeader = "local"
)

var (
	oauthAudiences = []string{"your-api"}
	oauthScopes    = []string{"write:data"}
)

func mustAES() string {
	raw, err := hex.DecodeString(aes256KeyHex)
	if err != nil || len(raw) != 32 {
		log.Fatalf("bad AES key: %v", err)
	}
	return string(raw)
}

func httpClientTLSInsecure() *http.Client {
	return &http.Client{
		Timeout: 6 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				MaxVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true,
			},
		},
	}
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(sendCount)*sendInterval+2*time.Second)
	defer cancel()

	logger := builder.NewLogger(
		builder.LoggerWithDevelopment(true),
		builder.LoggerWithLevel(logLevel),
	)
	defer func() {
		_ = logger.Flush()
	}()

	tlsCfg := builder.NewTlsClientConfig(
		true,
		tlsClientCert,
		tlsClientKey,
		tlsCA,
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	sec := builder.NewSecurityOptions(true, builder.ENCRYPTION_AES_GCM)
	staticHeaders := map[string]string{"x-tenant": staticTenantHeader}
	dynamicHeaders := func(ctx context.Context) map[string]string {
		return map[string]string{"x-trace-id": fmt.Sprintf("t-%d", time.Now().UnixNano())}
	}

	oauthHints := builder.NewForwardRelayOAuth2JWTOptions(
		oauthIssuer,
		oauthJWKSURL,
		oauthAudiences,
		oauthScopes,
		oauthJWKSCacheSecs,
	)
	authOpts := builder.NewForwardRelayAuthenticationOptionsOAuth2(oauthHints)
	tokenSource := builder.NewQuicForwardRelayRefreshingClientCredentialsSource(
		authBaseURL,
		oauthClientID,
		oauthClientSecret,
		oauthScopes,
		oauthTokenLeeway,
		httpClientTLSInsecure(),
	)
	if tok, err := tokenSource.AccessToken(ctx); err != nil {
		logger.Warn("Auth token prefetch failed",
			"event", "AuthToken",
			"result", "FAILURE",
			"error", err,
		)
	} else {
		logger.Debug("Auth token prefetched",
			"event", "AuthToken",
			"result", "SUCCESS",
			"token", tok,
		)
	}

	logger.Info("QUIC sender starting",
		"event", "Start",
		"result", "SUCCESS",
		"target", relayAddr,
		"count", sendCount,
		"interval", sendInterval.String(),
		"auth_required", true,
		"static_headers", staticHeaders,
		"issuer", oauthIssuer,
		"auth_base_url", authBaseURL,
		"jwks_url", oauthJWKSURL,
	)

	fr := builder.NewQuicForwardRelay(
		ctx,
		builder.QuicForwardRelayWithTarget[Feedback](relayAddr),
		builder.QuicForwardRelayWithLogger[Feedback](logger),
		builder.QuicForwardRelayWithTLSConfig[Feedback](tlsCfg),
		builder.QuicForwardRelayWithSecurityOptions[Feedback](sec, mustAES()),
		builder.QuicForwardRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.QuicForwardRelayWithOAuthBearer[Feedback](tokenSource),
		builder.QuicForwardRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.QuicForwardRelayWithDynamicHeaders[Feedback](dynamicHeaders),
		builder.QuicForwardRelayWithAuthRequired[Feedback](true),
	)

	start := time.Now()
	for i := 0; i < sendCount; i++ {
		fb := Feedback{
			CustomerID: fmt.Sprintf("cust-%03d", i%100),
			Content:    "Secure QUIC payload",
			Category:   "feedback",
			IsNegative: i%10 == 0,
			Tags:       []string{"quic", "secure"},
		}
		if err := fr.Submit(ctx, fb); err != nil {
			logger.Error("Submit failed",
				"event", "Submit",
				"result", "FAILURE",
				"error", err,
				"seq", i,
			)
		}
		time.Sleep(sendInterval)
	}

	logger.Info("Send complete",
		"event", "SendComplete",
		"result", "SUCCESS",
		"count", sendCount,
		"elapsed", time.Since(start).String(),
	)
}
