package websocketclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers logs a formatted message to all attached loggers.
func (c *WebSocketClientAdapter[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := c.snapshotLoggers()
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
