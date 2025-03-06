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

// attemptRecovery attempts to recover from a processing error by applying the configured insulator function.
// It retries the operation up to the retry threshold, logging each attempt.
// Parameters:
//   - elem: The element currently being processed.
//   - originalElem: The original element before any recovery attempts.
//   - originalErr: The error that triggered the recovery attempt.
//
// Returns:
//   - T: The recovered element (or the last attempted version if recovery fails).
//   - error: nil if recovery is successful, otherwise the final error encountered.
func (w *Wire[T]) attemptRecovery(elem T, originalElem T, originalErr error) (T, error) {
	var retryErr error
	for attempt := 1; attempt <= w.retryThreshold; attempt++ {
		retryElem, retryErr := w.insulatorFunc(w.ctx, elem, originalErr)
		if retryErr == nil {
			w.notifyInsulatorAttemptSuccess(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
			return retryElem, nil
		}
		if attempt == w.retryThreshold {
			if w.CircuitBreaker != nil {
				w.CircuitBreaker.RecordError()
			}
			w.notifyInsulatorFinalRetryFailure(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
			return retryElem, fmt.Errorf("retry threshold of %d reached with error: %v", w.retryThreshold, retryErr)
		}
		w.notifyInsulatorAttempt(retryElem, originalElem, retryErr, originalErr, attempt, w.retryThreshold, w.retryInterval)
		elem = retryElem // Update elem for the next retry attempt
		time.Sleep(w.retryInterval)
	}
	return elem, retryErr
}

// startCircuitBreakerTicker continuously monitors the status of the circuit breaker.
// It runs as a separate goroutine and periodically sends the current allowed state to the control channel.
func (w *Wire[T]) startCircuitBreakerTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.cbLock.Lock() // Lock before checking the CircuitBreaker status
			allowed := true
			if w.CircuitBreaker != nil {
				allowed = w.CircuitBreaker.Allow()
			}
			w.cbLock.Unlock() // Unlock after checking

			select {
			case w.controlChan <- allowed:
			default:
			}
		}
	}
}

// encodeElement safely encodes a processed element using the configured encoder.
// The encoded data is written to the output buffer, and any encoding error is handled.
func (w *Wire[T]) encodeElement(elem T) {
	w.bufferMutex.Lock()
	defer w.bufferMutex.Unlock()
	if w.encoder != nil {
		/* 		w.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, event: Submit, element: %v, => Encoding element.", w.GetComponentMetadata(), elem) */
		if encodeErr := w.encoder.Encode(w.OutputBuffer, elem); encodeErr != nil {
			w.handleError(elem, encodeErr)
		}
	}
}

// handleCircuitBreakerTrip manages the submission of an element when the circuit breaker has tripped.
// It attempts to submit the element to each available neutral (ground) wire.
// If no ground wires are available, the element is dropped.
func (w *Wire[T]) handleCircuitBreakerTrip(ctx context.Context, elem T) error {
	groundWires := w.CircuitBreaker.GetNeutralWires()
	if len(groundWires) > 0 {
		for _, gw := range groundWires {
			if err := gw.Submit(ctx, elem); err != nil {
				/* 				w.NotifyLoggers(types.WarnLevel, "component: %s, level: WARN, result: FAILURE, event: Submit, element: %v, error: %v target: %s => Error submitting to ground wire!", w.GetComponentMetadata(), elem, err, gw.GetComponentMetadata()) */
				continue // Log but do not return; try other ground wires.
			}
			w.notifyNeutralWireSubmission(elem)
			/* 			w.NotifyLoggers(types.InfoLevel, "component: %s, level: INFO, result: SUCCESS, event: Submit, element: %v, target: %s => Submitted to ground wire after CircuitBreaker trip.", w.GetComponentMetadata(), elem, gw.GetComponentMetadata()) */
		}
		return nil // Successfully handled by ground wires.
	}
	w.notifyCircuitBreakerDropElement(elem)
	/* 	w.NotifyLoggers(types.WarnLevel, "component: %s, level: WARN, result: DROPPED, event: Submit, element: %v => No ground wires available, dropping element!", w.GetComponentMetadata(), elem) */
	return nil
}

// handleError sends an error along with its associated element to the error channel, if configured.
func (w *Wire[T]) handleError(elem T, err error) {
	if w.errorChan != nil {
		w.errorChan <- types.ElementError[T]{Err: err, Elem: elem}
	}
}

// handleErrorElements continuously listens for errors on the error channel and processes them.
// It logs each error via the notifyElementTransformError function.
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

// handleProcessingError manages errors that occur during element processing.
// If an insulator function is configured and the circuit breaker allows it,
// the method attempts recovery; otherwise, it records the error via the circuit breaker.
func (w *Wire[T]) handleProcessingError(elem T, originalElem T, originalErr error) (T, error) {
	if w.insulatorFunc != nil && (w.CircuitBreaker == nil || w.CircuitBreaker.Allow()) {
		return w.attemptRecovery(elem, originalElem, originalErr)
	}
	if w.CircuitBreaker != nil {
		w.CircuitBreaker.RecordError()
	}
	return elem, originalErr
}

