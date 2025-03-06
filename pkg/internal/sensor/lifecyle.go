package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (m *Sensor[T]) RegisterOnRestart(callback ...func(c types.ComponentMetadata)) {
	m.callbackLock.Lock()
	defer m.callbackLock.Unlock()
	m.OnRestart = append(m.OnRestart, callback...)
	for _, callback := range callback {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RegisterOnSurgeProtectorDetachedBackups, target: %v => Registered OnSurgeProtectorDetachedBackups function", m.componentMetadata, callback)
	}
}

func (m *Sensor[T]) InvokeOnRestart(c types.ComponentMetadata) {
	for _, callback := range m.OnRestart {
		m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: InvokeOnSurgeProtectorDetachedBackups, target: %v => Invoked OnSurgeProtectorDetachedBackups function", m.componentMetadata, callback)
		m.callbackLock.Lock()
		callback(c)
		m.callbackLock.Unlock()
	}
}
