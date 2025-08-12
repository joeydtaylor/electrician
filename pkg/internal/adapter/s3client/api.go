package s3client

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	s3api "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ---------- deps & config setters ----------

func (a *S3Client[T]) SetS3ClientDeps(d types.S3ClientDeps) {
	a.cli = d.Client
	a.bucket = strings.TrimSpace(d.Bucket)
}

func (a *S3Client[T]) SetWriterConfig(c types.S3WriterConfig) {
	if c.PrefixTemplate != "" {
		a.prefixTemplate = c.PrefixTemplate
	}
	if c.FileNameTmpl != "" {
		a.fileNameTmpl = c.FileNameTmpl
	}
	if c.Format != "" {
		a.format = strings.ToLower(c.Format)
	}
	if c.Compression != "" {
		a.compression = strings.ToLower(c.Compression)
	}
	a.sseMode, a.kmsKey = c.SSEMode, c.KMSKeyID
	if c.BatchMaxRecords > 0 {
		a.batchMaxRecords = c.BatchMaxRecords
	}
	if c.BatchMaxBytes > 0 {
		a.batchMaxBytes = c.BatchMaxBytes
	}
	if c.BatchMaxAge > 0 {
		a.batchMaxAge = c.BatchMaxAge
	}
}

func (a *S3Client[T]) SetReaderConfig(c types.S3ReaderConfig) {
	a.listPrefix = c.Prefix
	a.listStartAfter = c.StartAfterKey
	if c.PageSize > 0 {
		a.listPageSize = c.PageSize
	}
	if c.ListInterval > 0 {
		a.listPollInterval = c.ListInterval
	}
	if c.Format != "" {
		a.format = strings.ToLower(c.Format)
	}
	if c.Compression != "" {
		a.compression = strings.ToLower(c.Compression)
	}
}

// ---------- plumbing (loggers/sensors/metadata) ----------

func (a *S3Client[T]) ConnectSensor(s ...types.Sensor[T]) {
	a.sensorLock.Lock()
	defer a.sensorLock.Unlock()
	a.sensors = append(a.sensors, s...)
	for _, m := range s {
		a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: ConnectSensor, target: %v", a.componentMetadata, m.GetComponentMetadata())
	}
}

func (a *S3Client[T]) ConnectLogger(l ...types.Logger) {
	a.loggersLock.Lock()
	defer a.loggersLock.Unlock()
	a.loggers = append(a.loggers, l...)
}

func (a *S3Client[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if len(a.loggers) == 0 {
		return
	}
	msg := fmt.Sprintf(format, args...)
	a.loggersLock.Lock()
	defer a.loggersLock.Unlock()
	for _, logger := range a.loggers {
		if logger == nil || logger.GetLevel() > level {
			continue
		}
		switch level {
		case types.DebugLevel:
			logger.Debug(msg)
		case types.InfoLevel:
			logger.Info(msg)
		case types.WarnLevel:
			logger.Warn(msg)
		case types.ErrorLevel:
			logger.Error(msg)
		case types.DPanicLevel:
			logger.DPanic(msg)
		case types.PanicLevel:
			logger.Panic(msg)
		case types.FatalLevel:
			logger.Fatal(msg)
		}
	}
}

func (a *S3Client[T]) GetComponentMetadata() types.ComponentMetadata { return a.componentMetadata }
func (a *S3Client[T]) SetComponentMetadata(name, id string) {
	a.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}
func (a *S3Client[T]) Name() string { return "S3_CLIENT" }
func (a *S3Client[T]) Stop()        { a.cancel() }

// ---------- writer: channel → S3 ----------

func (a *S3Client[T]) ServeWriter(ctx context.Context, in <-chan T) error {
	if a.cli == nil || a.bucket == "" {
		return fmt.Errorf("s3client: ServeWriter requires client and bucket")
	}
	if !atomic.CompareAndSwapInt32(&a.isServing, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServing, 0)

	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()

	a.lastFlush = time.Now()
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ServeWriter, bucket: %s, prefixTpl: %s", a.componentMetadata, a.bucket, a.prefixTemplate)

	for {
		select {
		case <-ctx.Done():
			if a.count > 0 {
				_ = a.flush(ctx, time.Now())
			}
			return nil

		case v, ok := <-in:
			if !ok {
				if a.count > 0 {
					_ = a.flush(ctx, time.Now())
				}
				return nil
			}
			if err := a.writeOne(v); err != nil {
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: writeOne, err: %v", a.componentMetadata, err)
				return err
			}
			if a.count >= a.batchMaxRecords || a.buf.Len() >= a.batchMaxBytes {
				if err := a.flush(ctx, time.Now()); err != nil {
					return err
				}
			}

		case now := <-tick.C:
			if a.count > 0 && now.Sub(a.lastFlush) >= a.batchMaxAge {
				if err := a.flush(ctx, now); err != nil {
					return err
				}
			}
		}
	}
}

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

// ---------- reader: S3 → submit(T) ----------

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
	lo, err := a.cli.ListObjectsV2(a.ctx, in)
	if err != nil {
		return types.HttpResponse[[]T]{}, err
	}

	var out []T
	for _, obj := range lo.Contents {
		get, err := a.cli.GetObject(a.ctx, &s3api.GetObjectInput{
			Bucket: &a.bucket,
			Key:    obj.Key,
		})
		if err != nil {
			return types.HttpResponse[[]T]{}, err
		}
		rc := get.Body
		var r io.Reader = rc

		if get.ContentEncoding != nil && strings.EqualFold(*get.ContentEncoding, "gzip") {
			gr, e := gzip.NewReader(rc)
			if e != nil {
				rc.Close()
				return types.HttpResponse[[]T]{}, e
			}
			r = gr
		}

		sc := bufio.NewScanner(r)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if line == "" {
				continue
			}
			var v T
			if err := json.Unmarshal([]byte(line), &v); err != nil {
				rc.Close()
				return types.HttpResponse[[]T]{}, err
			}
			out = append(out, v)
		}
		_ = rc.Close()

		if obj.Key != nil {
			a.listStartAfter = *obj.Key // advance cursor
		}
	}

	return types.HttpResponse[[]T]{StatusCode: 200, Body: out}, nil
}

func (a *S3Client[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	if a.cli == nil || a.bucket == "" {
		return fmt.Errorf("s3client: Serve requires client and bucket")
	}
	if !atomic.CompareAndSwapInt32(&a.isServing, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServing, 0)

	tick := time.NewTicker(a.listPollInterval)
	defer tick.Stop()

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ServeReader, prefix: %s", a.componentMetadata, a.listPrefix)

	for {
		select {
		case <-ctx.Done():
			a.NotifyLoggers(types.WarnLevel, "%s => level: WARN, event: Cancel => reader stopped", a.componentMetadata)
			return nil
		case <-tick.C:
			resp, err := a.Fetch()
			if err != nil {
				a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Fetch, err: %v", a.componentMetadata, err)
				continue
			}
			for _, v := range resp.Body {
				if err := submit(ctx, v); err != nil {
					a.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Submit, err: %v", a.componentMetadata, err)
					return err
				}
			}
		}
	}
}
