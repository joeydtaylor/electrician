package s3client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// writeOne appends one NDJSON record (Parquet can plug in later).
func (a *S3Client[T]) writeOne(v T) error {
	if a.format != "ndjson" {
		return fmt.Errorf("format %q not implemented", a.format)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	a.buf.Write(b)
	a.buf.WriteByte('\n')
	a.count++
	return nil
}

func (a *S3Client[T]) flush(ctx context.Context, now time.Time) error {
	if a.count == 0 {
		return nil
	}
	key := a.renderKey(now)

	var body io.Reader = bytes.NewReader(a.buf.Bytes())
	var gz bytes.Buffer

	put := &s3api.PutObjectInput{
		Bucket: &a.bucket,
		Key:    &key,
	}

	// compression (ndjson)
	if strings.EqualFold(a.compression, "gzip") {
		zw := gzip.NewWriter(&gz)
		if _, err := io.Copy(zw, body); err != nil {
			return err
		}
		_ = zw.Close()
		put.Body = bytes.NewReader(gz.Bytes())
		put.ContentEncoding = aws.String("gzip")
	} else {
		put.Body = body
	}

	// server-side encryption
	switch strings.ToLower(a.sseMode) {
	case "aes256":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAes256
	case "aws:kms":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAwsKms
		if a.kmsKey != "" {
			put.SSEKMSKeyId = &a.kmsKey
		}
	}

	if _, err := a.cli.PutObject(ctx, put); err != nil {
		return err
	}

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Flush, key: %s, records: %d, bytes: %d",
		a.componentMetadata, key, a.count, a.buf.Len())

	a.buf.Reset()
	a.count = 0
	a.lastFlush = now
	return nil
}

func (a *S3Client[T]) renderKey(now time.Time) string {
	ts := now.UTC()
	repl := map[string]string{
		"{yyyy}": ts.Format("2006"),
		"{MM}":   ts.Format("01"),
		"{dd}":   ts.Format("02"),
		"{HH}":   ts.Format("15"),
		"{mm}":   ts.Format("04"),
		"{ts}":   fmt.Sprintf("%d", ts.UnixMilli()),
		"{ulid}": utils.GenerateUniqueHash(), // reuse existing util for uniqueness
	}
	prefix := a.prefixTemplate
	for k, v := range repl {
		prefix = strings.ReplaceAll(prefix, k, v)
	}
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}
	name := a.fileNameTmpl
	for k, v := range repl {
		name = strings.ReplaceAll(name, k, v)
	}
	return path.Join(prefix, name)
}
