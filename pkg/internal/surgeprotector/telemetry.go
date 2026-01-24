package surgeprotector

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers emits a log entry to all configured loggers.
func (s *SurgeProtector[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := s.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	type levelChecker interface {
		IsLevelEnabled(types.LogLevel) bool
	}

	for _, logger := range loggers {
		if logger == nil {
			continue
		}
		if lc, ok := logger.(levelChecker); ok {
			if !lc.IsLevelEnabled(level) {
				continue
			}
		} else if logger.GetLevel() > level {
			continue
		}

		switch level {
		case types.DebugLevel:
			logger.Debug(msg, keysAndValues...)
		case types.InfoLevel:
			logger.Info(msg, keysAndValues...)
		case types.WarnLevel:
			logger.Warn(msg, keysAndValues...)
		case types.ErrorLevel:
			logger.Error(msg, keysAndValues...)
		case types.DPanicLevel:
			logger.DPanic(msg, keysAndValues...)
		case types.PanicLevel:
			logger.Panic(msg, keysAndValues...)
		case types.FatalLevel:
			logger.Fatal(msg, keysAndValues...)
		}
	}
}

func (s *SurgeProtector[T]) snapshotLoggers() []types.Logger {
	s.loggersLock.Lock()
	loggers := append([]types.Logger(nil), s.loggers...)
	s.loggersLock.Unlock()
	return loggers
}

func (s *SurgeProtector[T]) snapshotSensors() []types.Sensor[T] {
	s.sensorLock.Lock()
	sensors := append([]types.Sensor[T](nil), s.sensors...)
	s.sensorLock.Unlock()
	return sensors
}

func (s *SurgeProtector[T]) snapshotManagedComponents() []types.Submitter[T] {
	s.managedLock.Lock()
	components := append([]types.Submitter[T](nil), s.managedComponents...)
	s.managedLock.Unlock()
	return components
}

func (s *SurgeProtector[T]) snapshotBackups() []types.Submitter[T] {
	s.backupLock.Lock()
	backups := append([]types.Submitter[T](nil), s.backupSystems...)
	s.backupLock.Unlock()
	return backups
}

func (s *SurgeProtector[T]) snapshotResister() types.Resister[T] {
	s.resisterLock.Lock()
	r := s.resister
	s.resisterLock.Unlock()
	return r
}

func (s *SurgeProtector[T]) snapshotMetadata() types.ComponentMetadata {
	s.metadataLock.Lock()
	metadata := s.componentMetadata
	s.metadataLock.Unlock()
	return metadata
}

func (s *SurgeProtector[T]) notifySurgeProtectorTrip() {
	metadata := s.snapshotMetadata()
	for _, sensor := range s.snapshotSensors() {
		sensor.InvokeOnSurgeProtectorTrip(metadata)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorReset() {
	metadata := s.snapshotMetadata()
	for _, sensor := range s.snapshotSensors() {
		sensor.InvokeOnSurgeProtectorReset(metadata)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorBackupFailure(err error) {
	metadata := s.snapshotMetadata()
	for _, sensor := range s.snapshotSensors() {
		sensor.InvokeOnSurgeProtectorBackupFailure(metadata, err)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorBackupSubmission(elem T) {
	metadata := s.snapshotMetadata()
	for _, sensor := range s.snapshotSensors() {
		sensor.InvokeOnSurgeProtectorBackupWireSubmit(metadata, elem)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorCreation(sensors []types.Sensor[T]) {
	if len(sensors) == 0 {
		return
	}
	for _, sensor := range sensors {
		for _, meter := range sensor.GetMeters() {
			meter.IncrementCount(types.MetricSurgeCount)
		}
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorReleaseToken() {
	metadata := s.snapshotMetadata()
	for _, sensor := range s.snapshotSensors() {
		sensor.InvokeOnSurgeProtectorReleaseToken(metadata)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorConnectResister(r types.Resister[T]) {
	metadata := s.snapshotMetadata()
	rmeta := r.GetComponentMetadata()
	for _, sensor := range s.snapshotSensors() {
		sensor.InvokeOnSurgeProtectorConnectResister(metadata, rmeta)
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorDetachedBackups(backups ...types.Submitter[T]) {
	metadata := s.snapshotMetadata()
	for _, sensor := range s.snapshotSensors() {
		for _, backup := range backups {
			if backup == nil {
				continue
			}
			sensor.InvokeOnSurgeProtectorDetachedBackups(metadata, backup.GetComponentMetadata())
		}
	}
}

func (s *SurgeProtector[T]) notifySurgeProtectorSubmit(elem T) {
	metadata := s.snapshotMetadata()
	for _, sensor := range s.snapshotSensors() {
		sensor.InvokeOnSurgeProtectorSubmit(metadata, elem)
	}
}
