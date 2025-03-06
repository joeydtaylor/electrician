package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *Sensor[T]) decorateResisterCallbacks() []types.Option[types.Sensor[T]] {
	var resisterOptions []types.Option[types.Sensor[T]]
	resisterOptions = append(resisterOptions,
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
	)
	return resisterOptions
}

// RegisterOnSurgeProtectorQueueProcessed registers callbacks to be invoked when the surge protector processes its queue.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector queue processing, each accepting no parameters.
func (m *Sensor[T]) RegisterOnResisterDequeued(callback ...func(r types.ComponentMetadata, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnResisterDequeued = append(m.OnResisterDequeued, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorQueueProcessed, target: %v => Registered OnResisterDequeued function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorResisterProcessed triggers callbacks registered for surge protector queue processed events.
func (m *Sensor[T]) InvokeOnResisterDequeued(r types.ComponentMetadata, elem T) {
	for _, callback := range m.OnResisterDequeued {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnResisterDequeued, target: %v => Invoked OnResisterDequeued function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(r, elem)
		m.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorQueueProcessed registers callbacks to be invoked when the surge protector processes its queue.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector queue processing, each accepting no parameters.
func (m *Sensor[T]) RegisterOnResisterQueued(callback ...func(r types.ComponentMetadata, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnResisterQueued = append(m.OnResisterQueued, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnResisterQueued, target: %v => Registered OnResisterQueued function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorResisterProcessed triggers callbacks registered for surge protector queue processed events.
func (m *Sensor[T]) InvokeOnResisterQueued(r types.ComponentMetadata, elem T) {
	for _, callback := range m.OnResisterQueued {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnResisterQueued, target: %v => Invoked OnResisterQueued function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(r, elem)
		m.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorQueueEmpty registers callbacks to be invoked when the surge protector's queue becomes empty.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector queue empty, each accepting no parameters.
func (m *Sensor[T]) RegisterOnResisterEmpty(callback ...func(r types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnResisterEmpty = append(m.OnResisterEmpty, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorQueueEmpty, target: %v => Registered OnSurgeProtectorQueueEmpty function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorQueueEmpty triggers callbacks registered for surge protector queue empty events.
func (m *Sensor[T]) InvokeOnResisterEmpty(r types.ComponentMetadata) {
	for _, callback := range m.OnResisterEmpty {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorQueueEmpty, target: %v => Invoked OnSurgeProtectorQueueEmpty function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(r)
		m.callbackLock.Unlock()
	}
}

// RegisterOnSurgeProtectorQueueEmpty registers callbacks to be invoked when the surge protector's queue becomes empty.
//
// Parameters:
//   - callback: One or more callback functions to be invoked upon surge protector queue empty, each accepting no parameters.
func (m *Sensor[T]) RegisterOnResisterRequeued(callback ...func(r types.ComponentMetadata, elem T)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnResisterRequeued = append(m.OnResisterRequeued, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnResisterRequeued, target: %v => Registered RegisterOnResisterRequeued function", m.componentMetadata, callback)
	}
}

// InvokeOnSurgeProtectorQueueEmpty triggers callbacks registered for surge protector queue empty events.
func (m *Sensor[T]) InvokeOnResisterRequeued(r types.ComponentMetadata, elem T) {
	for _, callback := range m.OnResisterRequeued {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnResisterRequeued, target: %v => Invoked InvokeOnResisterRequeued function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(r, elem)
		m.callbackLock.Unlock()
	}
}
