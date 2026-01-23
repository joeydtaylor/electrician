package wire

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (w *Wire[T]) notifyComplete() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire completed a cycle",
		"component", w.componentMetadata,
		"event", "Complete",
		"result", "SUCCESS",
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnComplete(w.componentMetadata)
	}
}

func (w *Wire[T]) notifyElementProcessed(elem T) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Successfully processed element",
		"component", w.componentMetadata,
		"event", "ElementProcessed",
		"result", "SUCCESS",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnElementProcessed(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyElementTransformError(elem T, err error) {
	w.NotifyLoggers(
		types.ErrorLevel,
		"Error occurred during element transformation",
		"component", w.componentMetadata,
		"event", "ElementProcessError",
		"result", "FAILURE",
		"element", elem,
		"error", err,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnError(w.componentMetadata, err, elem)
	}
}

func (w *Wire[T]) notifyCancel(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Element processing cancelled due to context cancellation",
		"component", w.componentMetadata,
		"event", "Submit",
		"result", "CANCELLED",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnCancel(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyNeutralWireSubmission(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Submitting element to neutral wire due to circuit breaker trip",
		"component", w.componentMetadata,
		"event", "NeutralWireSubmit",
		"result", "PENDING",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnCircuitBreakerNeutralWireSubmission(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyCircuitBreakerDropElement(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"No neutral wires available and circuit breaker tripped; dropping element",
		"component", w.componentMetadata,
		"event", "CircuitBreakerDropElement",
		"result", "DROPPED",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnCircuitBreakerDrop(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyStart() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire started",
		"component", w.componentMetadata,
		"event", "Start",
		"result", "SUCCESS",
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnStart(w.componentMetadata)
	}
}

func (w *Wire[T]) notifyStop() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire stopped",
		"component", w.componentMetadata,
		"event", "Stop",
		"result", "SUCCESS",
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnStop(w.componentMetadata)
	}
}

func (w *Wire[T]) notifySubmit(elem T) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Element submitted",
		"component", w.componentMetadata,
		"event", "Submit",
		"result", "SUCCESS",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnSubmit(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifySurgeProtectorSubmit(elem T) {
	sp := w.surgeProtector
	meta := types.ComponentMetadata{}
	if sp != nil {
		meta = sp.GetComponentMetadata()
	}

	w.NotifyLoggers(
		types.WarnLevel,
		"Surge protector tripped; submitting element for handling",
		"component", w.componentMetadata,
		"event", "SurgeProtectorSubmit",
		"result", "PENDING",
		"surge_protector", meta,
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnSurgeProtectorSubmit(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyRateLimit(elem T, nextAttempt string) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Hitting rate limit; submission deferred",
		"component", w.componentMetadata,
		"event", "Submit",
		"result", "PENDING",
		"nextAttempt", nextAttempt,
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnSurgeProtectorRateLimitExceeded(w.componentMetadata, elem)
	}
}
