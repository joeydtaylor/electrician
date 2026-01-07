// Package wire contains internal operations and utilities for the Wire component within the Electrician framework.
//
// This file is the hot-path execution engine:
//   - bounded worker pool (size = maxConcurrency)
//   - buffered channels provide backpressure (maxBufferSize)
//   - zero goroutine-per-element spawning in steady state
//
// Design goals:
//   - predictable concurrency and memory
//   - minimal scheduler/GC pressure
//   - cancellation-safe shutdown (workers exit on ctx.Done or channel close)
//   - keep existing external behavior/API unchanged
package wire

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// attemptRecovery attempts to recover from a processing error by applying the configured insulator function.
// It retries up to retryThreshold. If the wire is cancelled, it aborts early.
func (w *Wire[T]) attemptRecovery(elem T, originalElem T, originalErr error) (T, error) {
	ins := w.insulatorFunc
	if ins == nil {
		return elem, originalErr
	}

	threshold := w.retryThreshold
	interval := w.retryInterval

	var retryErr error
	for attempt := 1; attempt <= threshold; attempt++ {
		// Respect cancellation; don't sleep/loop pointlessly.
		if w.ctx.Err() != nil {
			return elem, originalErr
		}

		retryElem, err := ins(w.ctx, elem, originalErr)
		retryErr = err
		if retryErr == nil {
			w.notifyInsulatorAttemptSuccess(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
			return retryElem, nil
		}

		if attempt == threshold {
			if w.CircuitBreaker != nil {
				w.CircuitBreaker.RecordError()
			}
			w.notifyInsulatorFinalRetryFailure(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
			return retryElem, fmt.Errorf("retry threshold of %d reached with error: %v", threshold, retryErr)
		}

		w.notifyInsulatorAttempt(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
		elem = retryElem

		if interval > 0 {
			// Cancellation-aware sleep.
			select {
			case <-time.After(interval):
			case <-w.ctx.Done():
				return elem, originalErr
			}
		}
	}

	return elem, retryErr
}

// startCircuitBreakerTicker monitors the circuit breaker and pushes the current allow state onto controlChan.
func (w *Wire[T]) startCircuitBreakerTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.cbLock.Lock()
			cb := w.CircuitBreaker
			w.cbLock.Unlock()

			allowed := true
			if cb != nil {
				allowed = cb.Allow()
			}

			select {
			case w.controlChan <- allowed:
			default:
			}
		}
	}
}

// encodeElement encodes a processed element using the configured encoder.
// Fast-path: if no encoder, do nothing (avoid lock).
func (w *Wire[T]) encodeElement(elem T) {
	enc := w.encoder
	if enc == nil {
		return
	}

	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()

	if err := enc.Encode(w.OutputBuffer, elem); err != nil {
		w.handleError(elem, err)
	}
}

