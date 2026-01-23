package wire

import (
	"fmt"
	"sync/atomic"
)

func (w *Wire[T]) encodeElement(elem T) {
	enc := w.encoder
	if enc == nil {
		return
	}

	w.bufferMutex.Lock()
	err := enc.Encode(w.OutputBuffer, elem)
	w.bufferMutex.Unlock()

	if err != nil {
		w.handleError(elem, err)
	}
}

func (w *Wire[T]) shouldDivertError(err error) bool {
	return err != nil && w.CircuitBreaker != nil && !w.CircuitBreaker.Allow()
}

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

func (w *Wire[T]) submitProcessedElement(processedElem T) {
	if w.encoder != nil {
		w.encodeElement(processedElem)
	}

	if w.OutputChan == nil {
		return
	}

	select {
	case w.OutputChan <- processedElem:
	case <-w.ctx.Done():
		w.notifyCancel(processedElem)
	}
}

// canUseFastPath reports whether the fast path can be used.
func (w *Wire[T]) canUseFastPath() bool {
	if w.CircuitBreaker != nil ||
		w.surgeProtector != nil ||
		w.insulatorFunc != nil ||
		w.encoder != nil ||
		atomic.LoadInt32(&w.loggerCount) != 0 ||
		atomic.LoadInt32(&w.sensorCount) != 0 {
		return false
	}

	if w.transformerFactory != nil {
		return len(w.transformations) == 0
	}
	return len(w.transformations) == 1
}
