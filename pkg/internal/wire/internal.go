// Package wire contains internal operations and utilities for the Wire component within the Electrician framework.
// The Wire component is a fundamental part of the data processing pipeline, responsible for receiving, transforming,
// and forwarding data. This file includes lower-level functions and methods that handle data processing, error management,
// and interactions with related components such as circuit breakers and sensors.
//
// The provided methods include:
// - Encoding and processing individual elements received through the wire's input channels.
// - Handling errors that occur during data processing and reporting them through designated error channels.
// - Managing the lifecycle and operational state of the wire, including starting and stopping processing based on context signals.
//
// This internal API focuses on robust error handling, efficient data processing, and thorough logging. It ensures that
// the Wire can operate reliably and transparently in a concurrent processing environment, making extensive use of
// synchronization primitives to manage state safely across multiple goroutines.
//
// Highlights include:
//   - Methods for connecting various components like circuit breakers, which enhance fault tolerance,
//     and sensors, which provide monitoring capabilities.
//   - Utilities for encoding data and safely handling potential encoding errors.
//   - Comprehensive logging throughout processing steps to aid in debugging and operational monitoring.
//   - Detailed error handling and reporting mechanisms that integrate seamlessly with the wire's operational logic,
//     ensuring that all errors are accounted for and handled appropriately.
//
// The design and implementation of these methods prioritize high-performance and reliable data processing within
// the broader Electrician framework, aiming to provide a solid foundation for building complex data-intensive applications.
package wire

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// isClosedSafe retrieves the w.isClosed flag under lock, preventing data races.
func (w *Wire[T]) isClosedSafe() bool {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	return w.isClosed
}

// setClosedSafe sets w.isClosed = true under lock. (Called during shutdown.)
func (w *Wire[T]) setClosedSafe() {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	w.isClosed = true
}

// attemptRecovery attempts to recover from a processing error by applying the configured insulator function.
// It retries up to w.retryThreshold times, logging each attempt and returning either the successful element or
// the final error if all retries fail.
func (w *Wire[T]) attemptRecovery(elem T, originalElem T, originalErr error) (T, error) {
	var retryErr error
	for attempt := 1; attempt <= w.retryThreshold; attempt++ {
		retryElem, retryErr := w.insulatorFunc(w.ctx, elem, originalErr)
		if retryErr == nil {
			w.notifyInsulatorAttemptSuccess(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
			return retryElem, nil
		}
		// If final attempt fails, record error in CircuitBreaker and return the error.
		if attempt == w.retryThreshold {
			if w.CircuitBreaker != nil {
				w.CircuitBreaker.RecordError()
			}
			w.notifyInsulatorFinalRetryFailure(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
			return retryElem, fmt.Errorf("retry threshold of %d reached with error: %v", w.retryThreshold, retryErr)
		}
		w.notifyInsulatorAttempt(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
		elem = retryElem
		time.Sleep(w.retryInterval)
	}
	return elem, retryErr
}

// startCircuitBreakerTicker continuously monitors the circuit breaker status in a separate goroutine.
func (w *Wire[T]) startCircuitBreakerTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.cbLock.Lock()
			allowed := true
			if w.CircuitBreaker != nil {
				allowed = w.CircuitBreaker.Allow()
			}
			w.cbLock.Unlock()

			// Only send if wire is not closed.
			w.closeLock.Lock()
			if !w.isClosed {
				select {
				case w.controlChan <- allowed:
				default:
				}
			}
			w.closeLock.Unlock()
		}
	}
}

// encodeElement uses w.encoder to encode the element into w.OutputBuffer, handling errors with handleError.
func (w *Wire[T]) encodeElement(elem T) {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()
	if w.encoder != nil {
		if encodeErr := w.encoder.Encode(w.OutputBuffer, elem); encodeErr != nil {
			w.handleError(elem, encodeErr)
		}
	}
}

// handleCircuitBreakerTrip is called when the circuit breaker disallows processing (breaker is open).
// The element is diverted to any neutral (ground) wires; if none are available, the element is effectively dropped.
func (w *Wire[T]) handleCircuitBreakerTrip(ctx context.Context, elem T) error {
	groundWires := w.CircuitBreaker.GetNeutralWires()
	if len(groundWires) > 0 {
		for _, gw := range groundWires {
			if err := gw.Submit(ctx, elem); err != nil {
				continue // Try other ground wires if one fails.
			}
			w.notifyNeutralWireSubmission(elem)
		}
		return nil
	}
	w.notifyCircuitBreakerDropElement(elem)
	return nil
}

// handleError sends an error + element to w.errorChan if the wire isn't closed.
// We lock around isClosed check and the send to prevent "send on closed channel."
func (w *Wire[T]) handleError(elem T, err error) {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	if w.isClosed || w.errorChan == nil {
		return
	}
	w.errorChan <- types.ElementError[T]{Err: err, Elem: elem}
}

// handleErrorElements listens for errors from w.errorChan in a loop, calling w.notifyElementTransformError for each.
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

