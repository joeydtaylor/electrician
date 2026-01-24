package kafkaclient

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers sends a formatted message to all attached loggers.
func (a *KafkaClient[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := a.snapshotLoggers()
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
