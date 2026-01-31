package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"os/signal"
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

const (
	relayAddr         = "localhost:50072"
	logLevel          = "info"
	aes256KeyHex      = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
	oauthIssuer       = "auth-service"
	oauthJWKSURL      = "https://localhost:3000/api/auth/oauth/jwks.json"
	oauthJWKSCacheSec = 300
	staticTenant      = "local"
	tlsCert           = "../tls/server.crt"
	tlsKey            = "../tls/server.key"
	tlsCA             = "../tls/ca.crt"
	tlsSAN            = "localhost"
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

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := builder.NewLogger(
		builder.LoggerWithDevelopment(true),
		builder.LoggerWithLevel(logLevel),
	)

	wire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithLogger[Feedback](logger),
	)
	if err := wire.Start(ctx); err != nil {
		log.Fatalf("wire start: %v", err)
	}

	tlsCfg := builder.NewTlsServerConfig(
		true,
		tlsCert,
		tlsKey,
		tlsCA,
		tlsSAN,
		tls.VersionTLS13,
		tls.VersionTLS13,
	)

	jwtOpts := builder.NewQuicReceivingRelayOAuth2JWTOptions(
		oauthIssuer,
		oauthJWKSURL,
		oauthAudiences,
		oauthScopes,
		oauthJWKSCacheSec,
	)
	authOpts := builder.NewQuicReceivingRelayAuthenticationOptionsOAuth2(jwtOpts)

	staticHeaders := map[string]string{"x-tenant": staticTenant}

	recv := builder.NewQuicReceivingRelay(
		ctx,
		builder.QuicReceivingRelayWithAddress[Feedback](relayAddr),
		builder.QuicReceivingRelayWithLogger[Feedback](logger),
		builder.QuicReceivingRelayWithOutput(wire),
		builder.QuicReceivingRelayWithTLSConfig[Feedback](tlsCfg),
		builder.QuicReceivingRelayWithDecryptionKey[Feedback](mustAES()),
		builder.QuicReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.QuicReceivingRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.QuicReceivingRelayWithAuthRequired[Feedback](true),
	)

	logger.Info("QUIC receiver starting",
		"event", "Start",
		"result", "SUCCESS",
		"address", relayAddr,
		"auth_required", true,
		"static_headers", staticHeaders,
		"issuer", oauthIssuer,
		"jwks_url", oauthJWKSURL,
		"required_audience", oauthAudiences,
		"required_scopes", oauthScopes,
	)

	if err := recv.Start(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Secure QUIC receiver up. Ctrl+C to stop.")
	<-ctx.Done()
	fmt.Println("Shutting down...")

	logger.Info("QUIC receiver shutting down",
		"event", "Stop",
		"result", "SUCCESS",
	)

	recv.Stop()

	if b, err := wire.LoadAsJSONArray(); err == nil {
		fmt.Println("---- Receiver ----")
		fmt.Println(string(b))
	}
}
