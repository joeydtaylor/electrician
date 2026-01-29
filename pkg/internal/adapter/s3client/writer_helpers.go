package s3client

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

const (
	defaultMaxAttempts   = 5
	defaultBaseBackoff   = 100 * time.Millisecond
	defaultMaxBackoff    = 3 * time.Second
	defaultJitterEnabled = true
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func backoffDuration(attempt int) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := defaultBaseBackoff << (attempt - 1)
	if d > defaultMaxBackoff {
		d = defaultMaxBackoff
	}
	if defaultJitterEnabled {
		return time.Duration(rng.Int63n(int64(d) + 1))
	}
	return d
}

func isRetryable(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "throttl"),
		strings.Contains(msg, "slowdown"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "tempor"),
		strings.Contains(msg, "connection reset"),
		strings.Contains(msg, "eof"),
		strings.Contains(msg, "internalerror"),
		strings.Contains(msg, "service unavailable"),
		strings.Contains(msg, "503"),
		strings.Contains(msg, "500"):
		return true
	default:
		return false
	}
}

func (a *S3Client[T]) putWithRetry(
	ctx context.Context,
	put *s3api.PutObjectInput,
	key string,
	payloadSize int,
) (time.Duration, error) {
	rs, ok := put.Body.(io.ReadSeeker)
	if !ok {
		return 0, fmt.Errorf("putWithRetry requires io.ReadSeeker body")
	}

	var lastErr error
	for attempt := 1; attempt <= defaultMaxAttempts; attempt++ {
		if _, err := rs.Seek(0, io.SeekStart); err != nil {
			return 0, err
		}

		for _, sensor := range a.snapshotSensors() {
			if sensor == nil {
				continue
			}
			sensor.InvokeOnS3PutAttempt(a.componentMetadata, a.bucket, key, payloadSize, a.sseMode, a.kmsKey)
		}

		start := time.Now()
		_, err := a.cli.PutObject(ctx, put)
		dur := time.Since(start)
		if err == nil {
			for _, sensor := range a.snapshotSensors() {
				if sensor == nil {
					continue
				}
				sensor.InvokeOnS3PutSuccess(a.componentMetadata, a.bucket, key, payloadSize, dur)
			}
			return dur, nil
		}

		lastErr = err
		a.NotifyLoggers(
			types.WarnLevel,
			"PutObject retry",
			"component", a.componentMetadata,
			"event", "PutObject",
			"attempt", attempt,
			"max_attempts", defaultMaxAttempts,
			"key", key,
			"error", err,
		)

		if !isRetryable(err) || attempt == defaultMaxAttempts || ctx.Err() != nil {
			for _, sensor := range a.snapshotSensors() {
				if sensor == nil {
					continue
				}
				sensor.InvokeOnS3PutError(a.componentMetadata, a.bucket, key, payloadSize, err)
			}
			return 0, err
		}

		sleep := backoffDuration(attempt)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			for _, sensor := range a.snapshotSensors() {
				if sensor == nil {
					continue
				}
				sensor.InvokeOnS3PutError(a.componentMetadata, a.bucket, key, payloadSize, ctx.Err())
			}
			return 0, ctx.Err()
		}
	}

	return 0, lastErr
}

func (a *S3Client[T]) writeOne(v T) error {
	if !strings.EqualFold(a.formatName, "ndjson") {
		return fmt.Errorf("record-oriented writes only support ndjson; got %q (use ServeWriterRaw for %q)", a.formatName, a.formatName)
	}

	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, _ = a.buf.Write(b)
	_ = a.buf.WriteByte('\n')
	a.count++
	return nil
}

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

	var payload []byte
	var gz bytes.Buffer
	var contentEncoding string

	if a.ndjsonEncGz || strings.EqualFold(a.formatOpts["gzip"], "true") {
		zw := gzip.NewWriter(&gz)
		if _, err := io.Copy(zw, bytes.NewReader(a.buf.Bytes())); err != nil {
			return err
		}
		_ = zw.Close()
		payload = gz.Bytes()
		contentEncoding = "gzip"
	} else {
		payload = a.buf.Bytes()
	}

	payload, ct, contentEncoding, meta, err := a.applyCSE(payload, ct, contentEncoding)
	if err != nil {
		return err
	}

	reader := bytes.NewReader(payload)
	put := &s3api.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         &key,
		Body:        reader,
		ContentType: aws.String(ct),
	}
	if contentEncoding != "" {
		put.ContentEncoding = aws.String(contentEncoding)
	}
	if len(meta) > 0 {
		put.Metadata = meta
	}

	switch strings.ToLower(a.sseMode) {
	case "aes256":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAes256
	case "aws:kms":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAwsKms
		if a.kmsKey != "" {
			put.SSEKMSKeyId = &a.kmsKey
		}
	}

	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnS3KeyRendered(a.componentMetadata, key)
	}

	_, err = a.putWithRetry(ctx, put, key, len(payload))
	if err != nil {
		return err
	}

	a.NotifyLoggers(
		types.InfoLevel,
		"Flush",
		"component", a.componentMetadata,
		"event", "Flush",
		"key", key,
		"records", a.count,
		"bytes", len(payload),
	)

	a.buf.Reset()
	a.count = 0
	a.lastFlush = now
	return nil
}

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

	payload := append([]byte(nil), a.buf.Bytes()...)
	payload, ct, _, meta, err := a.applyCSE(payload, ct, "")
	if err != nil {
		return err
	}
	reader := bytes.NewReader(payload)

	put := &s3api.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         &key,
		Body:        reader,
		ContentType: aws.String(ct),
	}
	if len(meta) > 0 {
		put.Metadata = meta
	}

	switch strings.ToLower(a.sseMode) {
	case "aes256":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAes256
	case "aws:kms":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAwsKms
		if a.kmsKey != "" {
			put.SSEKMSKeyId = &a.kmsKey
		}
	}

	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnS3KeyRendered(a.componentMetadata, key)
	}

	_, err = a.putWithRetry(ctx, put, key, len(payload))
	if err != nil {
		return err
	}

	a.NotifyLoggers(
		types.InfoLevel,
		"FlushRaw",
		"component", a.componentMetadata,
		"event", "FlushRaw",
		"key", key,
		"chunks", a.count,
		"bytes", len(payload),
	)

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
		"{ulid}": utils.GenerateUniqueHash(),
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
