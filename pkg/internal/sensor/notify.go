package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers emits a log entry to configured loggers.
func (s *Sensor[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
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

func (s *Sensor[T]) snapshotLoggers() []types.Logger {
	s.loggersLock.Lock()
	loggers := append([]types.Logger(nil), s.loggers...)
	s.loggersLock.Unlock()
	return loggers
}
