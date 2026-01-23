package wire

import (
	"bytes"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// GetCircuitBreaker returns the configured circuit breaker.
func (w *Wire[T]) GetCircuitBreaker() types.CircuitBreaker[T] {
	w.cbLock.Lock()
	cb := w.CircuitBreaker
	w.cbLock.Unlock()
	return cb
}

// GetComponentMetadata returns the wire metadata.
func (w *Wire[T]) GetComponentMetadata() types.ComponentMetadata {
	return w.componentMetadata
}

// GetInputChannel returns the input channel.
func (w *Wire[T]) GetInputChannel() chan T {
	return w.inChan
}

// GetGenerators returns configured generators.
func (w *Wire[T]) GetGenerators() []types.Generator[T] {
	return w.generators
}

// GetOutputChannel returns the output channel.
func (w *Wire[T]) GetOutputChannel() chan T {
	return w.OutputChan
}

// GetOutputBuffer returns the output buffer.
func (w *Wire[T]) GetOutputBuffer() *bytes.Buffer {
	return w.OutputBuffer
}

// IsStarted reports whether the wire is running.
func (w *Wire[T]) IsStarted() bool {
	return atomic.LoadInt32(&w.started) == 1
}
