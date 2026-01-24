package s3client

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ServeWriterRaw writes raw byte chunks to S3 (one object per chunk).
func (a *S3Client[T]) ServeWriterRaw(ctx context.Context, in <-chan []byte) error {
	if a.cli == nil || a.bucket == "" {
		return fmt.Errorf("s3client: ServeWriterRaw requires client and bucket")
	}
	if !atomic.CompareAndSwapInt32(&a.isServing, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&a.isServing, 0)

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ServeWriterRaw, bucket: %s, prefixTpl: %s",
		a.componentMetadata, a.bucket, a.prefixTemplate)

	for {
		select {
		case <-ctx.Done():
			return nil
		case chunk, ok := <-in:
			if !ok {
				return nil
			}
			if len(chunk) == 0 {
				continue
			}

			a.buf.Reset()
			if _, err := a.buf.Write(chunk); err != nil {
				return err
			}
			a.count = 1

			if err := a.flushRaw(ctx, time.Now()); err != nil {
				return err
			}
		}
	}
}
