package meter

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

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
func (m *Meter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := m.snapshotLoggers()
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
