package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `parquet:"name=customerId, type=BYTE_ARRAY, convertedtype=UTF8" json:"customerId"`
	Content    string   `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8" json:"content"`
	Category   string   `parquet:"name=category, type=BYTE_ARRAY, convertedtype=UTF8" json:"category,omitempty"`
	IsNegative bool     `parquet:"name=isNegative, type=BOOLEAN" json:"isNegative"`
	Tags       []string `parquet:"name=tags, type=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags,omitempty"`
}

const (
	AES256KeyHex     = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
	defaultOrgID     = "4d948fa0-084e-490b-aad5-cfd01eeab79a"
	assertJWTEnvName = "ASSERT_JWT" // for testing: export ASSERT_JWT="<your assert cookie JWT>"
	bearerJWTEnvName = "BEARER_JWT" // for testing: export BEARER_JWT="<your bearer access token>"
	orgIDEnvName     = "ORG_ID"     // for forcing org id explicitly
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

func base64URLDecode(s string) ([]byte, error) {
	// add missing padding if any
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}

func orgFromJWT(tok string) (string, bool) {
	parts := strings.Split(tok, ".")
	if len(parts) != 3 {
		return "", false
	}
	payloadB, err := base64URLDecode(parts[1])
	if err != nil {
		return "", false
	}
	var claims map[string]any
	if err := json.Unmarshal(payloadB, &claims); err != nil {
		return "", false
	}
	// try common keys
	for _, k := range []string{"org", "org_id", "orgId", "organization"} {
		if v, ok := claims[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s, true
			}
		}
	}
	return "", false
}

func resolveOrgID() string {
	// 1) explicit env override
	if s := os.Getenv(orgIDEnvName); s != "" {
		return s
	}
	// 2) ASSERT_JWT (cookie value without the "assert=" prefix)
	if s := os.Getenv(assertJWTEnvName); s != "" {
		if org, ok := orgFromJWT(s); ok {
			return org
		}
	}
	// 3) BEARER_JWT
	if s := os.Getenv(bearerJWTEnvName); s != "" {
		if org, ok := orgFromJWT(s); ok {
			return org
		}
	}
	// 4) fallback
	return defaultOrgID
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Ctrl+C
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() { <-sigs; fmt.Println("Shutting down..."); cancel() }()

	logger := builder.NewLogger(builder.LoggerWithDevelopment(true))

	// Wires (no transforms)
	wireA := builder.NewWire[Feedback](ctx, builder.WireWithLogger[Feedback](logger))
	wireB := builder.NewWire[Feedback](ctx, builder.WireWithLogger[Feedback](logger))
	_ = wireA.Start(ctx)
	_ = wireB.Start(ctx)

	// TLS (server)
	tlsCfg := builder.NewTlsServerConfig(
		true,
		envOr("TLS_CERT", "../tls/server.crt"),
		envOr("TLS_KEY", "../tls/server.key"),
		envOr("TLS_CA", "../tls/ca.crt"),
		"localhost",
		tls.VersionTLS13, tls.VersionTLS13,
	)

	// OAuth (JWKS only)
	issuerBase := envOr("OAUTH_ISSUER_BASE", "https://localhost:3000")
	jwksURL := envOr("OAUTH_JWKS_URL", "https://localhost:3000/api/auth/.well-known/jwks.json")
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
	decKey := mustAES()

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

	// Start receivers
	if err := recvA.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := recvB.Start(ctx); err != nil {
		log.Fatal(err)
	}

	// S3 (LocalStack)
	cli, err := builder.NewS3ClientAssumeRole(
		ctx,
		"us-east-1",
		"arn:aws:iam::000000000000:role/exodus-dev-role",
		"electrician-writer",
		15*time.Minute,
		"",
		aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "")),
		"http://localhost:4566",
		true,
	)
	if err != nil {
		log.Fatal(err)
	}

	const bucket = "steeze-dev"
	const kmsAliasArn = "arn:aws:kms:us-east-1:000000000000:alias/electrician-dev"

	// derive org id from env/JWT
	orgID := resolveOrgID()
	logger.Info(fmt.Sprintf("using org id: %s", orgID))

	// Prefix nested under org=<orgID>/...
	prefix := fmt.Sprintf("org=%s/feedback/demo/{yyyy}/{MM}/{dd}/{HH}/{mm}/", orgID)

	// Idempotent create for LocalStack
	_, _ = cli.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})

	// S3 adapter: Parquet + SSE-KMS
	s3ad := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),

		// Parquet writer + key layout
		builder.S3ClientAdapterWithFormat[Feedback]("parquet", ""),
		builder.S3ClientAdapterWithWriterPrefixTemplate[Feedback](prefix),

		// rolling thresholds (also used by parquet roller)
		builder.S3ClientAdapterWithBatchSettings[Feedback](5000, 128<<20, 1*time.Second),

		// codec knob
		builder.S3ClientAdapterWithWriterFormatOptions[Feedback](map[string]string{
			"parquet_compression": "snappy", // snappy|zstd|gzip
		}),

		// SSE-KMS
		builder.S3ClientAdapterWithSSE[Feedback]("aws:kms", kmsAliasArn),

		// Fan-in both wires
		builder.S3ClientAdapterWithWire[Feedback](wireA, wireB),

		builder.S3ClientAdapterWithLogger[Feedback](logger),
	)

	// Start parquet stream writer (runs until ctx is canceled)
	type hasStartWriter[T any] interface {
		StartWriter(context.Context) error
		Stop()
	}
	if ws, ok := s3ad.(hasStartWriter[Feedback]); ok {
		if err := ws.StartWriter(ctx); err != nil {
			logger.Error(fmt.Sprintf("StartWriter failed: %v", err))
			cancel()
		}
	} else {
		logger.Error("S3 adapter doesn't expose StartWriter; ensure your adapter version includes ConnectInput + StartWriter")
		cancel()
	}

	fmt.Println("Receivers up; streaming to S3 as rolling Parquet (SSE-KMS). Ctrl+C to stop.")
	<-ctx.Done()

	// Graceful shutdown
	recvA.Stop()
	recvB.Stop()
	wireA.Stop()
	wireB.Stop()
	if ws, ok := s3ad.(hasStartWriter[Feedback]); ok {
		ws.Stop()
	}
	fmt.Println("Done.")
}
