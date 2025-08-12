package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
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

	// List objects under prefix
	lo, err := cli.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		panic(err)
	}

	var out []Feedback

	for _, obj := range lo.Contents {
		if obj.Key == nil {
			continue
		}
		key := *obj.Key
		if !strings.EqualFold(filepath.Ext(key), ".parquet") {
			continue
		}

		get, err := cli.GetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			panic(err)
		}

		// Load object into memory; bytes.Reader satisfies io.ReaderAt.
		data, err := io.ReadAll(get.Body)
		_ = get.Body.Close()
		if err != nil {
			panic(err)
		}
		br := bytes.NewReader(data)

		// NewGenericReader requires a ReaderAt (bytes.Reader OK). Some versions also take size.
		gr := parquet.NewGenericReader[Feedback](br /*, int64(len(data)) optional in older API */)

		batch := make([]Feedback, 1024)
		for {
			n, rErr := gr.Read(batch)
			if n > 0 {
				out = append(out, batch[:n]...)
			}
			if rErr == io.EOF {
				break
			}
			if rErr != nil {
				_ = gr.Close()
				panic(rErr)
			}
		}
		if err := gr.Close(); err != nil {
			panic(err)
		}
	}

	// Example transform: uppercase content, drop empties
	filtered := make([]Feedback, 0, len(out))
	for _, r := range out {
		if r.CustomerID == "" && r.Content == "" {
			continue
		}
		r.Content = strings.ToUpper(r.Content)
		filtered = append(filtered, r)
	}

	b, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(b))
}
