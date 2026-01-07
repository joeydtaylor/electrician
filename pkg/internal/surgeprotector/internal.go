package surgeprotector

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (s *SurgeProtector[T]) snapshotSensors() []types.Sensor[T] {
	s.sensorLock.Lock()
	defer s.sensorLock.Unlock()
	if len(s.sensors) == 0 {
		return nil
	}
	out := make([]types.Sensor[T], len(s.sensors))
	copy(out, s.sensors)
	return out
}

func (s *SurgeProtector[T]) snapshotLoggers() []types.Logger {
	s.loggersLock.Lock()
	defer s.loggersLock.Unlock()
	if len(s.loggers) == 0 {
		return nil
	}
	out := make([]types.Logger, len(s.loggers))
	copy(out, s.loggers)
	return out
}

func (s *SurgeProtector[T]) notifySurgeProtectorTrip() {
	for _, sensor := range s.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorTrip(s.componentMetadata)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorReset() {
	for _, sensor := range s.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorReset(s.componentMetadata)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorBackupFailure(err error) {
	for _, sensor := range s.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorBackupFailure(s.componentMetadata, err)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorBackupSubmission(elem T) {
	for _, sensor := range s.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorBackupWireSubmit(s.componentMetadata, elem)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorCreation() {
	for _, sensor := range s.snapshotSensors() {
		if sensor == nil {
			continue
		}
		for _, ms := range sensor.GetMeters() {
			ms.IncrementCount(types.MetricSurgeCount)
		}
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorReleaseToken() {
	for _, sensor := range s.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorReleaseToken(s.componentMetadata)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorConnectResister(r types.Resister[T]) {
	meta := r.GetComponentMetadata()
	for _, sensor := range s.snapshotSensors() {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnSurgeProtectorConnectResister(s.componentMetadata, meta)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorDetachedBackups(bs ...types.Submitter[T]) {
	sensors := s.snapshotSensors()
	for _, sensor := range sensors {
		if sensor == nil {
			continue
		}
		for _, b := range bs {
			if b == nil {
				continue
			}
			sensor.InvokeOnSurgeProtectorDetachedBackups(s.componentMetadata, b.GetComponentMetadata())
		}
	}
}

func (s *SurgeProtector[T]) submitToComponents(ctx context.Context, elem T) error {
	s.mu.Lock()
	components := make([]types.Submitter[T], len(s.managedComponents))
	copy(components, s.managedComponents)
	s.mu.Unlock()

	for _, component := range components {
		if component == nil {
			continue
		}
		if err := component.Submit(ctx, elem); err != nil {
			return err
		}
	}
	return nil
}
