package forwardrelay

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers emits a log event to all configured loggers.
func (fr *ForwardRelay[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := fr.snapshotLoggers()
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

func (fr *ForwardRelay[T]) snapshotLoggers() []types.Logger {
	fr.loggersLock.Lock()
	defer fr.loggersLock.Unlock()

	if len(fr.Loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(fr.Loggers))
	copy(loggers, fr.Loggers)
	return loggers
}
