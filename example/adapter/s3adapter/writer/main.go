package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

	// LocalStack S3 client with static creds
	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...interface{}) (aws.Endpoint, error) {
		if service == s3.ServiceID {
			return aws.Endpoint{URL: "http://localhost:4566", HostnameImmutable: true}, nil
		}
		return aws.Endpoint{}, &aws.EndpointNotFoundError{}
	})
	awsCfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("us-east-1"),
		config.WithEndpointResolverWithOptions(resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		panic(err)
	}
	cli := s3.NewFromConfig(awsCfg, func(o *s3.Options) { o.UsePathStyle = true })

	const bucket = "steeze-dev"
	const prefix = "feedback/demo/" // scope writes here

	_, _ = cli.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})

	// seed
	records := []Feedback{
		{CustomerID: "cust-001", Content: "I love this", Tags: []string{"raw"}},
		{CustomerID: "cust-002", Content: "Not great", IsNegative: true},
		{CustomerID: "cust-003", Content: "Absolutely fantastic!"},
	}

	// writer: NDJSON, no compression, scoped prefix, small batch
	writer := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithFormat[Feedback]("ndjson", ""),
		builder.S3ClientAdapterWithBatchSettings[Feedback](1000, 1<<20, 1*time.Second),
		builder.S3ClientAdapterWithWriterPrefixTemplate[Feedback](prefix),
	)

	// channel feed; skip zero-values
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