// handleCircuitBreakerTrip submits elem to neutral wires if available; otherwise drops.
func (w *Wire[T]) handleCircuitBreakerTrip(ctx context.Context, elem T) error {
	cb := w.CircuitBreaker
	if cb == nil {
		return nil
	}

	groundWires := cb.GetNeutralWires()
	if len(groundWires) == 0 {
		w.notifyCircuitBreakerDropElement(elem)
		return nil
	}

	for _, gw := range groundWires {
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

// handleError reports a processing error.
// Cancellation-safe: don't block forever on a full error channel during shutdown.
func (w *Wire[T]) handleError(elem T, err error) {
	if w.errorChan == nil {
		return
	}
	select {
	case w.errorChan <- types.ElementError[T]{Err: err, Elem: elem}:
	case <-w.ctx.Done():
	}
}

// handleErrorElements drains errorChan and notifies sensors/loggers.
// Exits on ctx cancellation or when the channel is closed.
func (w *Wire[T]) handleErrorElements() {
	defer w.wg.Done()
	for {
		select {
		case e, ok := <-w.errorChan:
			if !ok {
				return
			}
			w.notifyElementTransformError(e.Elem, e.Err)
		case <-w.ctx.Done():
			return
		}
	}
}

// handleProcessingError manages errors from transformElement: insulator first (if configured), else record error.
func (w *Wire[T]) handleProcessingError(elem T, originalElem T, originalErr error) (T, error) {
	if w.insulatorFunc != nil && (w.CircuitBreaker == nil || w.CircuitBreaker.Allow()) {
		return w.attemptRecovery(elem, originalElem, originalErr)
	}
	if w.CircuitBreaker != nil {
		w.CircuitBreaker.RecordError()
	}
	return elem, originalErr
}

// transformElement applies all transformations in order. On success, sensors are notified.
func (w *Wire[T]) transformElement(elem T) (T, error) {
	var err error
	originalElem := elem

	// Localize slice ref for slightly cheaper iteration.
	transforms := w.transformations
	for i := 0; i < len(transforms); i++ {
		elem, err = transforms[i](elem)
		if err != nil {
			return w.handleProcessingError(elem, originalElem, err)
		}
	}

	// Avoid per-element notify overhead when no sensors are attached.
	if atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0 {
		w.notifyElementProcessed(elem)
	}

	return elem, nil
}

func (w *Wire[T]) processChannelFast() {
	defer w.wg.Done()

	in := w.inChan
	out := w.OutputChan
	done := w.ctx.Done()
	transform := w.fastTransform

	for {
		select {
		case <-done:
			return
		case elem, ok := <-in:
			if !ok {
				return
			}

			processed, err := transform(elem)
			if err != nil {
				// No observers in fast path; keep it minimal.
				// If you want to preserve error channel semantics, keep this line:
				// w.handleError(elem, err)
				continue
			}

			select {
			case out <- processed:
			case <-done:
				return
			}
		}
	}
}

// transformElements starts a bounded worker pool.
// maxConcurrency is the worker count; maxBufferSize is queue depth (channel capacity) only.
func (w *Wire[T]) transformElements() {
	workers := w.maxConcurrency
	if workers <= 0 {
		workers = 1
	}

	for i := 0; i < workers; i++ {
		w.wg.Add(1)
		if w.fastPathEnabled {
			go w.processChannelFast()
		} else {
			go w.processChannel()
		}
	}
}

// processChannel is a worker loop.
// No semaphore; no goroutine-per-element. Shutdown via ctx cancellation or input channel close.
func (w *Wire[T]) processChannel() {
	defer w.wg.Done()

	for {
		select {
		case <-w.ctx.Done():
			return
		case elem, ok := <-w.inChan:
			if !ok {
				return
			}
			w.processElementOrDivert(elem)
		}
	}
}

// processElementOrDivert transforms elem, handles errors, and emits to output.
func (w *Wire[T]) processElementOrDivert(elem T) {
	// Cheap cancellation check to avoid wasting compute after shutdown starts.
	if w.ctx.Err() != nil {
		return
	}

	processedElem, err := w.transformElement(elem)

	// If error exists and circuit breaker is open, divert original element.
	if w.shouldDivertError(err) {
		w.divertToGroundWires(elem)
		return
	}

	// If processing failed, report error.
	if err != nil {
		w.handleError(elem, err)
		return
	}

	// Success: encode and emit.
	w.submitProcessedElement(processedElem)
}

// shouldDivertError returns true if err exists and the circuit breaker is present and not allowing processing.
func (w *Wire[T]) shouldDivertError(err error) bool {
	return err != nil && w.CircuitBreaker != nil && !w.CircuitBreaker.Allow()
}

// divertToGroundWires sends elem to neutral wires, or emits an error if none exist.
func (w *Wire[T]) divertToGroundWires(elem T) {
	cb := w.CircuitBreaker
	if cb == nil {
		w.handleError(elem, fmt.Errorf("circuit breaker diversion requested but CircuitBreaker is nil"))
		return
	}

	neutral := cb.GetNeutralWires()
	if len(neutral) == 0 {
		w.handleError(elem, fmt.Errorf("circuit breaker tripped but no ground wires to handle"))
		return
	}

	for _, gw := range neutral {
		if gw == nil {
			continue
		}
		w.notifyNeutralWireSubmission(elem)
		_ = gw.Submit(w.ctx, elem)
	}
}

// submitProcessedElement encodes elem and emits to OutputChan.
// IMPORTANT: No closeLock here. With the worker-pool model, Stop() closes channels only after w.wg.Wait(),
// so sends cannot race with close as long as ONLY workers send to OutputChan (true for this wire).
//
// Cancellation path: notifyCancel() calls Stop() in your public file; calling that inline from a worker would deadlock
// (Stop waits on w.wg which includes this worker). So we invoke notifyCancel asynchronously.
func (w *Wire[T]) submitProcessedElement(processedElem T) {
	if w.encoder != nil {
		w.encodeElement(processedElem)
	}

	if w.OutputChan == nil {
		return
	}

	select {
	case w.OutputChan <- processedElem:
		return
	case <-w.ctx.Done():
		w.notifyCancel(processedElem)
		return
	}
}

// processResisterElements manages the processing of elements under rate limiting.
func (w *Wire[T]) processResisterElements(ctx context.Context) {
	if w.surgeProtector == nil || !w.surgeProtector.IsBeingRateLimited() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		w.processItems(ctx, ticker)
		return
	}

	_, fillrate, _, _ := w.surgeProtector.GetRateLimit()
	if fillrate > 0 {
		ticker := time.NewTicker(fillrate)
		defer ticker.Stop()
		w.processItems(ctx, ticker)
	}
}

// processItems periodically checks for items from the surge protector and submits them for processing.
func (w *Wire[T]) processItems(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.surgeProtector != nil && w.surgeProtector.TryTake() {
				item, err := w.surgeProtector.Dequeue()
				if err == nil {
					_ = w.Submit(ctx, item.Data)
					continue
				}
				if w.shouldStopProcessing() {
					return
				}
			}
		}
	}
}

// shouldStopProcessing returns true if the surge protector queue is empty.
func (w *Wire[T]) shouldStopProcessing() bool {
	return w.surgeProtector.GetResisterQueue() == 0
}

// submitNormally enqueues elem onto inChan, respecting context cancellation.
func (w *Wire[T]) submitNormally(ctx context.Context, elem T) error {
	select {
	case w.inChan <- elem:
		// Only notify if anyone is listening.
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

func (w *Wire[T]) computeFastPathEnabled() bool {
	// Fast path only when "extras" are completely off and there's exactly 1 transform.
	return len(w.transformations) == 1 &&
		w.CircuitBreaker == nil &&
		w.surgeProtector == nil &&
		w.insulatorFunc == nil &&
		w.encoder == nil &&
		atomic.LoadInt32(&w.loggerCount) == 0 &&
		atomic.LoadInt32(&w.sensorCount) == 0
}
