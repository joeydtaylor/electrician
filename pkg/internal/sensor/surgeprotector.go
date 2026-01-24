package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *Sensor[T]) decorateSurgeProtectorCallbacks() []types.Option[types.Sensor[T]] {
	return []types.Option[types.Sensor[T]]{
		WithSurgeProtectorTripFunc[T](func(c types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricSurgeTripCount)
			s.incrementMeterCounters(types.MetricSurgeProtectorCurrentTripCount)
		}),
		WithSurgeProtectorResetFunc[T](func(c types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricSurgeResetCount)
			s.decrementMeterCounters(types.MetricSurgeProtectorCurrentTripCount)
		}),
		WithSurgeProtectorBackupFailureFunc[T](func(c types.ComponentMetadata, err error) {
			s.incrementMeterCounters(types.MetricSurgeBackupFailureCount)
		}),
		WithSurgeProtectorRateLimitExceededFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricSurgeRateLimitExceedCount)
		}),
		WithSurgeProtectorBackupWireSubmissionFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricSurgeProtectorBackupWireSubmissionCount)
		}),
		WithSurgeProtectorDroppedSubmissionFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricSurgeProtectorDropCount)
		}),
		WithSurgeProtectorSubmitFunc[T](func(c types.ComponentMetadata, elem T) {
			s.incrementMeterCounters(types.MetricSurgeProtectorAttachedCount)
		}),
		WithSurgeProtectorConnectResisterFunc[T](func(c types.ComponentMetadata, r types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricResisterConnectedCount)
		}),
	}
}

// RegisterOnSurgeProtectorTrip registers callbacks for trip events.
func (s *Sensor[T]) RegisterOnSurgeProtectorTrip(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorTrip = append(s.OnSurgeProtectorTrip, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorTrip invokes callbacks for trip events.
func (s *Sensor[T]) InvokeOnSurgeProtectorTrip(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorTrip) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// RegisterOnSurgeProtectorReset registers callbacks for reset events.
func (s *Sensor[T]) RegisterOnSurgeProtectorReset(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorReset = append(s.OnSurgeProtectorReset, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorReset invokes callbacks for reset events.
func (s *Sensor[T]) InvokeOnSurgeProtectorReset(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorReset) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// RegisterOnSurgeProtectorBackupFailure registers callbacks for backup failures.
func (s *Sensor[T]) RegisterOnSurgeProtectorBackupFailure(callback ...func(types.ComponentMetadata, error)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorBackupFailure = append(s.OnSurgeProtectorBackupFailure, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorBackupFailure invokes callbacks for backup failures.
func (s *Sensor[T]) InvokeOnSurgeProtectorBackupFailure(c types.ComponentMetadata, err error) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorBackupFailure) {
		if cb == nil {
			continue
		}
		cb(c, err)
	}
}

// RegisterOnSurgeProtectorRateLimitExceeded registers callbacks for rate limit events.
func (s *Sensor[T]) RegisterOnSurgeProtectorRateLimitExceeded(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorRateLimitExceeded = append(s.OnSurgeProtectorRateLimitExceeded, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorRateLimitExceeded invokes callbacks for rate limit events.
func (s *Sensor[T]) InvokeOnSurgeProtectorRateLimitExceeded(c types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorRateLimitExceeded) {
		if cb == nil {
			continue
		}
		cb(c, elem)
	}
}

// RegisterOnSurgeProtectorBackupWireSubmit registers callbacks for backup submissions.
func (s *Sensor[T]) RegisterOnSurgeProtectorBackupWireSubmit(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorBackupWireSubmit = append(s.OnSurgeProtectorBackupWireSubmit, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorBackupWireSubmit invokes callbacks for backup submissions.
func (s *Sensor[T]) InvokeOnSurgeProtectorBackupWireSubmit(c types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorBackupWireSubmit) {
		if cb == nil {
			continue
		}
		cb(c, elem)
	}
}

// RegisterOnSurgeProtectorSubmit registers callbacks for submissions.
func (s *Sensor[T]) RegisterOnSurgeProtectorSubmit(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorSubmit = append(s.OnSurgeProtectorSubmit, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorSubmit invokes callbacks for submissions.
func (s *Sensor[T]) InvokeOnSurgeProtectorSubmit(c types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorSubmit) {
		if cb == nil {
			continue
		}
		cb(c, elem)
	}
}

// RegisterOnSurgeProtectorDrop registers callbacks for dropped elements.
func (s *Sensor[T]) RegisterOnSurgeProtectorDrop(callback ...func(types.ComponentMetadata, T)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorDrop = append(s.OnSurgeProtectorDrop, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorDrop invokes callbacks for dropped elements.
func (s *Sensor[T]) InvokeOnSurgeProtectorDrop(c types.ComponentMetadata, elem T) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorDrop) {
		if cb == nil {
			continue
		}
		cb(c, elem)
	}
}

// RegisterOnSurgeProtectorReleaseToken registers callbacks for token release events.
func (s *Sensor[T]) RegisterOnSurgeProtectorReleaseToken(callback ...func(types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorReleaseToken = append(s.OnSurgeProtectorReleaseToken, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorReleaseToken invokes callbacks for token release events.
func (s *Sensor[T]) InvokeOnSurgeProtectorReleaseToken(c types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorReleaseToken) {
		if cb == nil {
			continue
		}
		cb(c)
	}
}

// RegisterOnSurgeProtectorConnectResister registers callbacks for resister connections.
func (s *Sensor[T]) RegisterOnSurgeProtectorConnectResister(callback ...func(types.ComponentMetadata, types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorConnectResister = append(s.OnSurgeProtectorConnectResister, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorConnectResister invokes callbacks for resister connections.
func (s *Sensor[T]) InvokeOnSurgeProtectorConnectResister(c types.ComponentMetadata, r types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorConnectResister) {
		if cb == nil {
			continue
		}
		cb(c, r)
	}
}

// RegisterOnSurgeProtectorDetachedBackups registers callbacks for detached backup systems.
func (s *Sensor[T]) RegisterOnSurgeProtectorDetachedBackups(callback ...func(types.ComponentMetadata, types.ComponentMetadata)) {
	if len(callback) == 0 {
		return
	}

	s.callbackLock.Lock()
	s.OnSurgeProtectorDetachedBackups = append(s.OnSurgeProtectorDetachedBackups, callback...)
	s.callbackLock.Unlock()
}

// InvokeOnSurgeProtectorDetachedBackups invokes callbacks for detached backup systems.
func (s *Sensor[T]) InvokeOnSurgeProtectorDetachedBackups(c types.ComponentMetadata, b types.ComponentMetadata) {
	for _, cb := range snapshotCallbacks(&s.callbackLock, s.OnSurgeProtectorDetachedBackups) {
		if cb == nil {
			continue
		}
		cb(c, b)
	}
}
