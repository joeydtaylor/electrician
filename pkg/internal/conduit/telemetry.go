package conduit

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers emits a log entry to configured loggers.
func (c *Conduit[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := c.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	msg := format
	if len(args) != 0 {
		msg = fmt.Sprintf(format, args...)
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

func (c *Conduit[T]) snapshotLoggers() []types.Logger {
	c.loggersLock.Lock()
	loggers := append([]types.Logger(nil), c.loggers...)
	c.loggersLock.Unlock()
	return loggers
}
