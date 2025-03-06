package wire

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers sends a log message with the specified level, message, and key/value pairs
// to all registered loggers. It first checks (per logger) if logging is enabled for the provided level.
// Parameters:
//   - level: The log level for the message.
//   - msg: A descriptive message.
//   - keysAndValues: Variadic key/value pairs to include in the log entry.
func (w *Wire[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	if w.loggers != nil {
		for _, logger := range w.loggers {
			if logger == nil {
				continue
			}
			// Lock once per logger to reduce locking overhead.
			w.loggersLock.Lock()
			// If the logger implements the levelChecker interface and the level is not enabled, skip logging.
			type levelChecker interface {
				IsLevelEnabled(types.LogLevel) bool
			}
			if lc, ok := logger.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				w.loggersLock.Unlock()
				continue
			}

			switch level {
			case types.DebugLevel:
				logger.Debug(msg, keysAndValues...)
			case types.InfoLevel:
				logger.Info(msg, keysAndValues...)
			case types.WarnLevel:
				logger.Warn(msg, keysAndValues...)
			case types.ErrorLevel:
				logger.Error(msg, keysAndValues...)
			case types.DPanicLevel:
				logger.DPanic(msg, keysAndValues...)
			case types.PanicLevel:
				logger.Panic(msg, keysAndValues...)
			case types.FatalLevel:
				logger.Fatal(msg, keysAndValues...)
			}
			w.loggersLock.Unlock()
		}
	}
}

// notifyInsulatorFinalRetryFailure logs and notifies sensors when all insulator retry attempts have been exhausted.
// Parameters:
//   - currentElement: The element after the final retry attempt.
//   - originalElement: The element as originally received.
//   - currentErr: The error from the final retry attempt.
//   - originalErr: The error that initiated the retry process.
//   - currentAttempt: The final retry attempt number.
//   - maxThreshold: The maximum number of retry attempts allowed.
//   - interval: The duration between retry attempts.
func (w *Wire[T]) notifyInsulatorFinalRetryFailure(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.InfoLevel,
		"All insulator retries exhausted",
		"component", w.GetComponentMetadata(),
		"event", "insulatorAttempt",
		"result", "FAILURE",
		"totalAttempts", currentAttempt,
		"threshold", maxThreshold,
		"interval", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"currentElement", currentElement,
		"currentError", currentErr,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnInsulatorFailure(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnInsulatorFailure Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyInsulatorFinalRetryFailure",
			"result", "SUCCESS",
			"totalAttempts", currentAttempt,
			"threshold", maxThreshold,
			"interval", interval.Milliseconds(),
			"originalElement", originalElement,
			"originalErr", originalErr,
			"currentElement", currentElement,
			"currentError", currentErr,
			"sensor", sensor.GetComponentMetadata(),
		)
	}
}

// notifyComplete logs the completion of a processing cycle and notifies all sensors.
// It is called when the wire has finished processing the current batch of elements.
func (w *Wire[T]) notifyComplete() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire completed a cycle",
		"component", w.GetComponentMetadata(),
		"event", "notifyComplete",
		"result", "SUCCESS",
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnComplete(w.componentMetadata)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnComplete Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyComplete",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnComplete(w.componentMetadata)",
			"sensor", sensor.GetComponentMetadata(),
		)
	}
}

// notifyElementProcessed logs the successful processing of an element and notifies sensors.
// Parameters:
//   - elem: The element that was processed.
func (w *Wire[T]) notifyElementProcessed(elem T) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Successfully processed element",
		"component", w.GetComponentMetadata(),
		"event", "ElementProcessed",
		"result", "SUCCESS",
		"element", elem,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnElementProcessed(w.componentMetadata, elem)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnElementProcessed Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyElementProcessed",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnElementProcessed(w.componentMetadata, elem)",
			"element", elem,
			"target_component", sensor.GetComponentMetadata(),
		)
	}
}

