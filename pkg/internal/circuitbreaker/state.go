package circuitbreaker

import (
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Allow returns true if the circuit breaker allows work to proceed.
func (cb *CircuitBreaker[T]) Allow() bool {
	now := time.Now()

	cb.stateLock.Lock()
	allowed := cb.allowed
	shouldReset := false
	if !allowed && (cb.timeWindow <= 0 || now.Sub(cb.lastTripped) >= cb.timeWindow) {
		cb.allowed = true
		cb.errorCount = 0
		allowed = true
		shouldReset = true
	}
	cb.stateLock.Unlock()

	if shouldReset {
		cb.notifyReset(now.UnixNano())
		cb.signalReset()
		cb.NotifyLoggers(types.InfoLevel, "Circuit breaker reset", "component", cb.snapshotMetadata(), "auto", true)
	}

	if allowed {
		cb.notifyAllow()
	}

	return allowed
}

// RecordError records a failure and trips the breaker when the threshold is exceeded.
func (cb *CircuitBreaker[T]) RecordError() {
	now := time.Now()
	debounce := time.Duration(atomic.LoadInt64(&cb.debounceNanos))

	cb.stateLock.Lock()
	if debounce > 0 && !cb.lastErrorTime.IsZero() && now.Sub(cb.lastErrorTime) < debounce {
		cb.stateLock.Unlock()
		return
	}

	cb.lastErrorTime = now
	cb.errorCount++
	errorCount := cb.errorCount
	shouldTrip := cb.allowed && errorCount >= cb.errorThreshold
	nextReset := time.Time{}
	if shouldTrip {
		cb.allowed = false
		cb.lastTripped = now
		nextReset = now.Add(cb.timeWindow)
	}
	cb.stateLock.Unlock()

	metadata := cb.snapshotMetadata()

	cb.notifyRecordError(now.UnixNano())
	cb.NotifyLoggers(types.DebugLevel, "Circuit breaker recorded error", "component", metadata, "errorCount", errorCount, "errorThreshold", cb.errorThreshold)

	if shouldTrip {
		cb.notifyTrip(now.UnixNano(), nextReset.UnixNano())
		cb.NotifyLoggers(types.WarnLevel, "Circuit breaker tripped", "component", metadata, "errorThreshold", cb.errorThreshold, "nextReset", nextReset)
	}
}

// Reset moves the breaker back to the closed state.
func (cb *CircuitBreaker[T]) Reset() {
	now := time.Now()

	cb.stateLock.Lock()
	if cb.allowed {
		cb.stateLock.Unlock()
		return
	}
	cb.allowed = true
	cb.errorCount = 0
	cb.stateLock.Unlock()

	cb.notifyReset(now.UnixNano())
	cb.signalReset()
	cb.NotifyLoggers(types.InfoLevel, "Circuit breaker reset", "component", cb.snapshotMetadata(), "auto", false)
}

// Trip forces the breaker into the open state.
func (cb *CircuitBreaker[T]) Trip() {
	now := time.Now()

	cb.stateLock.Lock()
	if !cb.allowed {
		cb.stateLock.Unlock()
		return
	}
	cb.allowed = false
	cb.lastTripped = now
	nextReset := now.Add(cb.timeWindow)
	cb.stateLock.Unlock()

	cb.notifyTrip(now.UnixNano(), nextReset.UnixNano())
	cb.NotifyLoggers(types.WarnLevel, "Circuit breaker tripped", "component", cb.snapshotMetadata(), "errorThreshold", cb.errorThreshold, "nextReset", nextReset)
}
