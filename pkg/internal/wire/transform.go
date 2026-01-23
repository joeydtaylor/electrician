package wire

import (
	"fmt"
	"time"
)

// attemptRecovery applies the insulator retry policy when configured.
func (w *Wire[T]) attemptRecovery(elem T, originalElem T, originalErr error) (T, error) {
	ins := w.insulatorFunc
	if ins == nil {
		return elem, originalErr
	}

	threshold := w.retryThreshold
	if threshold <= 0 {
		return elem, originalErr
	}
	interval := w.retryInterval

	var timer *time.Timer
	if interval > 0 && threshold > 1 {
		timer = time.NewTimer(interval)
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		defer timer.Stop()
	}

	var retryErr error
	for attempt := 1; attempt <= threshold; attempt++ {
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
			if cb := w.CircuitBreaker; cb != nil {
				cb.RecordError()
			}
			w.notifyInsulatorFinalRetryFailure(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
			return retryElem, fmt.Errorf("retry threshold of %d reached with error: %v", threshold, retryErr)
		}

		w.notifyInsulatorAttempt(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
		elem = retryElem

		if interval > 0 {
			if timer != nil {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(interval)
				select {
				case <-timer.C:
				case <-w.ctx.Done():
					return elem, originalErr
				}
			}
		}
	}

	return elem, retryErr
}

func (w *Wire[T]) handleProcessingError(elem T, originalElem T, originalErr error) (T, error) {
	if w.insulatorFunc != nil {
		cb := w.CircuitBreaker
		if cb == nil || cb.Allow() {
			return w.attemptRecovery(elem, originalElem, originalErr)
		}
	}

	if cb := w.CircuitBreaker; cb != nil {
		cb.RecordError()
	}
	return elem, originalErr
}

// transformElement runs the configured transform chain and emits telemetry when enabled.
func (w *Wire[T]) transformElement(elem T, telemetryEnabled bool) (T, error) {
	originalElem := elem
	transforms := w.transformations

	switch len(transforms) {
	case 0:
	case 1:
		var err error
		elem, err = transforms[0](elem)
		if err != nil {
			return w.handleProcessingError(elem, originalElem, err)
		}
	default:
		for i := 0; i < len(transforms); i++ {
			var err error
			elem, err = transforms[i](elem)
			if err != nil {
				return w.handleProcessingError(elem, originalElem, err)
			}
		}
	}

	if telemetryEnabled {
		w.notifyElementProcessed(elem)
	}
	return elem, nil
}
