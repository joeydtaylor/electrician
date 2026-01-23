package wire

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (w *Wire[T]) hasLoggers() bool {
	return atomic.LoadInt32(&w.loggerCount) != 0
}

func (w *Wire[T]) hasSensors() bool {
	return atomic.LoadInt32(&w.sensorCount) != 0
}

// NotifyLoggers emits a log event to all configured loggers.
func (w *Wire[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	if atomic.LoadInt32(&w.loggerCount) == 0 {
		return
	}

	loggers := w.loggers
	if len(loggers) == 0 {
		return
	}

	type levelChecker interface {
		IsLevelEnabled(types.LogLevel) bool
	}

	switch level {
	case types.DebugLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Debug(msg, keysAndValues...)
		}
	case types.InfoLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Info(msg, keysAndValues...)
		}
	case types.WarnLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Warn(msg, keysAndValues...)
		}
	case types.ErrorLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Error(msg, keysAndValues...)
		}
	case types.DPanicLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.DPanic(msg, keysAndValues...)
		}
	case types.PanicLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Panic(msg, keysAndValues...)
		}
	case types.FatalLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Fatal(msg, keysAndValues...)
		}
	}
}
