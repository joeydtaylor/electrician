package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// RegisterOnCircuitBreakerTrip registers a callback to be invoked when a circuit breaker trip event occurs.
func (s *Sensor[T]) RegisterOnCircuitBreakerTrip(callback ...func(cmd types.ComponentMetadata, time int64, nextReset int64)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnCircuitBreakerTrip = append(s.OnCircuitBreakerTrip, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerTrip, target: %v => Registered OnCircuitBreakerTrip function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerTrip invokes registered callbacks for the circuit breaker trip event.
func (s *Sensor[T]) InvokeOnCircuitBreakerTrip(cmd types.ComponentMetadata, time int64, nextReset int64) {
	for _, callback := range s.OnCircuitBreakerTrip {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerTrip, target: %v => Invoked OnCircuitBreakerTrip function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, time, nextReset)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerReset registers a callback to be invoked when a circuit breaker reset event occurs.
func (s *Sensor[T]) RegisterOnCircuitBreakerReset(callback ...func(cmd types.ComponentMetadata, time int64)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnCircuitBreakerReset = append(s.OnCircuitBreakerReset, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerReset, target: %v => Registered OnCircuitBreakerReset function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerReset invokes registered callbacks for the circuit breaker reset event.
func (s *Sensor[T]) InvokeOnCircuitBreakerReset(cmd types.ComponentMetadata, time int64) {
	for _, callback := range s.OnCircuitBreakerReset {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerReset, target: %v => Invoked OnCircuitBreakerReset function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, time)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerNeutralWireSubmission registers a callback for ground wire submission events related to circuit breakers.
func (s *Sensor[T]) RegisterOnCircuitBreakerNeutralWireSubmission(callback ...func(cmd types.ComponentMetadata, elem T)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnCircuitBreakerNeutralWireSubmission = append(s.OnCircuitBreakerNeutralWireSubmission, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerNeutralWireSubmission, target: %v => Registered OnCircuitBreakerNeutralWireSubmission function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerNeutralWireSubmission invokes callbacks when a ground wire submission event occurs in a circuit breaker context.
func (s *Sensor[T]) InvokeOnCircuitBreakerNeutralWireSubmission(cmd types.ComponentMetadata, elem T) {
	for _, callback := range s.OnCircuitBreakerNeutralWireSubmission {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerNeutralWireSubmission, target: %v => Invoked OnCircuitBreakerNeutralWireSubmission function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, elem)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerRecordError registers a callback to be invoked when a circuit breaker records an error.
func (s *Sensor[T]) RegisterOnCircuitBreakerRecordError(callback ...func(cmd types.ComponentMetadata, time int64)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnCircuitBreakerRecordError = append(s.OnCircuitBreakerRecordError, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerRecordError, target: %v => Registered OnCircuitBreakerRecordError function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerRecordError invokes registered callbacks for the circuit breaker record error event.
func (s *Sensor[T]) InvokeOnCircuitBreakerRecordError(cmd types.ComponentMetadata, time int64) {
	for _, callback := range s.OnCircuitBreakerRecordError {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerRecordError, target: %v => Invoked OnCircuitBreakerRecordError function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, time)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerAllow registers a callback to be invoked when a circuit breaker allows an operation.
func (s *Sensor[T]) RegisterOnCircuitBreakerAllow(callback ...func(cmd types.ComponentMetadata)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnCircuitBreakerAllow = append(s.OnCircuitBreakerAllow, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerAllow, target: %v => Registered OnCircuitBreakerAllow function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerAllow invokes registered callbacks for the circuit breaker allow event.
func (s *Sensor[T]) InvokeOnCircuitBreakerAllow(cmd types.ComponentMetadata) {
	for _, callback := range s.OnCircuitBreakerAllow {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerAllow, target: %v => Invoked OnCircuitBreakerAllow function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerDrop registers a callback to be invoked when a circuit breaker drops an element.
func (s *Sensor[T]) RegisterOnCircuitBreakerDrop(callback ...func(cmd types.ComponentMetadata, elem T)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnCircuitBreakerDrop = append(s.OnCircuitBreakerDrop, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerDrop, target: %v => Registered OnCircuitBreakerDrop function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerDrop invokes registered callbacks for the circuit breaker drop event, passing the dropped element to each callback.
func (s *Sensor[T]) InvokeOnCircuitBreakerDrop(cmd types.ComponentMetadata, elem T) {
	for _, callback := range s.OnCircuitBreakerDrop {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerDrop, target: %v => Invoked OnCircuitBreakerDrop function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, elem)
		s.callbackLock.Unlock()
	}
}

func (s *Sensor[T]) decorateCircuitBreakerCallbacks() []types.Option[types.Sensor[T]] {
	var cbOptions []types.Option[types.Sensor[T]]
	cbOptions = append(cbOptions,
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
	)
	return cbOptions
}
