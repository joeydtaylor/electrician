package main

import (
	"bytes"
	"context"
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

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

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

	const bucket = "steeze-dev"
	const prefix = "feedback/demo/"

	_, _ = cli.CreateBucket(ctx, &s3.CreateBucketInput{Bucket: aws.String(bucket)})

	records := []Feedback{
		{CustomerID: "cust-001", Content: "I love this", Tags: []string{"raw"}},
		{CustomerID: "cust-002", Content: "Not great", IsNegative: true},
		{CustomerID: "cust-003", Content: "Absolutely fantastic!"},
	}

	// Build Parquet file in memory
	var buf bytes.Buffer
	writer := parquet.NewGenericWriter[Feedback](
		&buf,
		parquet.Compression(&parquet.Snappy), // Snappy compression
	)
	if _, err := writer.Write(records); err != nil {
		panic(err)
	}
	if err := writer.Close(); err != nil {
		panic(err)
	}

	// Configure S3 adapter for parquet (ServeWriterRaw path)
	s3Writer := builder.NewS3ClientAdapter[Feedback](
		ctx,
		builder.S3ClientAdapterWithClientAndBucket[Feedback](cli, bucket),
		builder.S3ClientAdapterWithFormat[Feedback]("parquet", ""),
		builder.S3ClientAdapterWithBatchSettings[Feedback](1, 64<<20, time.Second),
		builder.S3ClientAdapterWithWriterPrefixTemplate[Feedback](prefix),
	)

	raw := make(chan []byte, 1)
	raw <- buf.Bytes()
	close(raw)

	if err := s3Writer.ServeWriterRaw(ctx, raw); err != nil {
		panic(err)
	}
}
