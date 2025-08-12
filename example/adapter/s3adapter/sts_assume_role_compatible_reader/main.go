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
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Assume role against LocalStack using builder helper
	cli, err := builder.NewS3ClientAssumeRole(
		ctx,
		"us-east-1",
		"arn:aws:iam::000000000000:role/exodus-dev-role",
		"electrician-reader",
		15*time.Minute,
		"", // externalID (optional)
		aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "")), // source creds to call STS
		"http://localhost:4566", // endpoint override for LocalStack
		true,                    // force path-style
	)
	if err != nil {
		panic(err)
	}

	const bucket = "steeze-dev"
	const prefix = "feedback/demo/"

	reader := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithReaderListSettings[Feedback](prefix, "", 1000, 0), // one-shot (no polling)
		builder.S3ClientAdapterWithFormat[Feedback]("ndjson", ""),
	)

	resp, err := reader.Fetch()
	if err != nil {
		panic(err)
	}

	// transform + filter zero-values
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
