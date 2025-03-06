package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *Sensor[T]) decorateSurgeProtectorCallbacks() []types.Option[types.Sensor[T]] {
	var surgeProtectorOptions []types.Option[types.Sensor[T]]
	surgeProtectorOptions = append(surgeProtectorOptions,
		WithSurgeProtectorTripFunc[T](func(c types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricComponentRunningCount)
			s.incrementMeterCounters(types.MetricSurgeTripCount)
			s.incrementMeterCounters(types.MetricSurgeProtectorCurrentTripCount)
		}),
		WithSurgeProtectorResetFunc[T](func(c types.ComponentMetadata) {
			s.decrementMeterCounters(types.MetricComponentRunningCount)
			s.incrementMeterCounters(types.MetricSurgeResetCount)
			s.decrementMeterCounters(types.MetricSurgeProtectorCurrentTripCount)
		}),
		WithSurgeProtectorBackupFailureFunc[T](func(c types.ComponentMetadata, err error) {
			s.incrementMeterCounters(types.MetricSurgeBackupFailureCount)
		}),
		WithSurgeProtectorBackupWireSubmissionFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricSurgeProtectorBackupWireSubmissionCount)
		}),
		WithSurgeProtectorDroppedSubmissionFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricSurgeProtectorDropCount)
		}),
		WithSurgeProtectorConnectResisterFunc[T](func(c, r types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricResisterConnectedCount)
		}),
	)
	return surgeProtectorOptions
}

// RegisterOnSurgeProtectorTrip registers callbacks to be invoked when the surge protector trips.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector trip, each accepting no parameters.
func (m *Sensor[T]) RegisterOnSurgeProtectorTrip(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSurgeProtectorTrip = append(m.OnSurgeProtectorTrip, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorTrip, target: %v => Registered OnSurgeProtectorTrip function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorTrip triggers callbacks registered for surge protector trip events.
func (m *Sensor[T]) InvokeOnSurgeProtectorTrip(c types.ComponentMetadata) {
	for _, callback := range m.OnSurgeProtectorTrip {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorTrip, target: %v => Invoked OnSurgeProtectorTrip function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorReset registers callbacks to be invoked when the surge protector resets.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector reset, each accepting no parameters.
func (m *Sensor[T]) RegisterOnSurgeProtectorReset(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSurgeProtectorReset = append(m.OnSurgeProtectorReset, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorReset, target: %v => Registered OnSurgeProtectorReset function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorReset triggers callbacks registered for surge protector reset events.
func (m *Sensor[T]) InvokeOnSurgeProtectorReset(c types.ComponentMetadata) {
	for _, callback := range m.OnSurgeProtectorReset {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorReset, target: %v => Invoked OnSurgeProtectorReset function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorBackupFailure registers callbacks to be invoked when the surge protector encounters backup failures.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector backup failure, each accepting an error parameter.
func (m *Sensor[T]) RegisterOnSurgeProtectorBackupFailure(callback ...func(c types.ComponentMetadata, err error)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSurgeProtectorBackupFailure = append(m.OnSurgeProtectorBackupFailure, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorBackupFailure, target: %v => Registered OnSurgeProtectorBackupFailure function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorBackupFailure triggers callbacks registered for surge protector backup failure events, passing the error to each callback.
//
// Parameters:
//   - err: The error encountered during the surge protector backup process.
func (m *Sensor[T]) InvokeOnSurgeProtectorBackupFailure(c types.ComponentMetadata, err error) {
	for _, callback := range m.OnSurgeProtectorBackupFailure {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorBackupFailure, error: %v, target: %v => Invoked OnSurgeProtectorBackupFailure function", m.componentMetadata, err, callback)
		m.callbackLock.Lock()
		callback(c, err)
		m.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorRateLimitExceeded registers callbacks to be invoked when the surge protector's rate limit is exceeded.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector rate limit exceeded, each accepting no parameters.
func (m *Sensor[T]) RegisterOnSurgeProtectorRateLimitExceeded(callback ...func(c types.ComponentMetadata, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSurgeProtectorRateLimitExceeded = append(m.OnSurgeProtectorRateLimitExceeded, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorRateLimitExceeded, target: %v => Registered OnSurgeProtectorRateLimitExceeded function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorRateLimitExceeded triggers callbacks registered for surge protector rate limit exceeded events.
func (m *Sensor[T]) InvokeOnSurgeProtectorRateLimitExceeded(c types.ComponentMetadata, elem T) {
	for _, callback := range m.OnSurgeProtectorRateLimitExceeded {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorRateLimitExceeded, target: %v => Invoked OnSurgeProtectorRateLimitExceeded function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c, elem)
		m.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerDrop registers a callback to be invoked when a circuit breaker drops an element.
func (s *Sensor[T]) RegisterOnSurgeProtectorSubmit(callback ...func(cmd types.ComponentMetadata, elem T)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnSurgeProtectorSubmit = append(s.OnSurgeProtectorSubmit, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerDrop, target: %v => Registered OnCircuitBreakerDrop function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerDrop invokes registered callbacks for the circuit breaker drop event, passing the dropped element to each callback.
func (s *Sensor[T]) InvokeOnSurgeProtectorSubmit(cmd types.ComponentMetadata, elem T) {
	for _, callback := range s.OnSurgeProtectorSubmit {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerDrop, target: %v => Invoked OnCircuitBreakerDrop function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, elem)
		s.callbackLock.Unlock()
	}
}
