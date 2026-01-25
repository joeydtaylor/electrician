package websocket

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers logs a formatted message to all attached loggers.
func (s *serverAdapter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := s.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	msg := fmt.Sprintf(format, args...)
	for _, logger := range loggers {
		if logger == nil || logger.GetLevel() > level {
			continue
		}
		switch level {
		case types.DebugLevel:
			logger.Debug(msg)
		case types.InfoLevel:
			logger.Info(msg)
		case types.WarnLevel:
			logger.Warn(msg)
		case types.ErrorLevel:
			logger.Error(msg)
		case types.DPanicLevel:
			logger.DPanic(msg)
		case types.PanicLevel:
			logger.Panic(msg)
		case types.FatalLevel:
			logger.Fatal(msg)
		}
	}
}

func (s *serverAdapter[T]) snapshotLoggers() []types.Logger {
	s.loggersLock.Lock()
	defer s.loggersLock.Unlock()

	if len(s.loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(s.loggers))
	copy(loggers, s.loggers)
	return loggers
}

func (s *serverAdapter[T]) snapshotSensors() []types.Sensor[T] {
	s.sensorsLock.Lock()
	defer s.sensorsLock.Unlock()

	if len(s.sensors) == 0 {
		return nil
	}

	sensors := make([]types.Sensor[T], len(s.sensors))
	copy(sensors, s.sensors)
	return sensors
}
