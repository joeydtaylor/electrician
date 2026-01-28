package s3client

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ServeWriter consumes records from in and writes NDJSON objects to S3.
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
	a.NotifyLoggers(
		types.InfoLevel,
		"ServeWriter",
		"component", a.componentMetadata,
		"event", "ServeWriter",
		"bucket", a.bucket,
		"prefix_template", a.prefixTemplate,
	)

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
				a.NotifyLoggers(
					types.ErrorLevel,
					"writeOne failed",
					"component", a.componentMetadata,
					"event", "writeOne",
					"error", err,
				)
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

// StartWriter fans in connected wires and serves records to S3.
func (a *S3Client[T]) StartWriter(ctx context.Context) error {
	if a.cli == nil || a.bucket == "" {
		return fmt.Errorf("s3client: StartWriter requires client and bucket")
	}
	if len(a.inputWires) == 0 {
		return fmt.Errorf("s3client: StartWriter requires at least one connected wire; call ConnectInput(...)")
	}
	if a.mergedIn == nil {
		size := a.batchMaxRecords
		if size <= 0 {
			size = 1024
		}
		a.mergedIn = make(chan T, size)
	}

	for _, w := range a.inputWires {
		if w == nil {
			continue
		}
		out := w.GetOutputChannel()
		go a.fanIn(ctx, a.mergedIn, out)
	}

	a.NotifyLoggers(
		types.InfoLevel,
		"StartWriter",
		"component", a.componentMetadata,
		"event", "StartWriter",
		"format", a.formatName,
		"wires", len(a.inputWires),
		"bucket", a.bucket,
		"prefix_template", a.prefixTemplate,
	)

	for _, sensor := range a.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnS3WriterStart(a.componentMetadata, a.bucket, a.prefixTemplate, strings.ToLower(a.formatName))
	}

	switch strings.ToLower(a.formatName) {
	case "", "ndjson":
		go func() {
			if err := a.ServeWriter(ctx, a.mergedIn); err != nil {
				a.NotifyLoggers(
					types.ErrorLevel,
					"ServeWriter failed",
					"component", a.componentMetadata,
					"event", "ServeWriter",
					"error", err,
				)
			}
		}()
		return nil
	case "parquet":
		go func() {
			if err := a.startParquetStream(ctx, a.mergedIn); err != nil {
				a.NotifyLoggers(
					types.ErrorLevel,
					"startParquetStream failed",
					"component", a.componentMetadata,
					"event", "startParquetStream",
					"error", err,
				)
			}
		}()
		return nil
	default:
		return fmt.Errorf("s3client: StartWriter unsupported format %q (use ndjson or parquet)", a.formatName)
	}
}
