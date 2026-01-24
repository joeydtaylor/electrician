package circuitbreaker

import (
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetComponentMetadata updates the circuit breaker name and id.
func (cb *CircuitBreaker[T]) SetComponentMetadata(name string, id string) {
	cb.configLock.Lock()
	cb.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: cb.componentMetadata.Type}
	cb.configLock.Unlock()
}

// SetDebouncePeriod configures the minimum spacing between recorded errors.
func (cb *CircuitBreaker[T]) SetDebouncePeriod(seconds int) {
	if seconds <= 0 {
		atomic.StoreInt64(&cb.debounceNanos, 0)
		return
	}
	atomic.StoreInt64(&cb.debounceNanos, int64(time.Second)*int64(seconds))
}
