package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

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

const (
	storjEndpoint = "https://gateway.storjshare.io"
	storjRegion   = "us-east-1"

	storjAccessKey = "REPLACE_WITH_STORJ_ACCESS_KEY"
	storjSecretKey = "REPLACE_WITH_STORJ_SECRET_KEY"
	storjBucket    = "REPLACE_WITH_BUCKET"

	clientSideKeyHex = "REPLACE_WITH_32_BYTE_HEX_KEY"
)

func mustSet(label, v string) string {
	if strings.Contains(v, "REPLACE_WITH") {
		panic(fmt.Sprintf("%s must be set before running", label))
	}
	return v
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	log := builder.NewLogger(builder.LoggerWithDevelopment(true))

	cli, err := builder.NewS3ClientStorj(ctx, builder.StorjS3Config{
		Endpoint:   storjEndpoint,
		Region:     storjRegion,
		AccessKey:  mustSet("storjAccessKey", storjAccessKey),
		SecretKey:  mustSet("storjSecretKey", storjSecretKey),
		RequireTLS: builder.BoolPtr(true),
		HTTPClient: builder.NewTLSHTTPClient(builder.TLSHTTPClientConfig{
			MinVersion: tls.VersionTLS12,
			Timeout:    30 * time.Second,
		}),
	})
	if err != nil {
		panic(err)
	}

	bucket := mustSet("storjBucket", storjBucket)
	prefix := "org=demo/feedback/storj/{yyyy}/{MM}/{dd}/{HH}/{mm}/"

	records := []Feedback{
		{CustomerID: "cust-001", Content: "Secure Storj payload", Tags: []string{"storj", "secure"}},
		{CustomerID: "cust-002", Content: "HIPAA-style example", Category: "feedback"},
	}

	var buf bytes.Buffer
	pw := parquet.NewGenericWriter[Feedback](&buf, parquet.Compression(&parquet.Zstd))
	if _, err := pw.Write(records); err != nil {
		panic(err)
	}
	if err := pw.Close(); err != nil {
		panic(err)
	}

	writer := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithFormat[Feedback]("parquet", ""),
		builder.S3ClientAdapterWithWriterPrefixTemplate[Feedback](prefix),
		builder.S3ClientAdapterWithWriterFileNameTemplate[Feedback]("feedback-{ts}-{ulid}"),
		builder.S3ClientAdapterWithBatchSettings[Feedback](1, 64<<20, 1*time.Second),
		builder.S3ClientAdapterWithSSE[Feedback]("AES256", ""),
		builder.S3ClientAdapterWithRequireSSE[Feedback](true),
		builder.S3ClientAdapterWithStorjSecureDefaults[Feedback](mustSet("clientSideKeyHex", clientSideKeyHex)),
		builder.S3ClientAdapterWithLogger[Feedback](log),
	)

	raw := make(chan []byte, 1)
	raw <- buf.Bytes()
	close(raw)

	if err := writer.ServeWriterRaw(ctx, raw); err != nil {
		panic(err)
	}

	fmt.Println("Storj write complete.")
}
