package wire

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// startWorkers starts worker goroutines.
func (w *Wire[T]) startWorkers() {
	workers := w.maxConcurrency
	if workers <= 0 {
		workers = 1
	}

	for i := 0; i < workers; i++ {
		w.wg.Add(1)

		if w.fastPathEnabled {
			if w.transformerFactory != nil {
				tf := w.transformerFactory()
				if tf == nil {
					panic("wire: transformerFactory returned nil transformer")
				}
				go w.runFastWorker(tf)
				continue
			}

			go w.runFastWorker(w.fastTransform)
			continue
		}

		go w.runWorker()
	}
}

// runWorker is the standard worker loop.
func (w *Wire[T]) runWorker() {
	defer w.wg.Done()

	in := w.inChan
	done := w.ctx.Done()

	telemetryEnabled := atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0

	for {
		select {
		case <-done:
			return
		case elem, ok := <-in:
			if !ok {
				return
			}
			w.handleElement(elem, telemetryEnabled)
		}
	}
}

func (w *Wire[T]) handleElement(elem T, telemetryEnabled bool) {
	if w.ctx.Err() != nil {
		return
	}

	processedElem, err := w.transformElement(elem, telemetryEnabled)

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

// runFastWorker runs a minimal worker loop with a preselected transformer.
func (w *Wire[T]) runFastWorker(transform types.Transformer[T]) {
	defer w.wg.Done()

	in := w.inChan
	out := w.OutputChan
	done := w.ctx.Done()

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
				w.handleError(elem, err)
				continue
			}

			select {
			case out <- processed:
			case <-done:
				w.notifyCancel(processed)
				return
			}
		}
	}
}
