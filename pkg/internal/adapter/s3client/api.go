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
		a.formatName = strings.ToLower(c.Format)
	}
	// merge encoder knobs
	if len(c.FormatOptions) > 0 {
		if a.formatOpts == nil {
			a.formatOpts = make(map[string]string, len(c.FormatOptions))
		}
		for k, v := range c.FormatOptions {
			a.formatOpts[k] = v
		}
	}
	// legacy compression hint (NDJSON only)
	if c.Compression != "" {
		a.ndjsonEncGz = strings.EqualFold(c.Compression, "gzip")
		if a.ndjsonEncGz {
			if a.formatOpts == nil {
				a.formatOpts = map[string]string{}
			}
			a.formatOpts["gzip"] = "true"
		}
	}

	// SSE
	a.sseMode, a.kmsKey = c.SSEMode, c.KMSKeyID

	// batching
	if c.BatchMaxRecords > 0 {
		a.batchMaxRecords = c.BatchMaxRecords
	}
	if c.BatchMaxBytes > 0 {
		a.batchMaxBytes = c.BatchMaxBytes
	}
	if c.BatchMaxAge > 0 {
		a.batchMaxAge = c.BatchMaxAge
	}

	// raw passthrough defaults/overrides
	if c.RawExtension != "" {
		a.rawWriterExt = c.RawExtension
	}
	if c.RawContentType != "" {
		a.rawWriterContentType = c.RawContentType
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
		a.readerFormatName = strings.ToLower(c.Format)
	}
	// merge decoder knobs
	if len(c.FormatOptions) > 0 {
		if a.readerFormatOpts == nil {
			a.readerFormatOpts = make(map[string]string, len(c.FormatOptions))
		}
		for k, v := range c.FormatOptions {
			a.readerFormatOpts[k] = v
		}
	}
	// legacy compression hint (NDJSON only)
	if c.Compression != "" && strings.EqualFold(c.Compression, "gzip") {
		if a.readerFormatOpts == nil {
			a.readerFormatOpts = map[string]string{}
		}
		a.readerFormatOpts["gzip"] = "true"
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

func (a *S3Client[T]) Stop() { a.cancel() }

// ---------- writer: channel → S3 (structured) ----------

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
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ServeWriter, bucket: %s, prefixTpl: %s",
		a.componentMetadata, a.bucket, a.prefixTemplate)

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

// ---------- writer: channel → S3 (raw bytes; parquet/anything) ----------

func (a *S3Client[T]) ServeWriterRaw(ctx context.Context, in <-chan []byte) error {
	if a.cli == nil || a.bucket == "" {
		return fmt.Errorf("s3client: ServeWriterRaw requires client and bucket")
	}
	if !atomic.CompareAndSwapInt32(&a.isServing, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServing, 0)

	tick := time.NewTicker(200 * time.Millisecond)
	defer tick.Stop()

	a.lastFlush = time.Now()
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ServeWriterRaw, bucket: %s, prefixTpl: %s",
		a.componentMetadata, a.bucket, a.prefixTemplate)

	for {
		select {
		case <-ctx.Done():
			if a.buf.Len() > 0 {
				_ = a.flushRaw(ctx, time.Now())
			}
			return nil

		case chunk, ok := <-in:
			if !ok {
				if a.buf.Len() > 0 {
					_ = a.flushRaw(ctx, time.Now())
				}
				return nil
			}
			if len(chunk) > 0 {
				a.buf.Write(chunk)
				a.count++ // count chunks; primarily for thresholds/metrics
			}
			if a.count >= a.batchMaxRecords || a.buf.Len() >= a.batchMaxBytes {
				if err := a.flushRaw(ctx, time.Now()); err != nil {
					return err
				}
			}

		case now := <-tick.C:
			if a.buf.Len() > 0 && now.Sub(a.lastFlush) >= a.batchMaxAge {
				if err := a.flushRaw(ctx, now); err != nil {
					return err
				}
			}
		}
	}
}

// ---------- reader: S3 → submit(T) ----------

// api.go — drop-in replacement
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

			// Parquet path (by extension or explicit reader format)
			if ext == ".parquet" || strings.EqualFold(a.readerFormatName, "parquet") {
				rows, err := a.parquetRowsFromBody(get)
				if err != nil {
					return types.HttpResponse[[]T]{}, err
				}
				out = append(out, rows...)
				seen += len(rows)
				continue
			}

			// NDJSON path (gzip-aware)
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
			buf := make([]byte, 0, 1<<20) // 1 MiB initial
			sc.Buffer(buf, 16<<20)        // 16 MiB max line
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

		// pagination (IsTruncated is *bool)
		if aws.ToBool(lo.IsTruncated) && lo.NextContinuationToken != nil {
			in.ContinuationToken = lo.NextContinuationToken
			continue
		}
		break
	}

	if lastKey != "" {
		a.listStartAfter = lastKey // advance cursor
	}

	if seen == 0 {
		return types.HttpResponse[[]T]{StatusCode: 204, Body: nil}, nil
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

	// One-shot mode if no poll interval set.
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

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ServeReader, prefix: %s",
		a.componentMetadata, a.listPrefix)

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
