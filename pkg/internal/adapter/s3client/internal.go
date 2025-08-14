// pkg/internal/adapter/s3client/internal.go
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
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	parquet "github.com/parquet-go/parquet-go"
)

// --- retry knobs (sensible defaults; can be made configurable later) ---
const (
	defaultMaxAttempts   = 5
	defaultBaseBackoff   = 100 * time.Millisecond
	defaultMaxBackoff    = 3 * time.Second
	defaultJitterEnabled = true
)

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func backoffDuration(attempt int) time.Duration {
	// Exponential backoff (1, 2, 4, 8...) * base, clamped at max; full jitter (0..delay)
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
	// Don't retry explicit cancel/overall deadline
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	// Heuristic string checks that cover common S3 + network failure modes
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "throttl"), // Throttling, ThrottlingException
		strings.Contains(msg, "slowdown"),
		strings.Contains(msg, "timeout"),
		strings.Contains(msg, "tempor"), // temporary network error
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

// putWithRetry performs PutObject with bounded retries + backoff + jitter.
// It re-seeks the body before every attempt (Body must be an io.ReadSeeker).
func (a *S3Client[T]) putWithRetry(
	ctx context.Context,
	put *s3api.PutObjectInput,
	key string,
	payloadSize int,
) (time.Duration, error) {
	var rs io.ReadSeeker
	if r, ok := put.Body.(io.ReadSeeker); ok {
		rs = r
	} else {
		return 0, fmt.Errorf("putWithRetry requires io.ReadSeeker body")
	}

	var lastErr error
	for attempt := 1; attempt <= defaultMaxAttempts; attempt++ {
		// Reset the reader for this attempt
		if _, err := rs.Seek(0, io.SeekStart); err != nil {
			return 0, err
		}

		// sensors: attempt (count every try)
		a.forEachSensor(func(s types.Sensor[T]) {
			s.InvokeOnS3PutAttempt(a.componentMetadata, a.bucket, key, payloadSize, a.sseMode, a.kmsKey)
		})

		start := time.Now()
		_, err := a.cli.PutObject(ctx, put)
		dur := time.Since(start)
		if err == nil {
			// sensors: success
			a.forEachSensor(func(s types.Sensor[T]) {
				s.InvokeOnS3PutSuccess(a.componentMetadata, a.bucket, key, payloadSize, dur)
			})
			return dur, nil
		}

		lastErr = err
		a.NotifyLoggers(types.WarnLevel, "%s => level: WARN, event: PutObject, attempt: %d/%d, key: %s, err: %v",
			a.componentMetadata, attempt, defaultMaxAttempts, key, err)

		// Give up if non-retryable, final attempt, or context done
		if !isRetryable(err) || attempt == defaultMaxAttempts || ctx.Err() != nil {
			a.forEachSensor(func(s types.Sensor[T]) {
				s.InvokeOnS3PutError(a.componentMetadata, a.bucket, key, payloadSize, err)
			})
			return 0, err
		}

		// Backoff w/ jitter (respect context)
		sleep := backoffDuration(attempt)
		select {
		case <-time.After(sleep):
		case <-ctx.Done():
			a.forEachSensor(func(s types.Sensor[T]) {
				s.InvokeOnS3PutError(a.componentMetadata, a.bucket, key, payloadSize, ctx.Err())
			})
			return 0, ctx.Err()
		}
	}
	// Shouldn't reach here, but return lastErr defensively
	return 0, lastErr
}

