package websocket

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers logs a formatted message to all attached loggers.
func (s *serverAdapter[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := s.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	for _, logger := range loggers {
		if logger == nil || logger.GetLevel() > level {
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
