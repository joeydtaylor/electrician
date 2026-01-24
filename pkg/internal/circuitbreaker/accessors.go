package circuitbreaker

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns the circuit breaker metadata.
func (cb *CircuitBreaker[T]) GetComponentMetadata() types.ComponentMetadata {
	cb.configLock.Lock()
	metadata := cb.componentMetadata
	cb.configLock.Unlock()
	return metadata
}

// GetNeutralWires returns the configured neutral wires.
func (cb *CircuitBreaker[T]) GetNeutralWires() []types.Wire[T] {
	cb.configLock.Lock()
	wires := append([]types.Wire[T](nil), cb.neutralWires...)
	cb.configLock.Unlock()
	return wires
}

// NotifyOnReset returns a channel signaled when the breaker resets.
func (cb *CircuitBreaker[T]) NotifyOnReset() <-chan struct{} {
	return cb.resetNotifyChan
}