// writeOne appends one record for legacy NDJSON mode.
// For other formats, use ServeWriterRaw and feed fully-built bytes (e.g., Parquet).
func (a *S3Client[T]) writeOne(v T) error {
	// Only NDJSON is supported for record-oriented writes here.
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

// parquetRoller turns a stream of T into rolling Parquet files.
// Each roll emits a []byte to 'out' (which ServeWriterRaw uploads).
func (a *S3Client[T]) parquetRoller(
	ctx context.Context,
	in <-chan T,
	out chan<- []byte,
	window time.Duration,
	maxRecords int,
	compression parquet.WriterOption,
	compName string, // for sensor reporting
) error {
	var (
		buf   bytes.Buffer
		pw    *parquet.GenericWriter[T]
		count int
	)

	newWriter := func() *parquet.GenericWriter[T] {
		return parquet.NewGenericWriter[T](&buf, compression)
	}
	pw = newWriter()

	flush := func() error {
		if count == 0 {
			return nil
		}
		if err := pw.Close(); err != nil {
			return err
		}

		// sensor: parquet roll flushed
		a.forEachSensor(func(s types.Sensor[T]) {
			s.InvokeOnS3ParquetRollFlush(a.componentMetadata, count, buf.Len(), compName)
		})

		// hand off a stable copy
		out <- append([]byte(nil), buf.Bytes()...)
		a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ParquetFlush, records: %d, bytes: %d",
			a.componentMetadata, count, buf.Len())
		buf.Reset()
		count = 0
		pw = newWriter()
		return nil
	}

	if window <= 0 {
		window = 60 * time.Second
	}
	if maxRecords <= 0 {
		maxRecords = 50_000
	}

	tick := time.NewTicker(window)
	defer tick.Stop()

	for {
		select {
		case <-ctx.Done():
			_ = flush()
			close(out)
			return nil

		case <-tick.C:
			if err := flush(); err != nil {
				close(out)
				return err
			}

		case ev, ok := <-in:
			if !ok {
				err := flush()
				close(out)
				return err
			}
			if _, err := pw.Write([]T{ev}); err != nil {
				close(out)
				return err
			}
			count++
			if count >= maxRecords {
				if err := flush(); err != nil {
					close(out)
					return err
				}
			}
		}
	}
}

// choose parquet compression from formatOpts
func (a *S3Client[T]) parquetCompressionFromOpts() parquet.WriterOption {
	val := ""
	if s, ok := a.formatOpts["compression"]; ok {
		val = s
	} else if s, ok := a.formatOpts["parquet_compression"]; ok {
		val = s
	}
	switch strings.ToLower(val) {
	case "zstd":
		return parquet.Compression(&parquet.Zstd)
	case "gzip", "gz":
		return parquet.Compression(&parquet.Gzip)
	default: // snappy + fallback
		return parquet.Compression(&parquet.Snappy)
	}
}

// name for parquet compression (for sensors/metrics)
func (a *S3Client[T]) parquetCompressionNameFromOpts() string {
	if s, ok := a.formatOpts["compression"]; ok {
		return strings.ToLower(s)
	}
	if s, ok := a.formatOpts["parquet_compression"]; ok {
		return strings.ToLower(s)
	}
	return "snappy"
}

// roll parquet by time/record thresholds and emit []byte chunks
func (a *S3Client[T]) startParquetStream(ctx context.Context, in <-chan T) error {
	// Prefer explicit knobs; fallback to batch settings
	window := 60 * time.Second
	if s, ok := a.formatOpts["roll_window_ms"]; ok {
		if ms, err := strconv.Atoi(s); err == nil && ms > 0 {
			window = time.Duration(ms) * time.Millisecond
		}
	} else if a.batchMaxAge > 0 {
		window = a.batchMaxAge
	}
	max := 50_000
	if s, ok := a.formatOpts["roll_max_records"]; ok {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			max = n
		}
	} else if a.batchMaxRecords > 0 {
		max = a.batchMaxRecords
	}

	comp := a.parquetCompressionFromOpts()
	compName := a.parquetCompressionNameFromOpts()
	raw := make(chan []byte, 2)

	go func() {
		if err := a.parquetRoller(ctx, in, raw, window, max, comp, compName); err != nil {
			a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: parquetRoller, err: %v",
				a.componentMetadata, err)
		}
	}()

	// Each raw chunk -> one object (per ServeWriterRaw above)
	return a.ServeWriterRaw(ctx, raw)
}

// flush uploads the buffered records as a single object (NDJSON path).
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

	// Build payload into bytes so we can rewind on retries
	var payload []byte
	var gz bytes.Buffer

	if a.ndjsonEncGz || strings.EqualFold(a.formatOpts["gzip"], "true") {
		zw := gzip.NewWriter(&gz)
		if _, err := io.Copy(zw, bytes.NewReader(a.buf.Bytes())); err != nil {
			return err
		}
		_ = zw.Close()
		payload = gz.Bytes()
	} else {
		payload = a.buf.Bytes()
	}

	// Prepare request
	reader := bytes.NewReader(payload)
	put := &s3api.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         &key,
		Body:        reader, // io.ReadSeeker
		ContentType: aws.String(ct),
	}
	if a.ndjsonEncGz || strings.EqualFold(a.formatOpts["gzip"], "true") {
		put.ContentEncoding = aws.String("gzip")
	}

	// SSE
	switch strings.ToLower(a.sseMode) {
	case "aes256":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAes256
	case "aws:kms":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAwsKms
		if a.kmsKey != "" {
			put.SSEKMSKeyId = &a.kmsKey
		}
	}

	// sensors: key rendered (once per object)
	a.forEachSensor(func(s types.Sensor[T]) {
		s.InvokeOnS3KeyRendered(a.componentMetadata, key)
	})

	// Put with retry (attempt/success/error sensors inside)
	_, err := a.putWithRetry(ctx, put, key, len(payload))
	if err != nil {
		return err
	}

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Flush, key: %s, records: %d, bytes: %d",
		a.componentMetadata, key, a.count, len(payload))

	a.buf.Reset()
	a.count = 0
	a.lastFlush = now
	return nil
}

