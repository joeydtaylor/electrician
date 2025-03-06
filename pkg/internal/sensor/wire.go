package sensor

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// RegisterOnCancel registers callbacks to be invoked when an operation is canceled.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon cancellation, each accepting an element of type T.
func (m *Sensor[T]) RegisterOnCancel(callback ...func(c types.ComponentMetadata, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnCancel = append(m.OnCancel, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCancel, target: %v => Registered OnCancel function", m.componentMetadata, callback)
	}
}

// RegisterOnCancel registers callbacks to be invoked when an operation is canceled.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon cancellation, each accepting an element of type T.
func (m *Sensor[T]) RegisterOnSubmit(callback ...func(c types.ComponentMetadata, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnSubmit = append(m.OnSubmit, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSubmit, target: %v => Registered OnSubmit function", m.componentMetadata, callback)
	}
}

// RegisterOnComplete registers callbacks to be invoked upon completion of an operation.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon completion, each accepting no parasensors.
func (m *Sensor[T]) RegisterOnComplete(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnComplete = append(m.OnComplete, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnComplete, target: %v => Registered OnComplete function", m.componentMetadata, callback)
	}
}

// RegisterOnElementProcessed registers callbacks to be invoked when an element is processed.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon processing an element, each accepting an element of type T.
func (m *Sensor[T]) RegisterOnElementProcessed(callback ...func(c types.ComponentMetadata, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnElementProcessed = append(m.OnElementProcessed, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnElementProcessed, target: %v => Registered OnElementProcessed function", m.componentMetadata, callback)
	}
}

// RegisterOnError registers callbacks to be invoked upon encountering an error during processing.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon error, each accepting an error and an element of type T.
func (m *Sensor[T]) RegisterOnError(callback ...func(c types.ComponentMetadata, err error, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnError = append(m.OnError, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnError, target: %v => Registered OnError function", m.componentMetadata, callback)
	}
}

// RegisterOnStart registers callbacks to be invoked at the start of an operation.
//
// Parameters:
//   - callback: One or more callback functions to be invoked at the start, each accepting no parasensors.
func (m *Sensor[T]) RegisterOnStart(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnStart = append(m.OnStart, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnStart, target: %v => Registered OnStart function", m.componentMetadata, callback)
	}
}

// RegisterOnTerminate registers callbacks to be invoked upon termination of the Sensor.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon termination, each accepting no parasensors.
func (m *Sensor[T]) RegisterOnTerminate(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnTerminate = append(m.OnTerminate, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnTerminate, target: %v => Registered OnTerminate function", m.componentMetadata, callback)
	}
}

// RegisterOnCircuitBreakerDrop registers a callback to be invoked when a circuit breaker drops an element.
func (s *Sensor[T]) RegisterOnInsulatorAttempt(callback ...func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnInsulatorAttempt = append(s.OnInsulatorAttempt, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerDrop, target: %v => Registered OnCircuitBreakerDrop function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerDrop invokes registered callbacks for the circuit breaker drop event, passing the dropped element to each callback.
func (s *Sensor[T]) InvokeOnInsulatorAttempt(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	for _, callback := range s.OnInsulatorAttempt {
		s.callbackLock.Lock()
		callback(c, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerDrop registers a callback to be invoked when a circuit breaker drops an element.
func (s *Sensor[T]) RegisterOnInsulatorSuccess(callback ...func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnInsulatorSuccess = append(s.OnInsulatorSuccess, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerDrop, target: %v => Registered OnCircuitBreakerDrop function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerDrop invokes registered callbacks for the circuit breaker drop event, passing the dropped element to each callback.
func (s *Sensor[T]) InvokeOnInsulatorSuccess(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	for _, callback := range s.OnInsulatorSuccess {
		s.callbackLock.Lock()
		callback(c, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
		s.callbackLock.Unlock()
	}
}

// RegisterOnCircuitBreakerDrop registers a callback to be invoked when a circuit breaker drops an element.
func (s *Sensor[T]) RegisterOnInsulatorFailure(callback ...func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)) {
	s.callbackLock.Lock()
	defer s.callbackLock.Unlock()
	s.OnInsulatorFailure = append(s.OnInsulatorFailure, callback...)
	for _, cb := range callback {
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnCircuitBreakerDrop, target: %v => Registered OnCircuitBreakerDrop function", s.componentMetadata, cb)
	}
}

// InvokeOnCircuitBreakerDrop invokes registered callbacks for the circuit breaker drop event, passing the dropped element to each callback.
func (s *Sensor[T]) InvokeOnInsulatorFailure(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	for _, callback := range s.OnInsulatorFailure {
		s.callbackLock.Lock()
		callback(c, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
		s.callbackLock.Unlock()
	}
}

// InvokeOnCancel triggers callbacks registered for the cancellation event, passing the relevant element to each callback.
//
// Parameters:
//   - elem: The relevant element associated with the cancellation event.
func (m *Sensor[T]) InvokeOnCancel(c types.ComponentMetadata, elem T) {
	for _, callback := range m.OnCancel {

		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnCancel, element: %v, target: %v => Invoked OnCancel function", m.componentMetadata, elem, callback)
		m.callbackLock.Lock()
		callback(c, elem)
		m.callbackLock.Unlock()
	}
}

// InvokeOnComplete triggers callbacks registered for the completion event.
func (m *Sensor[T]) InvokeOnComplete(c types.ComponentMetadata) {
	for _, callback := range m.OnComplete {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnComplete, target: %v => Invoked OnComplete function", m.componentMetadata, callback)
		m.callbackLock.Lock()

		callback(c)
		m.callbackLock.Unlock()
	}
}

// InvokeOnElementProcessed triggers callbacks registered for processing events, passing the processed element to each callback.
//
// Parameters:
//   - elem: The processed element associated with the processing event.
func (m *Sensor[T]) InvokeOnElementProcessed(c types.ComponentMetadata, elem T) {
	for _, callback := range m.OnElementProcessed {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnElementProcessed, element: %v, target: %v => Invoked OnElementProcessed function", m.componentMetadata, elem, callback)
		m.callbackLock.Lock()
		callback(c, elem)
		m.callbackLock.Unlock()
	}
}

// InvokeOnError triggers error-handling callbacks, passing the relevant error and element to each callback.
//
// Parameters:
//   - err: The error encountered during processing.
//   - elem: The relevant element associated with the error.
func (m *Sensor[T]) InvokeOnError(c types.ComponentMetadata, err error, elem T) {
	for _, callback := range m.OnError {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnError, element: %v, error: %v, target: %v => Invoked OnError function", m.componentMetadata, elem, err, callback)
		m.callbackLock.Lock()
		callback(c, err, elem)
		m.callbackLock.Unlock()
	}
}

// InvokeOnStart triggers callbacks registered for the start event.
func (m *Sensor[T]) InvokeOnStart(c types.ComponentMetadata) {
	for _, callback := range m.OnStart {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnStart, target: %v => Invoked OnStart function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

// InvokeOnTerminate triggers callbacks registered for the termination event.
func (m *Sensor[T]) InvokeOnStop(c types.ComponentMetadata) {
	for _, callback := range m.OnTerminate {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnTerminate, target: %v => Invoked OnTerminate function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}

func (m *Sensor[T]) InvokeOnSubmit(c types.ComponentMetadata, elem T) {
	for _, callback := range m.OnSubmit {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSubmit, element: %v, error: %v, target: %v => Invoked OnSubmit function", m.componentMetadata, elem, callback)
		m.callbackLock.Lock()
		callback(c, elem)
		m.callbackLock.Unlock()
	}
}

func (s *Sensor[T]) decorateWireCallbacks() []types.Option[types.Sensor[T]] {
	var wireOptions []types.Option[types.Sensor[T]]
	wireOptions = append(wireOptions,
		WithOnSubmitFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCountersAndReportActivity(types.MetricTotalSubmittedCount)
		}),
		WithOnErrorFunc[T](func(c types.ComponentMetadata, err error, elem T) {
			s.incrementMeterCounters(types.MetricTotalErrorCount)
			s.incrementMeterCountersAndReportActivity(types.MetricTotalProcessedCount)
			s.incrementMeterCountersAndReportActivity(types.MetricTransformationErrorPercentage)
		}),
		WithOnRestartFunc[T](func(c types.ComponentMetadata) {
			switch c.Type {
			case "WIRE":
				s.decrementMeterCounters(types.MetricComponentRunningCount)
				s.decrementMeterCounters(types.MetricWireRunningCount)
			}
		}),
		WithOnElementProcessedFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCountersAndReportActivity(types.MetricTotalProcessedCount)
			s.incrementMeterCountersAndReportActivity(types.MetricTotalTransformedCount)
		}),
		WithOnCancelFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricWireElementSubmitCancelCount)
		}),
	)
	return wireOptions
}
