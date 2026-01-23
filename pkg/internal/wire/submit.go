package wire

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// FastSubmit bypasses optional components when they are not configured.
func (w *Wire[T]) FastSubmit(ctx context.Context, elem T) error {
	if w.CircuitBreaker != nil || w.surgeProtector != nil || w.insulatorFunc != nil {
		return w.Submit(ctx, elem)
	}
	if atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0 {
		return w.Submit(ctx, elem)
	}

	in := w.inChan
	select {
	case in <- elem:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-w.ctx.Done():
		return w.ctx.Err()
	}
}

// Submit enqueues an element for processing.
func (w *Wire[T]) Submit(ctx context.Context, elem T) error {
	if cb := w.CircuitBreaker; cb != nil && !cb.Allow() {
		return w.handleCircuitBreakerTrip(ctx, elem)
	}

	if sp := w.surgeProtector; sp != nil {
		element := types.NewElementFast[T](elem)

		if sp.IsTripped() {
			w.notifySurgeProtectorSubmit(element.Data)
			return sp.Submit(ctx, element)
		}

		if sp.IsBeingRateLimited() && !sp.TryTake() {
			w.notifyRateLimit(elem, time.Now().Add(sp.GetTimeUntilNextRefill()).Format("2006-01-02 15:04:05"))
			return sp.Submit(ctx, element)
		}
	}

	return w.submitNormally(ctx, elem)
}

func (w *Wire[T]) handleCircuitBreakerTrip(ctx context.Context, elem T) error {
	cb := w.CircuitBreaker
	if cb == nil {
		return nil
	}

	neutralWires := cb.GetNeutralWires()
	if len(neutralWires) == 0 {
		w.notifyCircuitBreakerDropElement(elem)
		return nil
	}

	for _, gw := range neutralWires {
		if gw == nil {
			continue
		}
		if err := gw.Submit(ctx, elem); err != nil {
			continue
		}
		w.notifyNeutralWireSubmission(elem)
	}

	return nil
}

func (w *Wire[T]) submitFromResisterQueue(ctx context.Context, elem T) error {
	if cb := w.CircuitBreaker; cb != nil && !cb.Allow() {
		return w.handleCircuitBreakerTrip(ctx, elem)
	}
	return w.submitNormally(ctx, elem)
}

func (w *Wire[T]) submitNormally(ctx context.Context, elem T) error {
	select {
	case w.inChan <- elem:
		if atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0 {
			w.notifySubmit(elem)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-w.ctx.Done():
		return w.ctx.Err()
	}
}
