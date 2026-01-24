package httpclient

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Serve repeatedly fetches the configured endpoint on an interval.
func (hp *HTTPClientAdapter[T]) Serve(ctx context.Context, submitFunc func(context.Context, T) error) error {
	if submitFunc == nil {
		return fmt.Errorf("submit function is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if !atomic.CompareAndSwapInt32(&hp.isServing, 0, 1) {
		return nil
	}
	defer atomic.StoreInt32(&hp.isServing, 0)

	hp.setContext(ctx)

	cfg := hp.snapshotConfig()
	if cfg.interval <= 0 {
		return fmt.Errorf("interval must be greater than zero")
	}

	ticker := time.NewTicker(cfg.interval)
	defer ticker.Stop()

	hp.NotifyLoggers(types.InfoLevel, "httpclient serve started")

	for {
		select {
		case <-ctx.Done():
			hp.NotifyLoggers(types.WarnLevel, "httpclient serve canceled")
			return nil
		case <-ticker.C:
			if err := hp.attemptFetchAndSubmit(ctx, submitFunc, cfg.maxRetries); err != nil {
				hp.NotifyLoggers(types.ErrorLevel, "httpclient serve failed: %v", err)
				return err
			}
		}
	}
}

func (hp *HTTPClientAdapter[T]) attemptFetchAndSubmit(ctx context.Context, submitFunc func(context.Context, T) error, maxRetries int) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if ctx != nil && ctx.Err() != nil {
			return ctx.Err()
		}

		response, err := hp.Fetch()
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				if err := sleepWithContext(ctx, backoffDuration(attempt)); err != nil {
					return err
				}
				continue
			}
			break
		}

		if err := submitFunc(ctx, response.Body); err != nil {
			lastErr = err
			if attempt < maxRetries {
				if err := sleepWithContext(ctx, backoffDuration(attempt)); err != nil {
					return err
				}
				continue
			}
			break
		}

		return nil
	}

	if lastErr == nil {
		lastErr = fmt.Errorf("retries exhausted")
	}
	return fmt.Errorf("retries exhausted after %d attempts, last error: %v", maxRetries, lastErr)
}