// transformElement applies all transformation functions in sequence to the given element.
// If any transformation fails, it handles the error via handleProcessingError.
// On successful transformation, it notifies sensors of the processed element.
// Returns the transformed element and any error encountered.
func (w *Wire[T]) transformElement(elem T) (T, error) {
	var err error
	originalElem := elem // Preserve the original element for logging purposes

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

// transformElements launches multiple goroutines (equal to maxBufferSize) to process incoming elements concurrently.
func (w *Wire[T]) transformElements() {
	for i := 0; i < int(w.maxBufferSize); i++ {
		w.wg.Add(1)
		go w.processChannel()
	}
}

// processChannel continuously reads elements from the input channel and processes them.
// It uses a semaphore to limit the number of concurrent processing routines.
func (w *Wire[T]) processChannel() {
	defer w.wg.Done()
	for {
		select {

		case elem, ok := <-w.inChan:
			if !ok {
				return // Exit if the channel is closed.
			}

			// Acquire a slot from the semaphore before processing the element.
			w.concurrencySem <- struct{}{}

			go func(element T) {
				defer func() { <-w.concurrencySem }() // Release the semaphore slot.
				if !w.isClosed {
					w.processElementOrDivert(element)
				} else {
					return
				}
			}(elem)

		case <-w.ctx.Done():
			return // Terminate if the context is done.
		}
	}
}

// processElementOrDivert processes a single element by applying transformations.
// If an error occurs and the circuit breaker is tripped, the element is diverted to ground wires.
// Otherwise, the processed element is submitted for further processing.
func (w *Wire[T]) processElementOrDivert(elem T) {
	if w.isClosed {
		return
	} else if !w.isClosed {
		processedElem, err := w.transformElement(elem)

		// If an error occurs and the circuit breaker is tripped, divert the element.
		if w.shouldDivertError(err) {
			w.divertToGroundWires(elem)
			return
		}

		// If processing fails, handle the error.
		if err != nil {
			w.handleError(elem, err)
			return
		}

		// On successful processing, encode and submit the element.
		w.submitProcessedElement(processedElem)
	}
}

// shouldDivertError determines whether an error should trigger diversion to ground wires.
// It returns true if an error exists, a circuit breaker is configured, and the circuit breaker is not allowing processing.
func (w *Wire[T]) shouldDivertError(err error) bool {
	return err != nil && w.CircuitBreaker != nil && !w.CircuitBreaker.Allow()
}

// divertToGroundWires diverts an element to available neutral (ground) wires when the circuit breaker is open.
// If no ground wires are available, it handles the error accordingly.
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

// submitProcessedElement encodes a processed element and submits it to the output channel.
// It respects the shutdown state and cancellation signals.
func (w *Wire[T]) submitProcessedElement(processedElem T) {
	w.encodeElement(processedElem)
	if w.OutputChan != nil {
		w.closeLock.Lock()
		defer w.closeLock.Unlock()
		if w.isClosed {
			return
		}
		select {
		case w.OutputChan <- processedElem:
		case <-w.ctx.Done():
			w.notifyCancel(processedElem)
			return
		}
	}
}

// processResisterElements manages the processing of elements under rate limiting.
// It sets up a ticker based on the surge protector's configuration and delegates processing to processItems.
func (w *Wire[T]) processResisterElements(ctx context.Context) {
	if w.surgeProtector == nil || !w.surgeProtector.IsBeingRateLimited() {
		// Use a default ticker for normal operation if rate limiting is not enabled.
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		w.processItems(ctx, ticker)
	} else {
		_, fillrate, _, _ := w.surgeProtector.GetRateLimit()
		if fillrate > 0 {
			ticker := time.NewTicker(fillrate)
			defer ticker.Stop()
			w.processItems(ctx, ticker)
		}
	}
}

// processItems periodically checks for available items from the surge protector's queue using a ticker.
// If an item is available, it is submitted for processing.
// Processing stops if the context is cancelled or if the surge protector's queue is empty.
func (w *Wire[T]) processItems(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return // Exit on context cancellation.
		case <-ticker.C:
			if w.surgeProtector != nil && w.surgeProtector.TryTake() {
				if item, err := w.surgeProtector.Dequeue(); err == nil {
					// Only submit if items are available.
					w.Submit(ctx, item.Data)
				} else {
					// Stop processing if no items remain.
					if w.shouldStopProcessing() {
						return
					}
				}
			}
		}
	}
}

// shouldStopProcessing determines whether processing should stop by checking the surge protector's queue.
// Returns true if the queue is empty.
func (w *Wire[T]) shouldStopProcessing() bool {
	return w.surgeProtector.GetResisterQueue() == 0
}

// submitNormally attempts to submit an element to the wire's input channel.
// It notifies successful submission and returns an error if the context is cancelled.
func (w *Wire[T]) submitNormally(ctx context.Context, elem T) error {
	select {
	case w.inChan <- elem:
		w.notifySubmit(elem)
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}
