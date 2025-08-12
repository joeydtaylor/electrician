package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

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
		"electrician-writer",
		15*time.Minute,
		"", // externalID (optional)
		aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider("test", "test", "")), // source creds to call STS
		"http://localhost:4566", // endpoint override for LocalStack
		true,                    // force path-style for LocalStack/MinIO
	)
	if err != nil {
		panic(err)
	}

	const bucket = "steeze-dev"
	const prefix = "feedback/demo/"

	// Ensure bucket exists (idempotent for LocalStack)
	_, _ = cli.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})

	// Seed data
	records := []Feedback{
		{CustomerID: "cust-001", Content: "I love this", Tags: []string{"raw"}},
		{CustomerID: "cust-002", Content: "Not great", IsNegative: true},
		{CustomerID: "cust-003", Content: "Absolutely fantastic!"},
	}

	// Writer: NDJSON, small batch, specific prefix
	writer := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithFormat[Feedback]("ndjson", ""), // no gzip for demo; set "gzip" if you like
		builder.S3ClientAdapterWithBatchSettings[Feedback](1000, 1<<20, time.Second),
		builder.S3ClientAdapterWithWriterPrefixTemplate[Feedback](prefix),
	)

	// Feed the channel
	ch := make(chan Feedback, len(records))
	go func() {
		defer close(ch)
		for _, r := range records {
			if r.CustomerID == "" && r.Content == "" {
				continue
			}
			ch <- r
		}
	}()

	if err := writer.ServeWriter(ctx, ch); err != nil {
		panic(err)
	}
}
