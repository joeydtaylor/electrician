package s3client

import (
	"bufio"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Fetch lists objects and returns decoded records in a single response.
func (a *S3Client[T]) Fetch() (types.HttpResponse[[]T], error) {
	if a.cli == nil || a.bucket == "" {
		return types.HttpResponse[[]T]{}, fmt.Errorf("s3client: Fetch requires client and bucket")
	}

	in := &s3api.ListObjectsV2Input{
		Bucket:     &a.bucket,
		Prefix:     aws.String(a.listPrefix),
		StartAfter: aws.String(a.listStartAfter),
		MaxKeys:    aws.Int32(a.listPageSize),
	}

	var out []T
	var lastKey string
	seen := 0

	for {
		lo, err := a.cli.ListObjectsV2(a.ctx, in)
		if err != nil {
			return types.HttpResponse[[]T]{}, err
		}

		for _, obj := range lo.Contents {
			if obj.Key == nil {
				continue
			}
			key := *obj.Key
			lastKey = key
			ext := strings.ToLower(filepath.Ext(key))

			get, err := a.cli.GetObject(a.ctx, &s3api.GetObjectInput{
				Bucket: &a.bucket,
				Key:    obj.Key,
			})
			if err != nil {
				return types.HttpResponse[[]T]{}, err
			}

			if ext == ".parquet" || strings.EqualFold(a.readerFormatName, "parquet") {
				rows, err := a.parquetRowsFromBody(get)
				if err != nil {
					return types.HttpResponse[[]T]{}, err
				}
				out = append(out, rows...)
				seen += len(rows)
				continue
			}

			rc := get.Body
			var r io.Reader = rc

			var gz *gzip.Reader
			if get.ContentEncoding != nil && strings.EqualFold(*get.ContentEncoding, "gzip") {
				gz, err = gzip.NewReader(rc)
				if err != nil {
					_ = rc.Close()
					return types.HttpResponse[[]T]{}, err
				}
				r = gz
			}

			sc := bufio.NewScanner(r)
			buf := make([]byte, 0, 1<<20)
			sc.Buffer(buf, 16<<20)
			for sc.Scan() {
				line := strings.TrimSpace(sc.Text())
				if line == "" {
					continue
				}
				var v T
				if err := json.Unmarshal([]byte(line), &v); err != nil {
					if gz != nil {
						_ = gz.Close()
					}
					_ = rc.Close()
					return types.HttpResponse[[]T]{}, err
				}
				out = append(out, v)
				seen++
			}
			if err := sc.Err(); err != nil {
				if gz != nil {
					_ = gz.Close()
				}
				_ = rc.Close()
				return types.HttpResponse[[]T]{}, err
			}
			if gz != nil {
				_ = gz.Close()
			}
			_ = rc.Close()
		}

		if aws.ToBool(lo.IsTruncated) && lo.NextContinuationToken != nil {
			in.ContinuationToken = lo.NextContinuationToken
			continue
		}
		break
	}

	if lastKey != "" {
		a.listStartAfter = lastKey
	}

	if seen == 0 {
		return types.HttpResponse[[]T]{StatusCode: 204, Body: nil}, nil
	}
	return types.HttpResponse[[]T]{StatusCode: 200, Body: out}, nil
}

// Serve polls S3 for objects and streams decoded records to submit.
func (a *S3Client[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	if a.cli == nil || a.bucket == "" {
		return fmt.Errorf("s3client: Serve requires client and bucket")
	}
	if !atomic.CompareAndSwapInt32(&a.isServing, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServing, 0)

	if a.listPollInterval <= 0 {
		resp, err := a.Fetch()
		if err != nil {
			return err
		}
		for _, v := range resp.Body {
			if err := submit(ctx, v); err != nil {
				return err
			}
		}
		return nil
	}

	tick := time.NewTicker(a.listPollInterval)
	defer tick.Stop()

	a.NotifyLoggers(
		types.InfoLevel,
		"ServeReader",
		"component", a.componentMetadata,
		"event", "ServeReader",
		"prefix", a.listPrefix,
	)

	for {
		select {
		case <-ctx.Done():
			a.NotifyLoggers(
				types.WarnLevel,
				"ServeReader cancelled",
				"component", a.componentMetadata,
				"event", "ServeReaderStop",
				"result", "CANCELLED",
			)
			return nil
		case <-tick.C:
			resp, err := a.Fetch()
			if err != nil {
				a.NotifyLoggers(
					types.ErrorLevel,
					"Fetch failed",
					"component", a.componentMetadata,
					"event", "Fetch",
					"error", err,
				)
				continue
			}
			for _, v := range resp.Body {
				if err := submit(ctx, v); err != nil {
					a.NotifyLoggers(
						types.ErrorLevel,
						"Submit failed",
						"component", a.componentMetadata,
						"event", "Submit",
						"error", err,
					)
					return err
				}
			}
		}
	}
}
