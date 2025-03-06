package circuitbreaker

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (cb *CircuitBreaker[T]) notifyCircuitBreakerCreation() {
	for _, s := range cb.sensors {
		cb.sensorLock.Lock()
		defer cb.sensorLock.Unlock()
		for _, ms := range s.GetMeters() {
			ms.IncrementCount(types.MetricCircuitBreakerCount)
		}
	}
}

func (cb *CircuitBreaker[T]) notifyError(time int64) {
	for _, s := range cb.sensors {
		cb.sensorLock.Lock()
		defer cb.sensorLock.Unlock()
		s.InvokeOnCircuitBreakerRecordError(cb.componentMetadata, time)
	}
}

func (cb *CircuitBreaker[T]) notifyTrip(time int64, resetTime int64) {
	for _, s := range cb.sensors {
		cb.sensorLock.Lock()
		defer cb.sensorLock.Unlock()
		s.InvokeOnCircuitBreakerTrip(cb.componentMetadata, time, resetTime)
	}
}

func (cb *CircuitBreaker[T]) notifyReset(time int64) {
	for _, s := range cb.sensors {
		cb.sensorLock.Lock()
		defer cb.sensorLock.Unlock()
		s.InvokeOnCircuitBreakerReset(cb.componentMetadata, time)
	}
}
