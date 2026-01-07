package wire

import (
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// IMPORTANT:
// For zero-alloc hot path when no observers are attached, Wire[T] should track counts:
//   loggerCount int32
//   sensorCount int32
// and ConnectLogger / ConnectSensor should increment them for non-nil attachments.
//
// This file gates all logging/sensor work behind those counters so we don't even build
// []interface{} key/value args when nobody is listening.

// hasLoggers returns true if any logger is attached (atomic, no locks, no alloc).
func (w *Wire[T]) hasLoggers() bool {
	return atomic.LoadInt32(&w.loggerCount) != 0
}

// hasSensors returns true if any sensor is attached (atomic, no locks, no alloc).
func (w *Wire[T]) hasSensors() bool {
	return atomic.LoadInt32(&w.sensorCount) != 0
}

// snapshotLoggers returns a stable snapshot of the logger slice.
// Never hold w.loggersLock while invoking logger methods.
func (w *Wire[T]) snapshotLoggers() []types.Logger {
	// Fast-path: no loggers => no locks/allocs.
	if !w.hasLoggers() {
		return nil
	}

	w.loggersLock.Lock()
	defer w.loggersLock.Unlock()

	if len(w.loggers) == 0 {
		return nil
	}
	out := make([]types.Logger, len(w.loggers))
	copy(out, w.loggers)
	return out
}

// snapshotSensors returns a stable snapshot of the sensor slice.
// Never hold w.sensorLock while invoking sensor callbacks.
func (w *Wire[T]) snapshotSensors() []types.Sensor[T] {
	// Fast-path: no sensors => no locks/allocs.
	if !w.hasSensors() {
		return nil
	}

	w.sensorLock.Lock()
	defer w.sensorLock.Unlock()

	if len(w.sensors) == 0 {
		return nil
	}
	out := make([]types.Sensor[T], len(w.sensors))
	copy(out, w.sensors)
	return out
}

// NotifyLoggers sends a log message with the specified level, message, and key/value pairs
// to all registered loggers. It first checks (per logger) if logging is enabled for the provided level.
//
// NOTE: Even if this function early-returns, the allocations for variadic interface{} args can already
// happen at the call site. Therefore, callers on hot paths MUST gate calls with w.hasLoggers().
func (w *Wire[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := w.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	type levelChecker interface {
		IsLevelEnabled(types.LogLevel) bool
	}

	for _, logger := range loggers {
		if logger == nil {
			continue
		}
		if lc, ok := logger.(levelChecker); ok && !lc.IsLevelEnabled(level) {
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
	}
}

// notifyInsulatorFinalRetryFailure logs and notifies sensors when all insulator retry attempts have been exhausted.
func (w *Wire[T]) notifyInsulatorFinalRetryFailure(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.InfoLevel,
			"All insulator retries exhausted",
			"component", w.componentMetadata,
			"event", "insulatorAttempt",
			"result", "FAILURE",
			"totalAttempts", currentAttempt,
			"threshold", maxThreshold,
			"interval_ms", interval.Milliseconds(),
			"originalElement", originalElement,
			"originalErr", originalErr,
			"currentElement", currentElement,
			"currentError", currentErr,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnInsulatorFailure(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

// notifyComplete logs completion and notifies sensors.
func (w *Wire[T]) notifyComplete() {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.InfoLevel,
			"Wire completed a cycle",
			"component", w.componentMetadata,
			"event", "notifyComplete",
			"result", "SUCCESS",
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnComplete(w.componentMetadata)
	}
}

// notifyElementProcessed logs and notifies sensors.
func (w *Wire[T]) notifyElementProcessed(elem T) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.InfoLevel,
			"Successfully processed element",
			"component", w.componentMetadata,
			"event", "ElementProcessed",
			"result", "SUCCESS",
			"element", elem,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnElementProcessed(w.componentMetadata, elem)
	}
}

// notifyInsulatorAttemptSuccess logs and notifies sensors when an insulator retry attempt successfully recovers an element.
func (w *Wire[T]) notifyInsulatorAttemptSuccess(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.InfoLevel,
			"Successfully recovered element",
			"component", w.componentMetadata,
			"event", "insulatorAttempt",
			"result", "SUCCESS",
			"totalAttempts", currentAttempt,
			"threshold", maxThreshold,
			"interval_ms", interval.Milliseconds(),
			"originalElement", originalElement,
			"originalErr", originalErr,
			"element", currentElement,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnInsulatorSuccess(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

// notifyInsulatorAttempt logs and notifies sensors about an insulator retry attempt.
func (w *Wire[T]) notifyInsulatorAttempt(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.WarnLevel,
			"Error during processing; attempting recovery...",
			"component", w.componentMetadata,
			"event", "insulatorAttempt",
			"result", "PENDING",
			"currentAttempt", currentAttempt,
			"threshold", maxThreshold,
			"interval_ms", interval.Milliseconds(),
			"originalElement", originalElement,
			"originalErr", originalErr,
			"currentElement", currentElement,
			"currentErr", currentErr,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnInsulatorAttempt(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

// notifyElementTransformError logs an error and notifies sensors.
func (w *Wire[T]) notifyElementTransformError(elem T, err error) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.ErrorLevel,
			"Error occurred during element transformation",
			"component", w.componentMetadata,
			"event", "ElementProcessError",
			"result", "FAILURE",
			"element", elem,
			"error", err,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnError(w.componentMetadata, err, elem)
	}
}

// notifyCancel logs cancellation and notifies sensors.
// CRITICAL: do NOT call w.Stop() here. Stop() waits on goroutines that may be executing this callback.
func (w *Wire[T]) notifyCancel(elem T) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.WarnLevel,
			"Element processing cancelled due to context cancellation",
			"component", w.componentMetadata,
			"event", "Submit",
			"result", "CANCELLED",
			"element", elem,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnCancel(w.componentMetadata, elem)
	}
}

// notifyNeutralWireSubmission logs and notifies sensors.
func (w *Wire[T]) notifyNeutralWireSubmission(elem T) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.WarnLevel,
			"Submitting element to neutral wire due to circuit breaker trip",
			"component", w.componentMetadata,
			"event", "NeutralWireSubmit",
			"result", "PENDING",
			"element", elem,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnCircuitBreakerNeutralWireSubmission(w.componentMetadata, elem)
	}
}

