package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

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
	prefix := "org=demo/feedback/storj/"

	reader := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithReaderListSettings[Feedback](prefix, ".parquet", 1000, 0),
		builder.S3ClientAdapterWithReaderFormat[Feedback]("parquet", ""),
		builder.S3ClientAdapterWithReaderFormatOptions[Feedback](map[string]string{
			"spill_threshold_bytes": "134217728",
		}),
		builder.S3ClientAdapterWithStorjSecureDefaults[Feedback](mustSet("clientSideKeyHex", clientSideKeyHex)),
		builder.S3ClientAdapterWithLogger[Feedback](log),
	)

	resp, err := reader.Fetch()
	if err != nil {
		panic(err)
	}
	fmt.Printf("read status=%d rows=%d\n", resp.StatusCode, len(resp.Body))
	for i := 0; i < len(resp.Body) && i < 3; i++ {
		fmt.Printf("sample[%d]=%+v\n", i, resp.Body[i])
	}
}
