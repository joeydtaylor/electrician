package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// InvokeOnHTTPClientRequestComplete triggers callbacks registered for HTTP client request complete events.
func (m *Sensor[T]) InvokeOnHTTPClientRequestComplete(c types.ComponentMetadata) {
	for _, callback := range m.OnHTTPClientRequestComplete {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnHTTPClientRequestComplete, target: %v => Invoked OnHTTPClientRequestComplete function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

// InvokeOnHTTPClientRequestStart triggers callbacks registered for HTTP client request start events.
func (m *Sensor[T]) InvokeOnHTTPClientRequestStart(c types.ComponentMetadata) {
	for _, callback := range m.OnHTTPClientRequestStart {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnHTTPClientRequestStart, target: %v => Invoked OnHTTPClientRequestStart function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

// InvokeOnHTTPClientResponseReceived triggers callbacks registered for HTTP client response received events.
func (m *Sensor[T]) InvokeOnHTTPClientResponseReceived(c types.ComponentMetadata) {
	for _, callback := range m.OnHTTPClientResponseReceived {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnHTTPClientResponseReceived, target: %v => Invoked OnHTTPClientResponseReceived function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

// InvokeOnHTTPClientError triggers callbacks registered for HTTP client error events, passing the error to each callback.
//
// Parameters:
//   - err: The error encountered during the HTTP request.
func (m *Sensor[T]) InvokeOnHTTPClientError(c types.ComponentMetadata, err error) {
	for _, callback := range m.OnHTTPClientError {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnHTTPClientError, error: %v, target: %v => Invoked OnHTTPClientError function", m.componentMetadata, err, callback)
		m.callbackLock.Lock()
		callback(c, err)
		m.callbackLock.Unlock()
	}
}

// RegisterOnHTTPClientRequestStart registers callbacks to be invoked when an HTTP client request starts.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon HTTP client request start, each accepting no parameters.
func (m *Sensor[T]) RegisterOnHTTPClientRequestStart(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnHTTPClientRequestStart = append(m.OnHTTPClientRequestStart, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnHTTPClientRequestStart, target: %v => Registered OnHTTPClientRequestStart function", m.componentMetadata, callback)
	}
}

// RegisterOnHTTPClientResponseReceived registers callbacks to be invoked when an HTTP client response is received.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon HTTP client response received, each accepting no parameters.
func (m *Sensor[T]) RegisterOnHTTPClientResponseReceived(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnHTTPClientResponseReceived = append(m.OnHTTPClientResponseReceived, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnHTTPClientResponseReceived, target: %v => Registered OnHTTPClientResponseReceived function", m.componentMetadata, callback)
	}
}

// RegisterOnHTTPClientError registers callbacks to be invoked upon encountering an error during an HTTP request.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon HTTP client error, each accepting an error parameter.
func (m *Sensor[T]) RegisterOnHTTPClientError(callback ...func(c types.ComponentMetadata, err error)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnHTTPClientError = append(m.OnHTTPClientError, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnHTTPClientError, target: %v => Registered OnHTTPClientError function", m.componentMetadata, callback)
	}
}

// RegisterOnHTTPClientRequestComplete registers callbacks to be invoked when an HTTP client request is completed.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon HTTP client request complete, each accepting no parameters.
func (m *Sensor[T]) RegisterOnHTTPClientRequestComplete(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnHTTPClientRequestComplete = append(m.OnHTTPClientRequestComplete, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnHTTPClientRequestComplete, target: %v => Registered OnHTTPClientRequestComplete function", m.componentMetadata, callback)
	}
}

// RegisterOnSurgeProtectorReleaseToken registers callbacks to be invoked when the surge protector releases a token.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector token release, each accepting no parameters.
func (m *Sensor[T]) RegisterOnSurgeProtectorReleaseToken(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSurgeProtectorReleaseToken = append(m.OnSurgeProtectorReleaseToken, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorReleaseToken, target: %v => Registered OnSurgeProtectorReleaseToken function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorReleaseToken triggers callbacks registered for surge protector token release events.
func (m *Sensor[T]) InvokeOnSurgeProtectorReleaseToken(c types.ComponentMetadata) {
	for _, callback := range m.OnSurgeProtectorReleaseToken {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorReleaseToken, target: %v => Invoked OnSurgeProtectorReleaseToken function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorConnectResister registers callbacks to be invoked when the surge protector connects a resister.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector resister connection, each accepting no parameters.
func (m *Sensor[T]) RegisterOnSurgeProtectorConnectResister(callback ...func(c types.ComponentMetadata, r types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSurgeProtectorConnectResister = append(m.OnSurgeProtectorConnectResister, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorConnectResister, target: %v => Registered OnSurgeProtectorConnectResister function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorConnectResister triggers callbacks registered for surge protector resister connection events.
func (m *Sensor[T]) InvokeOnSurgeProtectorConnectResister(c types.ComponentMetadata, r types.ComponentMetadata) {
	for _, callback := range m.OnSurgeProtectorConnectResister {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorConnectResister, target: %v => Invoked OnSurgeProtectorConnectResister function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c, r)
		m.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerDrop registers a callback to be invoked when a circuit breaker drops an element.
func (s *Sensor[T]) RegisterOnSurgeProtectorBackupWireSubmit(callback ...func(cmd types.ComponentMetadata, elem T)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnSurgeProtectorBackupWireSubmit = append(s.OnSurgeProtectorBackupWireSubmit, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerDrop, target: %v => Registered OnCircuitBreakerDrop function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerDrop invokes registered callbacks for the circuit breaker drop event, passing the dropped element to each callback.
func (s *Sensor[T]) InvokeOnSurgeProtectorBackupWireSubmit(cmd types.ComponentMetadata, elem T) {
	for _, callback := range s.OnSurgeProtectorBackupWireSubmit {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerDrop, target: %v => Invoked OnCircuitBreakerDrop function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, elem)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerDrop registers a callback to be invoked when a circuit breaker drops an element.
func (s *Sensor[T]) RegisterOnSurgeProtectorDrop(callback ...func(cmd types.ComponentMetadata, elem T)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnSurgeProtectorDrop = append(s.OnSurgeProtectorDrop, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerDrop, target: %v => Registered OnCircuitBreakerDrop function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerDrop invokes registered callbacks for the circuit breaker drop event, passing the dropped element to each callback.
func (s *Sensor[T]) InvokeOnSurgeProtectorDrop(cmd types.ComponentMetadata, elem T) {
	for _, callback := range s.OnSurgeProtectorDrop {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCircuitBreakerDrop, target: %v => Invoked OnCircuitBreakerDrop function", s.componentMetadata, callback)
		s.callbackLock.Lock()
		callback(cmd, elem)
		s.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorDetachedBackups registers callbacks to be invoked when the surge protector detects detached backups.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector detached backups detection, each accepting no parameters.
func (m *Sensor[T]) RegisterOnSurgeProtectorDetachedBackups(callback ...func(c types.ComponentMetadata, bu types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSurgeProtectorDetachedBackups = append(m.OnSurgeProtectorDetachedBackups, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorDetachedBackups, target: %v => Registered OnSurgeProtectorDetachedBackups function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorDetachedBackups triggers callbacks registered for surge protector detached backups detection events.
func (m *Sensor[T]) InvokeOnSurgeProtectorDetachedBackups(c types.ComponentMetadata, bu types.ComponentMetadata) {
	for _, callback := range m.OnSurgeProtectorDetachedBackups {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorDetachedBackups, target: %v => Invoked OnSurgeProtectorDetachedBackups function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c, bu)
		m.callbackLock.Unlock()
	}
}

func (s *Sensor[T]) decorateHttpClientCallbacks() []types.Option[types.Sensor[T]] {
	var httpOptions []types.Option[types.Sensor[T]]
	httpOptions = append(httpOptions,
		WithOnHTTPClientErrorFunc[T](func(c types.ComponentMetadata, err error) {
			s.incrementMeterCounters(types.MetricTotalErrorCount)
			s.incrementMeterCounters(types.MetricHTTPClientErrorCount)
		}),
		WithOnHTTPClientRequestCompleteFunc[T](func(c types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricHTTPRequestCompletedCount)
		}),
		WithOnHTTPClientRequestStartFunc[T](func(c types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricHTTPRequestMadeCount)
		}),
		WithOnHTTPClientResponseReceivedFunc[T](func(c types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricHTTPResponseCount)
		}),
	)
	return httpOptions
}
