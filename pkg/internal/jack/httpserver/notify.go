package httpserver

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers logs a formatted message to all attached loggers.
func (h *httpServerAdapter[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := h.snapshotLoggers()
	if len(loggers) == 0 {
		return
	}

	for _, logger := range loggers {
		if logger == nil || logger.GetLevel() > level {
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

func (h *httpServerAdapter[T]) snapshotLoggers() []types.Logger {
	h.loggersLock.Lock()
	defer h.loggersLock.Unlock()

	if len(h.loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(h.loggers))
	copy(loggers, h.loggers)
	return loggers
}

func (h *httpServerAdapter[T]) snapshotSensors() []types.Sensor[T] {
	h.sensorsLock.Lock()
	defer h.sensorsLock.Unlock()

	if len(h.sensors) == 0 {
		return nil
	}

	sensors := make([]types.Sensor[T], len(h.sensors))
	copy(sensors, h.sensors)
	return sensors
}
