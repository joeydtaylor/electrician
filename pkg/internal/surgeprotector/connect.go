package surgeprotector

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ConnectComponent attaches managed submitters for direct submission.
func (s *SurgeProtector[T]) ConnectComponent(components ...types.Submitter[T]) {
	if len(components) == 0 {
		return
	}

	n := 0
	for _, c := range components {
		if c != nil {
			components[n] = c
			n++
		}
	}
	if n == 0 {
		return
	}
	components = components[:n]

	s.managedLock.Lock()
	s.managedComponents = append(s.managedComponents, components...)
	s.managedLock.Unlock()
}

// AttachBackup registers backup submitters for overload handling.
func (s *SurgeProtector[T]) AttachBackup(backups ...types.Submitter[T]) {
	if len(backups) == 0 {
		return
	}

	n := 0
	for _, b := range backups {
		if b != nil {
			backups[n] = b
			n++
		}
	}
	if n == 0 {
		return
	}
	backups = backups[:n]

	s.backupLock.Lock()
	s.backupSystems = append(s.backupSystems, backups...)
	s.backupLock.Unlock()
}

// DetachBackups removes all backup systems.
func (s *SurgeProtector[T]) DetachBackups() {
	s.backupLock.Lock()
	backups := append([]types.Submitter[T](nil), s.backupSystems...)
	s.backupSystems = nil
	s.backupLock.Unlock()

	if len(backups) == 0 {
		return
	}

	s.notifySurgeProtectorDetachedBackups(backups...)
}

// ConnectResister attaches a resister queue for buffering.
func (s *SurgeProtector[T]) ConnectResister(r types.Resister[T]) {
	if r == nil {
		atomic.StoreInt32(&s.resisterEnabled, 0)
		s.resisterLock.Lock()
		s.resister = nil
		s.resisterLock.Unlock()
		return
	}

	s.resisterLock.Lock()
	s.resister = r
	s.resisterLock.Unlock()

	atomic.StoreInt32(&s.resisterEnabled, 1)
	s.notifySurgeProtectorConnectResister(r)
}

// ConnectLogger registers loggers for surge protector events.
func (s *SurgeProtector[T]) ConnectLogger(loggers ...types.Logger) {
	if len(loggers) == 0 {
		return
	}

	n := 0
	for _, l := range loggers {
		if l != nil {
			loggers[n] = l
			n++
		}
	}
	if n == 0 {
		return
	}
	loggers = loggers[:n]

	s.loggersLock.Lock()
	s.loggers = append(s.loggers, loggers...)
	s.loggersLock.Unlock()
}

// ConnectSensor registers sensors for surge protector events.
func (s *SurgeProtector[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	if len(sensors) == 0 {
		return
	}

	n := 0
	for _, sensor := range sensors {
		if sensor != nil {
			sensors[n] = sensor
			n++
		}
	}
	if n == 0 {
		return
	}
	sensors = sensors[:n]

	s.sensorLock.Lock()
	s.sensors = append(s.sensors, sensors...)
	s.sensorLock.Unlock()

	s.notifySurgeProtectorCreation(sensors)
}