// flushRaw uploads a.buf bytes as-is (no extra compression), honoring SSE.
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

	// Copy payload bytes so we can safely rewind on retries
	payload := append([]byte(nil), a.buf.Bytes()...)
	reader := bytes.NewReader(payload)

	put := &s3api.PutObjectInput{
		Bucket:      &a.bucket,
		Key:         &key,
		Body:        reader, // io.ReadSeeker
		ContentType: aws.String(ct),
	}

	// SSE
	switch strings.ToLower(a.sseMode) {
	case "aes256":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAes256
	case "aws:kms":
		put.ServerSideEncryption = s3types.ServerSideEncryptionAwsKms
		if a.kmsKey != "" {
			put.SSEKMSKeyId = &a.kmsKey
		}
	}

	// sensors: key rendered (once per object)
	a.forEachSensor(func(s types.Sensor[T]) {
		s.InvokeOnS3KeyRendered(a.componentMetadata, key)
	})

	// Put with retry (attempt/success/error sensors inside)
	_, err := a.putWithRetry(ctx, put, key, len(payload))
	if err != nil {
		return err
	}

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: FlushRaw, key: %s, chunks: %d, bytes: %d",
		a.componentMetadata, key, a.count, len(payload))

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

// parquetRowsFromBody chooses in-memory vs spill-to-disk based on a threshold knob.
// types.S3ReaderConfig.FormatOptions["spill_threshold_bytes"] can override (default 64 MiB).
func (a *S3Client[T]) parquetRowsFromBody(get *s3api.GetObjectOutput) ([]T, error) {
	const def int64 = 64 << 20 // 64 MiB default spill threshold
	thr := def
	if s, ok := a.readerFormatOpts["spill_threshold_bytes"]; ok {
		if v, err := strconv.ParseInt(s, 10, 64); err == nil && v > 0 {
			thr = v
		}
	}

	// Prefer ContentLength when available; otherwise fall back to buffering.
	if get.ContentLength != nil && *get.ContentLength > thr {
		// Spill to temp file and use ReaderAt.
		f, err := os.CreateTemp("", "electrician-parquet-*")
		if err != nil {
			_ = get.Body.Close()
			return nil, err
		}
		defer func() {
			_ = f.Close()
			_ = os.Remove(f.Name())
		}()
		if _, err := io.Copy(f, get.Body); err != nil {
			_ = get.Body.Close()
			return nil, err
		}
		_ = get.Body.Close()

		if _, err := f.Seek(0, io.SeekStart); err != nil {
			return nil, err
		}
		return a.readParquetAllFromReaderAt(f)
	}

	// Small object: buffer in memory.
	data, err := io.ReadAll(get.Body)
	_ = get.Body.Close()
	if err != nil {
		return nil, err
	}
	return a.readParquetAllFromReaderAt(bytes.NewReader(data))
}

func (a *S3Client[T]) readParquetAllFromReaderAt(ra io.ReaderAt) ([]T, error) {
	gr := parquet.NewGenericReader[T](ra) // size not required in current API
	defer gr.Close()

	out := make([]T, 0, 1024)
	batch := make([]T, 1024)
	for {
		n, err := gr.Read(batch)
		if n > 0 {
			out = append(out, batch[:n]...)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// fanIn copies from src into dst until ctx is done or src closes.
func (a *S3Client[T]) fanIn(ctx context.Context, dst chan<- T, src <-chan T) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-src:
			if !ok {
				return
			}
			select {
			case dst <- v:
			case <-ctx.Done():
				return
			}
		}
	}
}