// handleProcessingError tries to run w.insulatorFunc if configured, else records via circuit breaker.
func (w *Wire[T]) handleProcessingError(elem T, originalElem T, originalErr error) (T, error) {
	if w.insulatorFunc != nil && (w.CircuitBreaker == nil || w.CircuitBreaker.Allow()) {
		return w.attemptRecovery(elem, originalElem, originalErr)
	}
	if w.CircuitBreaker != nil {
		w.CircuitBreaker.RecordError()
	}
	return elem, originalErr
}

// transformElement runs all transformations in sequence; if any fail, we handle or recover from the error.
func (w *Wire[T]) transformElement(elem T) (T, error) {
	var err error
	originalElem := elem
	for _, transform := range w.transformations {
		elem, err = transform(elem)
		if err != nil {
			return w.handleProcessingError(elem, originalElem, err)
		}
	}
	if w.sensors != nil {
		w.notifyElementProcessed(elem)
	}
	return elem, nil
}

// transformElements spawns w.maxBufferSize goroutines, each running processChannel to handle data from w.inChan.
func (w *Wire[T]) transformElements() {
	for i := 0; i < int(w.maxBufferSize); i++ {
		w.wg.Add(1)
		go w.processChannel()
	}
}

// processChannel reads elements from w.inChan, spawns transformations, and respects wire/context shutdowns.
func (w *Wire[T]) processChannel() {
	defer w.wg.Done()
	for {
		select {
		case elem, ok := <-w.inChan:
			if !ok {
				return
			}
			w.concurrencySem <- struct{}{}
			go func(e T) {
				defer func() { <-w.concurrencySem }()
				// If closed mid-flight, skip.
				if w.isClosedSafe() {
					return
				}
				w.processElementOrDivert(e)
			}(elem)
		case <-w.ctx.Done():
			return
		}
	}
}

// processElementOrDivert calls transformElement, checks circuit breaker, possibly diverts, or handles errors.
func (w *Wire[T]) processElementOrDivert(elem T) {
	if w.isClosedSafe() {
		return
	}
	processedElem, err := w.transformElement(elem)
	if w.shouldDivertError(err) {
		w.divertToGroundWires(elem)
		return
	}
	if err != nil {
		w.handleError(elem, err)
		return
	}
	w.submitProcessedElement(processedElem)
}

// shouldDivertError returns true if an error occurred and the circuit breaker is open (disallowing processing).
func (w *Wire[T]) shouldDivertError(err error) bool {
	return err != nil && w.CircuitBreaker != nil && !w.CircuitBreaker.Allow()
}

// divertToGroundWires sends the given element to any "neutral" wires if the breaker is open; else handleError.
func (w *Wire[T]) divertToGroundWires(elem T) {
	if len(w.CircuitBreaker.GetNeutralWires()) > 0 {
		for _, gw := range w.CircuitBreaker.GetNeutralWires() {
			w.notifyNeutralWireSubmission(elem)
			gw.Submit(w.ctx, elem)
		}
	} else {
		w.handleError(elem, fmt.Errorf("circuit breaker tripped but no ground wires to handle"))
	}
}

// submitProcessedElement encodes the element, then tries to send it on w.OutputChan if not closed.
// We lock around w.isClosed check and the send to avoid panics.
func (w *Wire[T]) submitProcessedElement(processedElem T) {
	w.encodeElement(processedElem)
	if w.OutputChan == nil {
		return
	}
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	if w.isClosed {
		return
	}
	select {
	case w.OutputChan <- processedElem:
		// success
	case <-w.ctx.Done():
		w.notifyCancel(processedElem)
	}
}

// processResisterElements runs w.processItems with a ticker, respecting any rate-limiting from SurgeProtector.
func (w *Wire[T]) processResisterElements(ctx context.Context) {
	var ticker *time.Ticker
	if w.surgeProtector == nil || !w.surgeProtector.IsBeingRateLimited() {
		ticker = time.NewTicker(1 * time.Second)
	} else {
		_, fillrate, _, _ := w.surgeProtector.GetRateLimit()
		if fillrate > 0 {
			ticker = time.NewTicker(fillrate)
		} else {
			// If fillrate <= 0, no point in ticking. Use 1s fallback.
			ticker = time.NewTicker(1 * time.Second)
		}
	}
	defer ticker.Stop()
	w.processItems(ctx, ticker)
}

// processItems repeatedly tries to dequeue items from surgeProtector (if any), then Submit them to wire.
func (w *Wire[T]) processItems(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if w.surgeProtector != nil && w.surgeProtector.TryTake() {
				if item, err := w.surgeProtector.Dequeue(); err == nil {
					w.Submit(ctx, item.Data)
				} else {
					// If no more items remain, possibly stop.
					if w.shouldStopProcessing() {
						return
					}
				}
			}
		}
	}
}

// shouldStopProcessing returns true if surgeProtector's queue is empty.
func (w *Wire[T]) shouldStopProcessing() bool {
	return w.surgeProtector.GetResisterQueue() == 0
}

// submitNormally attempts to send an item to w.inChan, guarded by closeLock to avoid sends after wire closes.
func (w *Wire[T]) submitNormally(ctx context.Context, elem T) error {
	w.closeLock.Lock()
	defer w.closeLock.Unlock()
	if w.isClosed {
		return fmt.Errorf("cannot submit: input channel is closed")
	}
	select {
	case w.inChan <- elem:
		w.notifySubmit(elem)
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
