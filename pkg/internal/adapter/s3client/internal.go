// pkg/internal/adapter/s3client/internal.go
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

// writeOne appends one record for legacy NDJSON mode.
// If/when a.format (types.Format[T]) is wired, this will be replaced by the pluggable encoder path.
func (a *S3Client[T]) writeOne(v T) error {
	// Legacy NDJSON only (record-oriented)
	if a.format != nil && a.formatName != "ndjson" {
		return fmt.Errorf("record encoder for format %q not wired yet", a.formatName)
	}
	if a.formatName != "ndjson" {
		return fmt.Errorf("format %q not implemented for record-oriented writes; use ServeWriterRaw for bytes", a.formatName)
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

// flush uploads the buffered records as a single object.
// Legacy path supports NDJSON (+ optional gzip). Extension and content-type are derived.
func (a *S3Client[T]) flush(ctx context.Context, now time.Time) error {
	if a.count == 0 {
		return nil
	}
	if a.formatName != "ndjson" {
		return fmt.Errorf("flush: record-oriented encoder for format %q not implemented", a.formatName)
	}

	ext := ".ndjson"
	ct := a.ndjsonMime
	key := a.renderKey(now) + ext

	var body io.Reader = bytes.NewReader(a.buf.Bytes())
	var gz bytes.Buffer

	put := &s3api.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         &key,
		ContentType: aws.String(ct),
	}

	// compression (ndjson)
	if a.ndjsonEncGz || strings.EqualFold(a.formatOpts["gzip"], "true") {
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

// flushRaw uploads a.buf bytes as-is (no extra compression), honoring SSE.
// Extension and content type come from raw defaults or the parquet preset.
func (a *S3Client[T]) flushRaw(ctx context.Context, now time.Time) error {
	if a.buf.Len() == 0 {
		return nil
	}

	ext := a.rawWriterExt
	if ext == "" {
		if a.formatName == "parquet" {
			ext = ".parquet"
		} else {
			ext = ".bin"
		}
	}
	ct := a.rawWriterContentType
	if ct == "" {
		if a.formatName == "parquet" {
			ct = "application/parquet"
		} else {
			ct = "application/octet-stream"
		}
	}

	key := a.renderKey(now) + ext

	put := &s3api.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         &key,
		Body:        bytes.NewReader(a.buf.Bytes()),
		ContentType: aws.String(ct),
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

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: FlushRaw, key: %s, chunks: %d, bytes: %d",
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
	if name == "" {
		name = "{ts}-{ulid}"
	}
	for k, v := range repl {
		name = strings.ReplaceAll(name, k, v)
	}
	return path.Join(prefix, name)
}
