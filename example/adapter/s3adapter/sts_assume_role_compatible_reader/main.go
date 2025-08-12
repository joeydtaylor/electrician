package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type Feedback struct {
	CustomerID string   `parquet:"name=customerId, type=BYTE_ARRAY, convertedtype=UTF8" json:"customerId"`
	Content    string   `parquet:"name=content, type=BYTE_ARRAY, convertedtype=UTF8" json:"content"`
	Category   string   `parquet:"name=category, type=BYTE_ARRAY, convertedtype=UTF8" json:"category,omitempty"`
	IsNegative bool     `parquet:"name=isNegative, type=BOOLEAN" json:"isNegative"`
	Tags       []string `parquet:"name=tags, type=LIST, valuetype=BYTE_ARRAY, valueconvertedtype=UTF8" json:"tags,omitempty"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cli, err := builder.NewS3ClientAssumeRole(
		ctx,
		"us-east-1",
		"arn:aws:iam::000000000000:role/exodus-dev-role",
		"electrician-reader",
		15*time.Minute,
		"",
		aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "")),
		"http://localhost:4566",
		true,
	)
	if err != nil {
		panic(err)
	}

	const bucket = "steeze-dev"
	const prefix = "feedback/demo/"

	reader := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithReaderListSettings[Feedback](prefix, "", 1000, 0), // one-shot
		builder.S3ClientAdapterWithFormat[Feedback]("parquet", ""),
		// spill to disk if object > 128 MiB
		builder.S3ClientAdapterWithReaderFormatOptions[Feedback](map[string]string{
			"spill_threshold_bytes": "134217728",
		}),
		// SSE-KMS is transparent on reads; no extra option needed here.
	)

	resp, err := reader.Fetch()
	if err != nil {
		panic(err)
	}

	out := make([]Feedback, 0, len(resp.Body))
	for _, r := range resp.Body {
		if r.CustomerID == "" && r.Content == "" {
			continue
		}
		r.Content = strings.ToUpper(r.Content)
		out = append(out, r)
	}

	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
