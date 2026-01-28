package plug

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers emits a log event to all configured loggers.
func (p *Plug[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := p.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	for _, logger := range loggers {
		if logger == nil {
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

func (p *Plug[T]) snapshotLoggers() []types.Logger {
	p.loggersLock.Lock()
	loggers := append([]types.Logger(nil), p.loggers...)
	p.loggersLock.Unlock()
	return loggers
}
