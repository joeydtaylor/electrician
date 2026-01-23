package wire

import (
	"bytes"
	"context"
	"sync"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Start initializes the wire and launches worker routines.
func (w *Wire[T]) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&w.started, 0, 1) {
		return nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			atomic.StoreInt32(&w.started, 0)
			return err
		}
	}

	w.notifyStart()

	w.wg.Add(1)
	go w.handleErrorElements()

	if sp := w.surgeProtector; sp != nil && sp.IsResisterConnected() {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.runResisterLoop(w.ctx)
		}()
	}

	if w.transformerFactory != nil {
		if len(w.transformations) != 0 {
			panic("wire: invalid config: transformerFactory set but transformations also present")
		}
	} else {
		if len(w.transformations) == 0 {
			// Preserve pass-through behavior when no transforms are configured.
			w.transformations = []types.Transformer[T]{
				func(v T) (T, error) { return v, nil },
			}
		}
	}

	w.fastPathEnabled = w.canUseFastPath()
	if w.fastPathEnabled && w.transformerFactory == nil {
		w.fastTransform = w.transformations[0]
	}

	w.startWorkers()

	for _, g := range w.generators {
		if g != nil && !g.IsStarted() {
			g.Start(w.ctx)
		}
	}

	return nil
}

// Stop cancels the wire and waits for workers to exit.
func (w *Wire[T]) Stop() error {
	if !atomic.CompareAndSwapInt32(&w.started, 1, 0) {
		return nil
	}

	w.notifyStop()

	w.terminateOnce.Do(func() {
		w.closeLock.Lock()
		w.isClosed = true
		w.closeLock.Unlock()

		w.cancel()
		w.wg.Wait()

		select {
		case <-w.completeSignal:
		default:
			w.closeOutputChanOnce.Do(func() { close(w.OutputChan) })
			w.closeErrorChanOnce.Do(func() { close(w.errorChan) })
		}
	})

	return nil
}

// Restart stops the wire, reinitializes channels, and starts again.
func (w *Wire[T]) Restart(ctx context.Context) error {
	_ = w.Stop()

	w.ctx, w.cancel = context.WithCancel(ctx)

	w.terminateOnce = sync.Once{}
	w.closeOutputChanOnce = sync.Once{}
	w.closeInputChanOnce = sync.Once{}
	w.closeErrorChanOnce = sync.Once{}

	w.closeLock.Lock()
	w.isClosed = false
	w.closeLock.Unlock()

	w.SetInputChannel(make(chan T, w.maxBufferSize))
	w.SetOutputChannel(make(chan T, w.maxBufferSize))
	w.SetErrorChannel(make(chan types.ElementError[T], w.maxBufferSize))
	w.OutputBuffer = &bytes.Buffer{}

	if w.CircuitBreaker != nil {
		go w.startCircuitBreakerTicker()
	}

	return w.Start(w.ctx)
}