// notifyInsulatorAttemptSuccess logs and notifies sensors when an insulator retry attempt successfully recovers an element.
// Parameters:
//   - currentElement: The element after a successful retry.
//   - originalElement: The element as originally received.
//   - currentErr: The error (if any) from the successful retry.
//   - originalErr: The error that initiated the retry process.
//   - currentAttempt: The current retry attempt number.
//   - maxThreshold: The maximum number of allowed retry attempts.
//   - interval: The duration between retry attempts.
func (w *Wire[T]) notifyInsulatorAttemptSuccess(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Successfully recovered element",
		"component", w.GetComponentMetadata(),
		"event", "insulatorAttempt",
		"result", "SUCCESS",
		"totalAttempts", currentAttempt,
		"threshold", maxThreshold,
		"interval", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"element", currentElement,
	)

	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnInsulatorSuccess(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnInsulatorSuccess Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyInsulatorAttemptSuccess",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnInsulatorSuccess(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)",
			"totalAttempts", currentAttempt,
			"threshold", maxThreshold,
			"interval", interval.Milliseconds(),
			"originalElement", originalElement,
			"originalErr", originalErr,
			"currentElement", currentElement,
			"currentError", currentErr,
			"sensor", sensor.GetComponentMetadata(),
		)
	}
}

// notifyInsulatorAttempt logs and notifies sensors about an insulator retry attempt.
// This is called before each retry attempt is made.
// Parameters:
//   - currentElement: The element state before the retry attempt.
//   - originalElement: The originally received element.
//   - currentErr: The error from the previous attempt that necessitated a retry.
//   - originalErr: The error that initiated the retry process.
//   - currentAttempt: The current retry attempt number.
//   - maxThreshold: The maximum number of retry attempts allowed.
//   - interval: The duration between retry attempts.
func (w *Wire[T]) notifyInsulatorAttempt(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Error during processing; attempting recovery...",
		"component", w.GetComponentMetadata(),
		"event", "insulatorAttempt",
		"result", "PENDING",
		"currentAttempt", currentAttempt,
		"threshold", maxThreshold,
		"interval", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"currentElement", currentElement,
		"currentErr", currentErr,
	)

	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnInsulatorAttempt(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnInsulatorAttempt Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyInsulatorAttempt",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnInsulatorAttempt(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)",
			"totalAttempts", currentAttempt,
			"threshold", maxThreshold,
			"interval", interval.Milliseconds(),
			"originalElement", originalElement,
			"originalErr", originalErr,
			"currentElement", currentElement,
			"currentError", currentErr,
			"sensor", sensor.GetComponentMetadata(),
		)
	}
}

// notifyElementTransformError logs an error that occurred during element transformation
// and notifies all sensors of the failure.
// Parameters:
//   - elem: The element that failed to process.
//   - err: The error encountered during processing.
func (w *Wire[T]) notifyElementTransformError(elem T, err error) {
	w.NotifyLoggers(
		types.ErrorLevel,
		"Error occurred during element transformation",
		"component", w.GetComponentMetadata(),
		"event", "ElementProcessError",
		"result", "FAILURE",
		"element", elem,
		"error", err,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnError(w.componentMetadata, err, elem)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnError Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyElementTransformError",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnError(w.componentMetadata, err, elem)",
			"element", elem,
			"target_component", sensor.GetComponentMetadata(),
		)
	}
}

// notifyCancel logs when element processing is cancelled due to context cancellation,
// notifies all sensors of the cancellation, and stops the wire.
// Parameters:
//   - elem: The element whose processing was cancelled.
func (w *Wire[T]) notifyCancel(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Element processing cancelled due to context cancellation",
		"component", w.GetComponentMetadata(),
		"event", "Submit",
		"result", "PENDING",
		"element", elem,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnCancel(w.componentMetadata, elem)
		w.Stop()
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnCancel Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyCancel",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnCancel(w.componentMetadata, elem)",
			"element", elem,
			"sensor", sensor.GetComponentMetadata(),
		)
	}
}

// notifyNeutralWireSubmission logs that an element is being submitted to a neutral (ground) wire
// due to a circuit breaker trip and notifies all sensors accordingly.
// Parameters:
//   - elem: The element being submitted to the neutral wire.
func (w *Wire[T]) notifyNeutralWireSubmission(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Submitting element to neutral wire due to circuit breaker trip",
		"component", w.GetComponentMetadata(),
		"event", "NeutralWireSubmit",
		"result", "PENDING",
		"element", elem,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnCircuitBreakerNeutralWireSubmission(w.componentMetadata, elem)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnCircuitBreakerNeutralWireSubmission Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyNeutralWireSubmission",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnCircuitBreakerNeutralWireSubmission(w.componentMetadata, elem)",
			"element", elem,
			"sensor", sensor.GetComponentMetadata(),
		)
	}
}

