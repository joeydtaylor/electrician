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
	"strconv"
	"strings"
	"sync/atomic"
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
	assertJWTEnvName = "ASSERT_JWT"
	bearerJWTEnvName = "BEARER_JWT"
	orgIDEnvName     = "ORG_ID"
)

func trueish(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "1" || s == "true" || s == "yes" || s == "y"
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}

func envDurMs(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			return time.Duration(n) * time.Millisecond
		}
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
	if s := os.Getenv(orgIDEnvName); s != "" {
		return s
	}
	if s := os.Getenv(assertJWTEnvName); s != "" {
		if org, ok := orgFromJWT(s); ok {
			return org
		}
	}
	if s := os.Getenv(bearerJWTEnvName); s != "" {
		if org, ok := orgFromJWT(s); ok {
			return org
		}
	}
	return defaultOrgID
}

type s3Stats struct {
	writerStarts    atomic.Int64
	writerStops     atomic.Int64
	keysRendered    atomic.Int64
	putAttempts     atomic.Int64
	putSuccesses    atomic.Int64
	putErrors       atomic.Int64
	bytesAttempted  atomic.Int64
	bytesWritten    atomic.Int64
	parquetRolls    atomic.Int64
	recordsFlushed  atomic.Int64
	lastCompression atomic.Value // string
	lastError       atomic.Value // string
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
	issuerBase := envOr("OAUTH_ISSUER_BASE", "auth-service")
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
	decKey := mustAES()

	recvA := builder.NewReceivingRelay[Feedback](
		ctx,
		builder.ReceivingRelayWithAddress[Feedback](envOr("RX_A", "localhost:50053")),
		builder.ReceivingRelayWithBufferSize[Feedback](1024),
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
		builder.ReceivingRelayWithBufferSize[Feedback](1024),
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
	orgID := resolveOrgID()
	logger.Info(fmt.Sprintf("using org id: %s", orgID))

	// Partitioning: default HOURLY (prod-ish). Set MINUTE_PARTITIONS=true for {mm}.
	useMinute := trueish(os.Getenv("MINUTE_PARTITIONS"))
	var prefix string
	if useMinute {
		prefix = fmt.Sprintf("org=%s/feedback/demo/{yyyy}/{MM}/{dd}/{HH}/{mm}/", orgID)
	} else {
		prefix = fmt.Sprintf("org=%s/feedback/demo/{yyyy}/{MM}/{dd}/{HH}/", orgID)
	}

	// Simulation toggles (default off in prod)
	simPutErr := trueish(os.Getenv("SIMULATE_S3_PUT_ERROR"))
	simBadKMS := trueish(os.Getenv("SIMULATE_S3_BAD_KMS"))

	// Effective bucket (non-existent when simulating PUT errors)
	effectiveBucket := bucket
	if simPutErr {
		effectiveBucket = bucket + "-missing"
		logger.Warn(fmt.Sprintf("[SIM] forcing PutObject error by writing to non-existent bucket: %s", effectiveBucket))
	}

	// Create only the real bucket (not the simulated-missing one)
	_, _ = cli.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})

	// ----------------------------
	// Sensor that counts S3 events
	// ----------------------------
	var stats s3Stats
	stats.lastCompression.Store("unknown")
	stats.lastError.Store("")

	s := builder.NewSensor[Feedback](
		builder.SensorWithLogger[Feedback](logger),

		builder.SensorWithOnS3WriterStartFunc[Feedback](func(_ builder.ComponentMetadata, _bkt, _pref, _fmt string) {
			stats.writerStarts.Add(1)
		}),
		builder.SensorWithOnS3WriterStopFunc[Feedback](func(_ builder.ComponentMetadata) {
			stats.writerStops.Add(1)
		}),

		builder.SensorWithOnS3KeyRenderedFunc[Feedback](func(_ builder.ComponentMetadata, _key string) {
			stats.keysRendered.Add(1)
		}),
		builder.SensorWithOnS3PutAttemptFunc[Feedback](func(_ builder.ComponentMetadata, _bkt, _key string, n int, _sse, _kms string) {
			stats.putAttempts.Add(1)
			stats.bytesAttempted.Add(int64(n))
		}),
		builder.SensorWithOnS3PutSuccessFunc[Feedback](func(_ builder.ComponentMetadata, _bkt, _key string, n int, _ time.Duration) {
			stats.putSuccesses.Add(1)
			stats.bytesWritten.Add(int64(n))
		}),
		builder.SensorWithOnS3PutErrorFunc[Feedback](func(_ builder.ComponentMetadata, _bkt, _key string, _ int, err error) {
			stats.putErrors.Add(1)
			if err != nil {
				stats.lastError.Store(err.Error())
			}
		}),
		builder.SensorWithOnS3ParquetRollFlushFunc[Feedback](func(_ builder.ComponentMetadata, recs int, _bytes int, comp string) {
			stats.parquetRolls.Add(1)
			stats.recordsFlushed.Add(int64(recs))
			stats.lastCompression.Store(comp)
		}),
	)

	// ---------- Prod-ish knobs (with env overrides) ----------
	// Targets: bigger Parquet files, fewer objects, better scan cost.
	parquetCompression := strings.ToLower(envOr("PARQUET_COMPRESSION", "zstd")) // zstd|snappy|gzip
	rollWindow := envInt("ROLL_WINDOW_MS", 300_000)                             // default 5 min
	rollMaxRecords := envInt("ROLL_MAX_RECORDS", 250_000)                       // tune per row width & memory
	batchMaxRecords := envInt("BATCH_MAX_RECORDS", 500_000)
	batchMaxBytes := envInt("BATCH_MAX_BYTES_MB", 256) * (1 << 20) // MiB -> bytes
	batchMaxAge := envDurMs("BATCH_MAX_AGE_MS", 5*time.Minute)

	// KMS alias (optionally bogus to simulate KMS failure)
	kmsAliasArn := "arn:aws:kms:us-east-1:000000000000:alias/electrician-dev"
	if simBadKMS {
		kmsAliasArn = "arn:aws:kms:us-east-1:000000000000:alias/does-not-exist"
		logger.Warn("[SIM] using bogus KMS alias to attempt KMS-related failure")
	}

	// S3 adapter: Parquet + SSE-KMS + Sensor
	s3ad := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, effectiveBucket),
		builder.S3ClientAdapterWithSensor[Feedback](s),

		builder.S3ClientAdapterWithFormat[Feedback]("parquet", ""),
		builder.S3ClientAdapterWithWriterPrefixTemplate[Feedback](prefix),

		// Batch thresholds (also used by parquet roller defaults; explicit roll_* below win)
		builder.S3ClientAdapterWithBatchSettings[Feedback](batchMaxRecords, batchMaxBytes, batchMaxAge),

		// Parquet writer knobs
		builder.S3ClientAdapterWithWriterFormatOptions[Feedback](map[string]string{
			"parquet_compression": parquetCompression,           // zstd (default), snappy, gzip
			"roll_window_ms":      strconv.Itoa(rollWindow),     // e.g. 300000
			"roll_max_records":    strconv.Itoa(rollMaxRecords), // e.g. 250000
			// tip: for very high volume, consider adding sort/bloom options when supported
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

	// ---------
	// Printout
	// ---------
	summary := map[string]any{
		"orgId":            orgID,
		"bucket":           effectiveBucket,
		"prefixTemplate":   prefix,
		"partitioning":     map[string]any{"minute": useMinute},
		"compression":      parquetCompression,
		"rollWindowMs":     rollWindow,
		"rollMaxRecords":   rollMaxRecords,
		"batchMaxRecords":  batchMaxRecords,
		"batchMaxBytesMiB": batchMaxBytes / (1 << 20),
		"batchMaxAgeMs":    int(batchMaxAge / time.Millisecond),
		"simulated":        map[string]any{"putError": simPutErr, "badKMS": simBadKMS},

		"writerStarts":    stats.writerStarts.Load(),
		"writerStops":     stats.writerStops.Load(),
		"keysRendered":    stats.keysRendered.Load(),
		"putAttempts":     stats.putAttempts.Load(),
		"putSuccesses":    stats.putSuccesses.Load(),
		"putErrors":       stats.putErrors.Load(),
		"bytesAttempted":  stats.bytesAttempted.Load(),
		"bytesWritten":    stats.bytesWritten.Load(),
		"parquetRolls":    stats.parquetRolls.Load(),
		"recordsFlushed":  stats.recordsFlushed.Load(),
		"lastCompression": stats.lastCompression.Load(),
		"lastError":       stats.lastError.Load(),
	}
	js, _ := json.MarshalIndent(summary, "", "  ")
	fmt.Println(string(js))
	fmt.Println("Done.")
}
