package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers emits a log event to all configured loggers.
func (rr *ReceivingRelay[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := rr.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	for _, logger := range loggers {
		if logger == nil {
			continue
		}
		type levelChecker interface {
			IsLevelEnabled(types.LogLevel) bool
		}
		if lc, ok := logger.(levelChecker); ok && !lc.IsLevelEnabled(level) {
			continue
		}
		if logger.GetLevel() > level {
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

func (rr *ReceivingRelay[T]) logKV(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	fields := make([]interface{}, 0, len(keysAndValues)+2)
	fields = append(fields, "component", rr.componentMetadata)
	fields = append(fields, keysAndValues...)
	rr.NotifyLoggers(level, msg, fields...)
}
