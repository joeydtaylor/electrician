package sensor

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// RegisterOnCancel registers callbacks for cancel events.
func (s *Sensor[T]) RegisterOnCancel(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnCancel = append(s.OnCancel, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnSubmit registers callbacks for submit events.
func (s *Sensor[T]) RegisterOnSubmit(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSubmit = append(s.OnSubmit, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnComplete registers callbacks for completion events.
func (s *Sensor[T]) RegisterOnComplete(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnComplete = append(s.OnComplete, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnElementProcessed registers callbacks for processed elements.
func (s *Sensor[T]) RegisterOnElementProcessed(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnElementProcessed = append(s.OnElementProcessed, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnError registers callbacks for error events.
func (s *Sensor[T]) RegisterOnError(callback ...func(types.ComponentMetadata, error, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnError = append(s.OnError, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnStart registers callbacks for start events.
func (s *Sensor[T]) RegisterOnStart(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnStart = append(s.OnStart, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnTerminate registers callbacks for stop events.
func (s *Sensor[T]) RegisterOnTerminate(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnTerminate = append(s.OnTerminate, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnInsulatorAttempt registers callbacks for insulator attempt events.
func (s *Sensor[T]) RegisterOnInsulatorAttempt(callback ...func(types.ComponentMetadata, T, T, error, error, int, int, time.Duration)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnInsulatorAttempt = append(s.OnInsulatorAttempt, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnInsulatorSuccess registers callbacks for insulator success events.
func (s *Sensor[T]) RegisterOnInsulatorSuccess(callback ...func(types.ComponentMetadata, T, T, error, error, int, int, time.Duration)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnInsulatorSuccess = append(s.OnInsulatorSuccess, callback...)
	s.callbackLock.Unlock()
}

// RegisterOnInsulatorFailure registers callbacks for insulator failure events.
func (s *Sensor[T]) RegisterOnInsulatorFailure(callback ...func(types.ComponentMetadata, T, T, error, error, int, int, time.Duration)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnInsulatorFailure = append(s.OnInsulatorFailure, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnInsulatorAttempt invokes callbacks for insulator attempts.
func (s *Sensor[T]) InvokeOnInsulatorAttempt(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnInsulatorAttempt) {
		if cb == nil {
			continue
		}
		cb(c, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

// InvokeOnInsulatorSuccess invokes callbacks for insulator success events.
func (s *Sensor[T]) InvokeOnInsulatorSuccess(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnInsulatorSuccess) {
		if cb == nil {
			continue
		}
		cb(c, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

// InvokeOnInsulatorFailure invokes callbacks for insulator failure events.
func (s *Sensor[T]) InvokeOnInsulatorFailure(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnInsulatorFailure) {
		if cb == nil {
			continue
		}
		cb(c, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

// InvokeOnCancel invokes callbacks for cancel events.
func (s *Sensor[T]) InvokeOnCancel(c types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnCancel) {
		if cb == nil {
			continue
		}
		cb(c, elem)
	}
}

// InvokeOnComplete invokes callbacks for completion events.
func (s *Sensor[T]) InvokeOnComplete(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnComplete) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// InvokeOnElementProcessed invokes callbacks for processed elements.
func (s *Sensor[T]) InvokeOnElementProcessed(c types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnElementProcessed) {
		if cb == nil {
			continue
		}
		cb(c, elem)
	}
}

// InvokeOnError invokes callbacks for error events.
func (s *Sensor[T]) InvokeOnError(c types.ComponentMetadata, err error, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnError) {
		if cb == nil {
			continue
		}
		cb(c, err, elem)
	}
}

// InvokeOnStart invokes callbacks for start events.
func (s *Sensor[T]) InvokeOnStart(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnStart) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// InvokeOnStop invokes callbacks for stop events.
func (s *Sensor[T]) InvokeOnStop(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnTerminate) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// InvokeOnSubmit invokes callbacks for submit events.
func (s *Sensor[T]) InvokeOnSubmit(c types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSubmit) {
		if cb == nil {
			continue
		}
		cb(c, elem)
	}
}

func (s *Sensor[T]) decorateWireCallbacks() []types.Option[types.Sensor[T]] {
	return []types.Option[types.Sensor[T]]{
		WithOnSubmitFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCountersAndReportActivity(types.MetricTotalSubmittedCount)
		}),
		WithOnErrorFunc[T](func(c types.ComponentMetadata, err error, elem T) {
			s.incrementMeterCounters(types.MetricTotalErrorCount)
			s.incrementMeterCountersAndReportActivity(types.MetricTotalProcessedCount)
			s.incrementMeterCountersAndReportActivity(types.MetricTransformationErrorPercentage)
		}),
		WithOnRestartFunc[T](func(c types.ComponentMetadata) {
			if c.Type == "WIRE" {
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
	}
}
