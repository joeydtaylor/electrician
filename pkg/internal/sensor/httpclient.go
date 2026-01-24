package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// RegisterOnHTTPClientRequestStart registers callbacks for HTTP request start.
func (s *Sensor[T]) RegisterOnHTTPClientRequestStart(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnHTTPClientRequestStart = append(s.OnHTTPClientRequestStart, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnHTTPClientResponseReceived registers callbacks for HTTP response received.
func (s *Sensor[T]) RegisterOnHTTPClientResponseReceived(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnHTTPClientResponseReceived = append(s.OnHTTPClientResponseReceived, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnHTTPClientError registers callbacks for HTTP errors.
func (s *Sensor[T]) RegisterOnHTTPClientError(callback ...func(types.ComponentMetadata, error)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnHTTPClientError = append(s.OnHTTPClientError, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnHTTPClientRequestComplete registers callbacks for HTTP request completion.
func (s *Sensor[T]) RegisterOnHTTPClientRequestComplete(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnHTTPClientRequestComplete = append(s.OnHTTPClientRequestComplete, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnHTTPClientRequestStart invokes callbacks for HTTP request start.
func (s *Sensor[T]) InvokeOnHTTPClientRequestStart(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnHTTPClientRequestStart) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// InvokeOnHTTPClientResponseReceived invokes callbacks for HTTP response received.
func (s *Sensor[T]) InvokeOnHTTPClientResponseReceived(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnHTTPClientResponseReceived) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// InvokeOnHTTPClientError invokes callbacks for HTTP errors.
func (s *Sensor[T]) InvokeOnHTTPClientError(c types.ComponentMetadata, err error) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnHTTPClientError) {
		if cb == nil {
			continue
		}
		cb(c, err)
	}
}

// InvokeOnHTTPClientRequestComplete invokes callbacks for HTTP request completion.
func (s *Sensor[T]) InvokeOnHTTPClientRequestComplete(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnHTTPClientRequestComplete) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

func (s *Sensor[T]) decorateHttpClientCallbacks() []types.Option[types.Sensor[T]] {
	return []types.Option[types.Sensor[T]]{
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
	}
}
