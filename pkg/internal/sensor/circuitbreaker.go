package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// RegisterOnCircuitBreakerTrip registers callbacks for circuit breaker trip events.
func (s *Sensor[T]) RegisterOnCircuitBreakerTrip(callback ...func(types.ComponentMetadata, int64, int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnCircuitBreakerTrip = append(s.OnCircuitBreakerTrip, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnCircuitBreakerTrip invokes callbacks for circuit breaker trips.
func (s *Sensor[T]) InvokeOnCircuitBreakerTrip(cmd types.ComponentMetadata, time int64, nextReset int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnCircuitBreakerTrip) {
		if cb == nil {
			continue
		}
		cb(cmd, time, nextReset)
	}
}

// RegisterOnCircuitBreakerReset registers callbacks for circuit breaker reset events.
func (s *Sensor[T]) RegisterOnCircuitBreakerReset(callback ...func(types.ComponentMetadata, int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnCircuitBreakerReset = append(s.OnCircuitBreakerReset, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnCircuitBreakerReset invokes callbacks for circuit breaker resets.
func (s *Sensor[T]) InvokeOnCircuitBreakerReset(cmd types.ComponentMetadata, time int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnCircuitBreakerReset) {
		if cb == nil {
			continue
		}
		cb(cmd, time)
	}
}

// RegisterOnCircuitBreakerNeutralWireSubmission registers callbacks for diverted submissions.
func (s *Sensor[T]) RegisterOnCircuitBreakerNeutralWireSubmission(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnCircuitBreakerNeutralWireSubmission = append(s.OnCircuitBreakerNeutralWireSubmission, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnCircuitBreakerNeutralWireSubmission invokes callbacks for diverted submissions.
func (s *Sensor[T]) InvokeOnCircuitBreakerNeutralWireSubmission(cmd types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnCircuitBreakerNeutralWireSubmission) {
		if cb == nil {
			continue
		}
		cb(cmd, elem)
	}
}

// RegisterOnCircuitBreakerRecordError registers callbacks for recorded errors.
func (s *Sensor[T]) RegisterOnCircuitBreakerRecordError(callback ...func(types.ComponentMetadata, int64)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnCircuitBreakerRecordError = append(s.OnCircuitBreakerRecordError, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnCircuitBreakerRecordError invokes callbacks for recorded errors.
func (s *Sensor[T]) InvokeOnCircuitBreakerRecordError(cmd types.ComponentMetadata, time int64) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnCircuitBreakerRecordError) {
		if cb == nil {
			continue
		}
		cb(cmd, time)
	}
}

// RegisterOnCircuitBreakerAllow registers callbacks for allow events.
func (s *Sensor[T]) RegisterOnCircuitBreakerAllow(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnCircuitBreakerAllow = append(s.OnCircuitBreakerAllow, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnCircuitBreakerAllow invokes callbacks for allow events.
func (s *Sensor[T]) InvokeOnCircuitBreakerAllow(cmd types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnCircuitBreakerAllow) {
		if cb == nil {
			continue
		}
		cb(cmd)
	}
}

// RegisterOnCircuitBreakerDrop registers callbacks for dropped elements.
func (s *Sensor[T]) RegisterOnCircuitBreakerDrop(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnCircuitBreakerDrop = append(s.OnCircuitBreakerDrop, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnCircuitBreakerDrop invokes callbacks for dropped elements.
func (s *Sensor[T]) InvokeOnCircuitBreakerDrop(cmd types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnCircuitBreakerDrop) {
		if cb == nil {
			continue
		}
		cb(cmd, elem)
	}
}

func (s *Sensor[T]) decorateCircuitBreakerCallbacks() []types.Option[types.Sensor[T]] {
	return []types.Option[types.Sensor[T]]{
		WithCircuitBreakerDropFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricCircuitBreakerDroppedElementCount)
		}),
		WithCircuitBreakerNeutralWireSubmissionFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricCircuitBreakerNeutralWireSubmissionCount)
		}),
		WithCircuitBreakerTripFunc[T](func(c types.ComponentMetadata, time int64, resetTime int64) {
			s.setMetricTimestampValue(types.MetricCircuitBreakerLastTripTime, time)
			s.setMetricTimestampValue(types.MetricCircuitBreakerNextResetTime, resetTime)
			s.incrementMeterCounters(types.MetricComponentRunningCount)
			s.incrementMeterCounters(types.MetricCircuitBreakerCurrentTripCount)
			s.incrementMeterCounters(types.MetricCircuitBreakerTripCount)
		}),
		WithCircuitBreakerResetFunc[T](func(c types.ComponentMetadata, time int64) {
			s.setMetricTimestampValue(types.MetricCircuitBreakerLastResetTime, time)
			s.decrementMeterCounters(types.MetricCircuitBreakerCurrentTripCount)
			s.incrementMeterCounters(types.MetricCircuitBreakerResetCount)
		}),
	}
}
