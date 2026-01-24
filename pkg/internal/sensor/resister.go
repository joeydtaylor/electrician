package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *Sensor[T]) decorateResisterCallbacks() []types.Option[types.Sensor[T]] {
	return []types.Option[types.Sensor[T]]{
		WithResisterDequeuedFunc[T](func(r types.ComponentMetadata, elem T) {
			s.decrementMeterCounters(types.MetricResisterElementCurrentlyQueuedCount)
			s.incrementMeterCounters(types.MetricResisterElementDequeued)
		}),
		WithResisterEmptyFunc[T](func(r types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricResisterClearedCount)
		}),
		WithResisterQueuedFunc[T](func(r types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricResisterElementQueuedCount)
			s.incrementMeterCounters(types.MetricResisterElementCurrentlyQueuedCount)
		}),
		WithResisterRequeuedFunc[T](func(r types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricResisterElementRequeued)
		}),
	}
}

// RegisterOnResisterDequeued registers callbacks for dequeued elements.
func (s *Sensor[T]) RegisterOnResisterDequeued(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnResisterDequeued = append(s.OnResisterDequeued, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnResisterDequeued invokes callbacks for dequeued elements.
func (s *Sensor[T]) InvokeOnResisterDequeued(r types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnResisterDequeued) {
		if cb == nil {
			continue
		}
		cb(r, elem)
	}
}

// RegisterOnResisterQueued registers callbacks for queued elements.
func (s *Sensor[T]) RegisterOnResisterQueued(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnResisterQueued = append(s.OnResisterQueued, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnResisterQueued invokes callbacks for queued elements.
func (s *Sensor[T]) InvokeOnResisterQueued(r types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnResisterQueued) {
		if cb == nil {
			continue
		}
		cb(r, elem)
	}
}

// RegisterOnResisterEmpty registers callbacks for empty queue events.
func (s *Sensor[T]) RegisterOnResisterEmpty(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnResisterEmpty = append(s.OnResisterEmpty, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnResisterEmpty invokes callbacks for empty queue events.
func (s *Sensor[T]) InvokeOnResisterEmpty(r types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnResisterEmpty) {
		if cb == nil {
			continue
		}
		cb(r)
	}
}

// RegisterOnResisterRequeued registers callbacks for requeued elements.
func (s *Sensor[T]) RegisterOnResisterRequeued(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnResisterRequeued = append(s.OnResisterRequeued, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnResisterRequeued invokes callbacks for requeued elements.
func (s *Sensor[T]) InvokeOnResisterRequeued(r types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnResisterRequeued) {
		if cb == nil {
			continue
		}
		cb(r, elem)
	}
}
