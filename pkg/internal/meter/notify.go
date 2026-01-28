package meter

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectLogger attaches loggers to the meter.
func (m *Meter[T]) ConnectLogger(loggers ...types.Logger) {
	if len(loggers) == 0 {
		return
	}
	m.loggersMu.Lock()
	m.loggers = append(m.loggers, loggers...)
	m.loggersMu.Unlock()
}

// NotifyLoggers emits a log event to all configured loggers.
func (m *Meter[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := m.snapshotLoggers()
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

func (m *Meter[T]) snapshotLoggers() []types.Logger {
	m.loggersMu.Lock()
	defer m.loggersMu.Unlock()
	if len(m.loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(m.loggers))
	copy(loggers, m.loggers)
	return loggers
}
