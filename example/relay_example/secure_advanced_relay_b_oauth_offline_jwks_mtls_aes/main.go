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

const AES256KeyHex = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"

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
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigs; fmt.Println("Shutting down..."); cancel() }()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// Wires & drains (no transforms)
	wireA := builder.NewWire[Feedback](ctx, builder.WireWithLogger[Feedback](logger))
	wireB := builder.NewWire[Feedback](ctx, builder.WireWithLogger[Feedback](logger))
	conduitA := builder.NewConduit(ctx, builder.ConduitWithWire(wireA))
	conduitB := builder.NewConduit(ctx, builder.ConduitWithWire(wireB))

	// TLS (server)
	tlsCfg := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS13, tls.VersionTLS13,
	)

	// Decryption must match forwarder
	decKey := mustAES()

	// =======================
	// AUTH: JWKS ONLY
	// =======================
	issuerBase := envOr("OAUTH_ISSUER_BASE", "https://localhost:3000")
	jwksURL := envOr("OAUTH_JWKS_URL", "https://localhost:3000/api/auth/.well-known/jwks.json")

	jwtOpts := builder.NewReceivingRelayOAuth2JWTOptions(
		issuerBase,
		jwksURL,
		[]string{"your-api"},   // required audience(s)
		[]string{"write:data"}, // required scope(s)
		300,                    // cache/leeway seconds
	)

	// no introspection configured here — jwks only
	oauth := builder.NewReceivingRelayMergeOAuth2Options(jwtOpts, nil)
	authOpts := builder.NewReceivingRelayAuthenticationOptionsOAuth2(oauth)

	// Static headers you expect from forwarder
	staticHeaders := map[string]string{"x-tenant": "local"}

	recvA := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback](envOr("RX_A", "localhost:50053")),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(wireA),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsCfg),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decKey),

		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ReceivingRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.ReceivingRelayWithAuthRequired[Feedback](true),
	)

	recvB := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback](envOr("RX_B", "localhost:50054")),
		builder.ReceivingRelayWithBufferSize[Feedback](1000),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput(wireB),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsCfg),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decKey),

		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ReceivingRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.ReceivingRelayWithAuthRequired[Feedback](true),
	)

	if err := recvA.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := recvB.Start(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Receivers up (JWKS ONLY). Ctrl+C to stop.")
	<-ctx.Done()

	fmt.Println("---- Receiver A ----")
	if b, err := conduitA.LoadAsJSONArray(); err == nil {
		fmt.Println(string(b))
	} else {
		fmt.Println("A err:", err)
	}
	fmt.Println("---- Receiver B ----")
	if b, err := conduitB.LoadAsJSONArray(); err == nil {
		fmt.Println(string(b))
	} else {
		fmt.Println("B err:", err)
	}

	recvA.Stop()
	recvB.Stop()
	wireA.Stop()
	wireB.Stop()
	fmt.Println("Done.")
}