// notifyCircuitBreakerDropElement logs that an element is dropped due to a tripped circuit breaker
// with no available neutral wires, and notifies all sensors.
// Parameters:
//   - elem: The element that is being dropped.
func (w *Wire[T]) notifyCircuitBreakerDropElement(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"No neutral wires available and circuit breaker tripped; dropping element",
		"component", w.GetComponentMetadata(),
		"event", "CircuitBreakerDropElement",
		"result", "DROPPED",
		"call", "sensor.InvokeOnCircuitBreakerDrop(w.componentMetadata, elem)",
		"element", elem,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnCircuitBreakerDrop(w.componentMetadata, elem)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnCircuitBreakerDrop Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyCircuitBreakerDropElement",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnCircuitBreakerDrop(w.componentMetadata, elem)",
			"element", elem,
			"sensor", sensor.GetComponentMetadata(),
		)
	}
}

// notifyStart logs that the wire has started and notifies all sensors.
// It is called when the wire begins processing.
func (w *Wire[T]) notifyStart() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire started",
		"component", w.GetComponentMetadata(),
		"event", "Start",
		"result", "SUCCESS",
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnStart(w.componentMetadata)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnStart Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyStart",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnStart(w.componentMetadata)",
			"target_component", sensor.GetComponentMetadata(),
		)
	}
}

// notifyStop logs that the wire has stopped and notifies all sensors.
// It is called when the wire is shut down.
func (w *Wire[T]) notifyStop() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire stopped",
		"component", w.GetComponentMetadata(),
		"event", "Stop",
		"result", "SUCCESS",
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnStop(w.componentMetadata)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnStop Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyStop",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnStop(w.componentMetadata)",
			"target_component", sensor.GetComponentMetadata(),
		)
	}
}

// notifySubmit logs that an element has been successfully submitted to the wire
// and notifies all sensors about the submission.
// Parameters:
//   - elem: The element that was submitted.
func (w *Wire[T]) notifySubmit(elem T) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Element submitted",
		"component", w.GetComponentMetadata(),
		"event", "Submit",
		"result", "SUCCESS",
		"element", elem,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnSubmit(w.componentMetadata, elem)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnSubmit Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifySubmit",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnSubmit(w.componentMetadata, elem)",
			"element", elem,
			"target_component", sensor.GetComponentMetadata(),
		)
	}
}

// notifySurgeProtectorSubmit logs that an element is being submitted to the surge protector
// because it has tripped, and notifies all sensors.
// Parameters:
//   - elem: The element being submitted to the surge protector.
func (w *Wire[T]) notifySurgeProtectorSubmit(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Surge protector tripped; submitting element for handling",
		"component", w.GetComponentMetadata(),
		"event", "SurgeProtectorSubmit",
		"result", "PENDING",
		"surge_protector", w.surgeProtector.GetComponentMetadata(),
		"element", elem,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorSubmit(w.componentMetadata, elem)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnSurgeProtectorSubmit Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifySurgeProtectorSubmit",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnSurgeProtectorSubmit(w.componentMetadata, elem)",
			"element", elem,
			"target_component", sensor.GetComponentMetadata(),
		)
	}
}

// notifyRateLimit logs that an element submission has hit the rate limit and notifies all sensors.
// Parameters:
//   - elem: The element that triggered the rate limit.
//   - nextAttempt: A string indicating when the next submission attempt can be made.
func (w *Wire[T]) notifyRateLimit(elem T, nextAttempt string) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Hitting rate limit; submission deferred",
		"component", w.GetComponentMetadata(),
		"event", "Submit",
		"result", "PENDING",
		"nextAttempt", nextAttempt,
		"element", elem,
	)
	for _, sensor := range w.sensors {
		w.sensorLock.Lock()
		defer w.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorRateLimitExceeded(w.componentMetadata, elem)
		w.NotifyLoggers(
			types.DebugLevel,
			"InvokeOnSurgeProtectorRateLimitExceeded Callback invoked",
			"component", w.GetComponentMetadata(),
			"event", "notifyRateLimit",
			"result", "SUCCESS",
			"call", "sensor.InvokeOnSurgeProtectorRateLimitExceeded(w.componentMetadata)",
			"element", elem,
			"target_component", sensor.GetComponentMetadata(),
		)
	}
}