// notifyCircuitBreakerDropElement logs and notifies sensors.
func (w *Wire[T]) notifyCircuitBreakerDropElement(elem T) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.WarnLevel,
			"No neutral wires available and circuit breaker tripped; dropping element",
			"component", w.componentMetadata,
			"event", "CircuitBreakerDropElement",
			"result", "DROPPED",
			"element", elem,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnCircuitBreakerDrop(w.componentMetadata, elem)
	}
}

// notifyStart logs and notifies sensors.
func (w *Wire[T]) notifyStart() {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.InfoLevel,
			"Wire started",
			"component", w.componentMetadata,
			"event", "Start",
			"result", "SUCCESS",
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnStart(w.componentMetadata)
	}
}

// notifyStop logs and notifies sensors.
func (w *Wire[T]) notifyStop() {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.InfoLevel,
			"Wire stopped",
			"component", w.componentMetadata,
			"event", "Stop",
			"result", "SUCCESS",
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnStop(w.componentMetadata)
	}
}

// notifySubmit logs and notifies sensors.
func (w *Wire[T]) notifySubmit(elem T) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.InfoLevel,
			"Element submitted",
			"component", w.componentMetadata,
			"event", "Submit",
			"result", "SUCCESS",
			"element", elem,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSubmit(w.componentMetadata, elem)
	}
}

// notifySurgeProtectorSubmit logs and notifies sensors.
func (w *Wire[T]) notifySurgeProtectorSubmit(elem T) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.WarnLevel,
			"Surge protector tripped; submitting element for handling",
			"component", w.componentMetadata,
			"event", "SurgeProtectorSubmit",
			"result", "PENDING",
			"surge_protector", w.surgeProtector.GetComponentMetadata(),
			"element", elem,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorSubmit(w.componentMetadata, elem)
	}
}

// notifyRateLimit logs and notifies sensors.
func (w *Wire[T]) notifyRateLimit(elem T, nextAttempt string) {
	if w.hasLoggers() {
		w.NotifyLoggers(
			types.WarnLevel,
			"Hitting rate limit; submission deferred",
			"component", w.componentMetadata,
			"event", "Submit",
			"result", "PENDING",
			"nextAttempt", nextAttempt,
			"element", elem,
		)
	}

	if !w.hasSensors() {
		return
	}
	for _, sensor := range w.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorRateLimitExceeded(w.componentMetadata, elem)
	}
}
