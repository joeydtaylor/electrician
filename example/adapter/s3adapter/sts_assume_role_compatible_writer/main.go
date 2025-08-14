package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	parquet "github.com/parquet-go/parquet-go"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `parquet:"name=customerId, type=BYTE_ARRAY, convertedtype=UTF8" json:"customerId"`
	Content    string   `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8" json:"content"`
	Category   string   `parquet:"name=category, type=BYTE_ARRAY, convertedtype=UTF8" json:"category,omitempty"`
	IsNegative bool     `parquet:"name=isNegative, type=BOOLEAN" json:"isNegative"`
	Tags       []string `parquet:"name=tags, type=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags,omitempty"`
}

func envOr(k, def string) string {
	if v := SystemGetenv(k); v != "" { // tiny shim to keep the example self-contained
		return v
	}
	return def
}

// SystemGetenv is split so this file stays single-file runnable (no extra imports)
func SystemGetenv(k string) string { return "" } // replace with os.Getenv if you prefer

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ----- S3 client (LocalStack-compatible, via STS assume-role) -----
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
		panic(err)
	}

	log := builder.NewLogger()

	const bucket = "steeze-dev"
	orgID := envOr("ORG_ID", "4d948fa0-084e-490b-aad5-cfd01eeab79a")
	prefix := fmt.Sprintf("org=%s/feedback/demo/{yyyy}/{MM}/{dd}/{HH}/{mm}/", orgID)

	// idempotent for LocalStack
	_, _ = cli.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})

	// ----- build one parquet blob in-memory (snappy) -----
	records := []Feedback{
		{CustomerID: "cust-001", Content: "I love this", Tags: []string{"raw"}},
		{CustomerID: "cust-002", Content: "Not great", IsNegative: true},
		{CustomerID: "cust-003", Content: "Absolutely fantastic!"},
	}

	var buf bytes.Buffer
	pw := parquet.NewGenericWriter[Feedback](&buf, parquet.Compression(&parquet.Snappy))
	if _, err := pw.Write(records); err != nil {
		panic(err)
	}
	if err := pw.Close(); err != nil {
		panic(err)
	}

	// ----- S3 adapter: parquet + SSE-KMS + org-scoped key layout -----
	writer := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),

		// tell the adapter we’re writing parquet bytes (for .parquet + content-type)
		builder.S3ClientAdapterWithFormat[Feedback]("parquet", ""),

		// where to put them (org-scoped)
		builder.S3ClientAdapterWithWriterPrefixTemplate[Feedback](prefix),

		// optional: control the basename (extension added automatically for parquet)
		builder.S3ClientAdapterWithWriterFileNameTemplate[Feedback]("feedback-{ts}-{ulid}"),

		// flush each chunk immediately (we’re sending one blob)
		builder.S3ClientAdapterWithBatchSettings[Feedback](1, 64<<20, 1*time.Second),

		// SSE-KMS (LocalStack accepts this header)
		builder.S3ClientAdapterWithSSE[Feedback](
			"aws:kms",
			"arn:aws:kms:us-east-1:000000000000:alias/electrician-dev",
		),

		builder.S3ClientAdapterWithLogger[Feedback](log),
	)

	// feed the parquet blob as a single chunk
	raw := make(chan []byte, 1)
	raw <- buf.Bytes()
	close(raw)

	if err := writer.ServeWriterRaw(ctx, raw); err != nil {
		panic(err)
	}
}
