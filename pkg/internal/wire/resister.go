package wire

import (
	"context"
	"time"
)

// runResisterLoop drains the resister queue based on rate limits.
func (w *Wire[T]) runResisterLoop(ctx context.Context) {
	sp := w.surgeProtector
	if sp == nil {
		return
	}

	if !sp.IsBeingRateLimited() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		w.drainResisterQueue(ctx, ticker)
		return
	}

	_, fillrate, _, _ := sp.GetRateLimit()
	if fillrate <= 0 {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		w.drainResisterQueue(ctx, ticker)
		return
	}

	ticker := time.NewTicker(fillrate)
	defer ticker.Stop()
	w.drainResisterQueue(ctx, ticker)
}

func (w *Wire[T]) drainResisterQueue(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sp := w.surgeProtector
			if sp == nil {
				return
			}

			if !sp.TryTake() {
				if w.resisterQueueEmpty() {
					return
				}
				continue
			}

			item, err := sp.Dequeue()
			if err != nil {
				if w.resisterQueueEmpty() {
					return
				}
				continue
			}

			_ = w.submitFromResisterQueue(ctx, item.Data)
		}
	}
}

func (w *Wire[T]) resisterQueueEmpty() bool {
	sp := w.surgeProtector
	if sp == nil {
		return true
	}
	return sp.GetResisterQueue() == 0
}
