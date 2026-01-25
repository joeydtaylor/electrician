package conduit

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// GetComponentMetadata returns conduit metadata.
func (c *Conduit[T]) GetComponentMetadata() types.ComponentMetadata {
	c.configLock.Lock()
	metadata := c.componentMetadata
	c.configLock.Unlock()
	return metadata
}

// GetCircuitBreaker returns the configured circuit breaker.
func (c *Conduit[T]) GetCircuitBreaker() types.CircuitBreaker[T] {
	c.cbLock.Lock()
	cb := c.CircuitBreaker
	c.cbLock.Unlock()
	return cb
}

// GetInputChannel returns the conduit input channel.
func (c *Conduit[T]) GetInputChannel() chan T {
	c.configLock.Lock()
	ch := c.InputChan
	c.configLock.Unlock()
	return ch
}

// GetOutputChannel returns the conduit output channel.
func (c *Conduit[T]) GetOutputChannel() chan T {
	c.configLock.Lock()
	ch := c.OutputChan
	c.configLock.Unlock()
	return ch
}

// GetGenerators returns configured generators.
func (c *Conduit[T]) GetGenerators() []types.Generator[T] {
	return c.snapshotGenerators()
}

// IsStarted reports whether the conduit is running.
func (c *Conduit[T]) IsStarted() bool {
	return atomic.LoadInt32(&c.started) == 1
}
