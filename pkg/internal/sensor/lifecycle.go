package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// RegisterOnRestart registers callbacks for restart events.
func (s *Sensor[T]) RegisterOnRestart(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnRestart = append(s.OnRestart, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnRestart invokes registered restart callbacks.
func (s *Sensor[T]) InvokeOnRestart(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnRestart) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}
