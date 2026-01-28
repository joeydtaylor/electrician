package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"os"
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
	AES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
)

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func mustAES() string {
	raw, err := hex.DecodeString(AES256KeyHex)
	if err != nil || len(raw) != 32 {
		log.Fatalf("bad AES key: %v", err)
	}
	return string(raw)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	wire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithLogger[Feedback](logger),
	)
	if err := wire.Start(ctx); err != nil {
		log.Fatalf("wire start: %v", err)
	}

	tlsCfg := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS13, tls.VersionTLS13,
	)

	decKey := mustAES()

	introspectionURL := envOr("OAUTH_INTROSPECTION_URL", "https://localhost:3000/api/auth/oauth/introspect")
	oauth := builder.NewQuicReceivingRelayOAuth2IntrospectionOptions(
		introspectionURL,
		envOr("OAUTH_INTROSPECTION_AUTH_TYPE", "none"),
		envOr("OAUTH_INTROSPECTION_CLIENT_ID", ""),
		envOr("OAUTH_INTROSPECTION_CLIENT_SECRET", ""),
		envOr("OAUTH_INTROSPECTION_BEARER_TOKEN", ""),
		300,
	)

	authOpts := builder.NewQuicReceivingRelayAuthenticationOptionsOAuth2(oauth)

	recv := builder.NewQuicReceivingRelay[Feedback](
		ctx,
		builder.QuicReceivingRelayWithAddress[Feedback](envOr("RX_ADDR", "localhost:50072")),
		builder.QuicReceivingRelayWithLogger[Feedback](logger),
		builder.QuicReceivingRelayWithOutput[Feedback](wire),
		builder.QuicReceivingRelayWithTLSConfig[Feedback](tlsCfg),
		builder.QuicReceivingRelayWithDecryptionKey[Feedback](decKey),
		builder.QuicReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.QuicReceivingRelayWithStaticHeaders[Feedback](map[string]string{"x-tenant": "local"}),
		builder.QuicReceivingRelayWithAuthRequired[Feedback](true),
	)

	if err := recv.Start(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Secure QUIC receiver up. Ctrl+C to stop.")
	<-ctx.Done()
	fmt.Println("Shutting down...")

	recv.Stop()

	if b, err := wire.LoadAsJSONArray(); err == nil {
		fmt.Println("---- Receiver ----")
		fmt.Println(string(b))
	}
}
