package receivingrelay

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers emits a log event to all configured loggers.
func (rr *ReceivingRelay[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := rr.snapshotLoggers()
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

func (rr *ReceivingRelay[T]) snapshotLoggers() []types.Logger {
	rr.loggersLock.Lock()
	defer rr.loggersLock.Unlock()

	if len(rr.Loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(rr.Loggers))
	copy(loggers, rr.Loggers)
	return loggers
}
