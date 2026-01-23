package wire

import (
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (w *Wire[T]) notifyInsulatorFinalRetryFailure(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.InfoLevel,
		"All insulator retries exhausted",
		"component", w.componentMetadata,
		"event", "InsulatorAttempt",
		"result", "FAILURE",
		"totalAttempts", currentAttempt,
		"threshold", maxThreshold,
		"interval_ms", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"currentElement", currentElement,
		"currentErr", currentErr,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnInsulatorFailure(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

func (w *Wire[T]) notifyInsulatorAttemptSuccess(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Successfully recovered element",
		"component", w.componentMetadata,
		"event", "InsulatorAttempt",
		"result", "SUCCESS",
		"totalAttempts", currentAttempt,
		"threshold", maxThreshold,
		"interval_ms", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"element", currentElement,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnInsulatorSuccess(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

func (w *Wire[T]) notifyInsulatorAttempt(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Error during processing; attempting recovery",
		"component", w.componentMetadata,
		"event", "InsulatorAttempt",
		"result", "PENDING",
		"currentAttempt", currentAttempt,
		"threshold", maxThreshold,
		"interval_ms", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"currentElement", currentElement,
		"currentErr", currentErr,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	for _, s := range w.sensors {
		if s == nil {
			continue
		}
		s.InvokeOnInsulatorAttempt(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}
