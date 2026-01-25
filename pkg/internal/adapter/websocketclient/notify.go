package websocketclient

import (
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NotifyLoggers logs a formatted message to all attached loggers.
func (c *WebSocketClientAdapter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	loggers := c.snapshotLoggers()
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

func (c *WebSocketClientAdapter[T]) snapshotLoggers() []types.Logger {
	c.loggersLock.Lock()
	defer c.loggersLock.Unlock()

	if len(c.loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(c.loggers))
	copy(loggers, c.loggers)
	return loggers
}

func (c *WebSocketClientAdapter[T]) snapshotSensors() []types.Sensor[T] {
	c.sensorsLock.Lock()
	defer c.sensorsLock.Unlock()

	if len(c.sensors) == 0 {
		return nil
	}

	sensors := make([]types.Sensor[T], len(c.sensors))
	copy(sensors, c.sensors)
	return sensors
}
