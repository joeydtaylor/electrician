package surgeprotector

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// submitToComponents submits the element to all managed components.
func (s *SurgeProtector[T]) submitToComponents(ctx context.Context, elem T) error {
	for _, component := range s.managedComponents {
		if err := component.Submit(s.ctx, elem); err != nil {
			return err
		}
	}
	return nil
}

func (s *SurgeProtector[T]) notifySurgeProtectorTrip() {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorTrip(s.componentMetadata)

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorTrip, target_component: %s => Invoked InvokeOnSurgeProtectorTrip for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorReset() {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorReset(s.componentMetadata)

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorReset, target_component: %s => Invoked InvokeOnSurgeProtectorReset for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorBackupFailure(err error) {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorBackupFailure(s.componentMetadata, err)

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorBackupFailure, error: %v, target_component: %s => Invoked InvokeOnSurgeProtectorBackupFailure for sensor", s.GetComponentMetadata(), err, sensor.GetComponentMetadata())
	}
}

/* func (s *SurgeProtector[T]) notifySurgeProtectorResisterProcessed() {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorResisterProcessed(s.componentMetadata, s.resister.GetComponentMetadata())

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorResisterProcessed, target_component: %s => Invoked InvokeOnSurgeProtectorResisterProcessed for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
} */

/* func (s *SurgeProtector[T]) notifySurgeProtectorResisterEmpty() {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorResisterEmpty(s.componentMetadata, s.resister.GetComponentMetadata())

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorQueueEmpty, target_component: %s => Invoked InvokeOnSurgeProtectorQueueEmpty for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
} */

func (s *SurgeProtector[T]) notifySurgeProtectorBackupSubmission(elem T) {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorBackupWireSubmit(s.componentMetadata, elem)
		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorReleaseToken, target_component: %s => Invoked InvokeOnSurgeProtectorReleaseToken for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorCreation() {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		for _, ms := range sensor.GetMeters() {
			ms.IncrementCount(types.MetricSurgeCount)
		}

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorReleaseToken, target_component: %s => Invoked InvokeOnSurgeProtectorReleaseToken for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorReleaseToken() {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorReleaseToken(s.componentMetadata)

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorReleaseToken, target_component: %s => Invoked InvokeOnSurgeProtectorReleaseToken for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorConnectResister(r types.Resister[T]) {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		sensor.InvokeOnSurgeProtectorConnectResister(s.componentMetadata, r.GetComponentMetadata())

		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorConnectResister, target_component: %s => Invoked InvokeOnSurgeProtectorConnectResister for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorDetachedBackups(bs ...types.Submitter[T]) {
	for _, sensor := range s.sensors {
		s.sensorLock.Lock()
		defer s.sensorLock.Unlock()
		for _, b := range bs {
			sensor.InvokeOnSurgeProtectorDetachedBackups(s.componentMetadata, b.GetComponentMetadata())
		}
		s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifySurgeProtectorDetachedBackups, target_component: %s => Invoked InvokeOnSurgeProtectorDetachedBackups for sensor", s.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}
