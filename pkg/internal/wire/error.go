package wire

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (w *Wire[T]) handleError(elem T, err error) {
	ch := w.errorChan
	if ch == nil {
		return
	}

	select {
	case ch <- types.ElementError[T]{Err: err, Elem: elem}:
	default:
	}
}

// handleErrorElements drains error notifications and forwards to telemetry when enabled.
func (w *Wire[T]) handleErrorElements() {
	defer w.wg.Done()

	if !w.hasLoggers() && !w.hasSensors() {
		ch := w.errorChan
		if ch == nil {
			return
		}
		for {
			select {
			case <-w.ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
		}
	}

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
