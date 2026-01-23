package plug

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers emits a log event to all configured loggers.
func (p *Plug[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := p.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	msg := fmt.Sprintf(format, args...)
	for _, logger := range loggers {
		if logger == nil {
			continue
		}
		if logger.GetLevel() > level {
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

func (p *Plug[T]) snapshotLoggers() []types.Logger {
	p.loggersLock.Lock()
	loggers := append([]types.Logger(nil), p.loggers...)
	p.loggersLock.Unlock()
	return loggers
}
