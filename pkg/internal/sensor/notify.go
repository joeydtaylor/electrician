package sensor

import (
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers sends a formatted log message to all attached loggers, enhancing traceability and debugging.
//
// Parameters:
//   - level: The log level of the message.
//   - format: The format string for the log message.
//   - args: Additional arguments to be formatted into the message.
//
// NotifyLoggers notifies loggers with the specified log level, format, and arguments.
// This function notifies registered loggers with the specified log level, format, and arguments,
// allowing them to log relevant events and information during data processing.
// Parameters:
//   - level: The log level of the message.
//   - format: The format string for the log message.
//   - args: The arguments to be formatted into the log message.
func (s *Sensor[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	if s.loggers != nil {
		for _, logger := range s.loggers {
			if logger == nil {
				continue
			}
			// Ensure we only acquire the lock once per logger to avoid deadlock or excessive locking overhead
			s.loggersLock.Lock()
			// Check if the logger supports the IsLevelEnabled method dynamically
			type levelChecker interface {
				IsLevelEnabled(types.LogLevel) bool
			}
			if lc, ok := logger.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				s.loggersLock.Unlock()
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
			s.loggersLock.Unlock()
		}
	}
}
