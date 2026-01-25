package httpserver

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers logs a formatted message to all attached loggers.
func (h *httpServerAdapter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := h.snapshotLoggers()
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
