package main

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
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
	defer signal.Stop(sigs)
	go func() {
		<-sigs
		fmt.Println("Shutting down...")
		cancel()
	}()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))
	workers := runtime.GOMAXPROCS(0)

	wire := builder.NewWire[Feedback](
		ctx,
		builder.WireWithLogger[Feedback](logger),
		builder.WireWithConcurrencyControl[Feedback](16384, workers),
	)
	_ = wire.Start(ctx)

	tlsCfg := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS13, tls.VersionTLS13,
	)

	decKey := mustAES()

	issuerBase := envOr("OAUTH_ISSUER_BASE", "https://localhost:3000")
	jwksURL := envOr("OAUTH_JWKS_URL", "https://localhost:3000/api/auth/oauth/jwks.json")

	jwtOpts := builder.NewReceivingRelayOAuth2JWTOptions(
		issuerBase,
		jwksURL,
		[]string{"your-api"},
		[]string{"write:data"},
		300,
	)

	oauth := builder.NewReceivingRelayMergeOAuth2Options(jwtOpts, nil)
	authOpts := builder.NewReceivingRelayAuthenticationOptionsOAuth2(oauth)

	staticHeaders := map[string]string{"x-tenant": "local"}
	// Allow any browser origin for CORS; auth and TLS are still enforced.
	grpcWebCfg := builder.NewGRPCWebConfigAllowAllOrigins()

	recv := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback](envOr("RX_ADDR", "localhost:50051")),
		builder.ReceivingRelayWithBufferSize[Feedback](16384),
		builder.ReceivingRelayWithLogger[Feedback](logger),
		builder.ReceivingRelayWithOutput[Feedback](wire),
		builder.ReceivingRelayWithTLSConfig[Feedback](tlsCfg),
		builder.ReceivingRelayWithDecryptionKey[Feedback](decKey),
		builder.ReceivingRelayWithAuthenticationOptions[Feedback](authOpts),
		builder.ReceivingRelayWithStaticHeaders[Feedback](staticHeaders),
		builder.ReceivingRelayWithAuthRequired[Feedback](true),
		builder.ReceivingRelayWithGRPCWebConfig[Feedback](grpcWebCfg),
	)

	if err := recv.StartGRPCWeb(ctx); err != nil {
		log.Fatal(err)
	}

	fmt.Println("Receiver up (gRPC-Web + JWKS). Ctrl+C to stop.")
	<-ctx.Done()

	recv.Stop()

	fmt.Println("---- Receiver ----")
	if b, err := wire.LoadAsJSONArray(); err == nil {
		fmt.Println(string(b))
	} else {
		fmt.Println("err:", err)
	}

	fmt.Println("Done.")
}
