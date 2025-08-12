package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
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
	const prefix = "feedback/demo/" // match writer

	reader := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithReaderListSettings[Feedback](prefix, "", 1000, 0),
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
