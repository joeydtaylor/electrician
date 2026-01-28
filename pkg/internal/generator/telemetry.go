package generator

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers emits a log event to all configured loggers.
func (g *Generator[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := g.snapshotLoggers()
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

func (g *Generator[T]) notifyStart() {
	for _, sensor := range g.snapshotSensors() {
		sensor.InvokeOnStart(g.componentMetadata)
	}
}

func (g *Generator[T]) notifyStop() {
	for _, sensor := range g.snapshotSensors() {
		sensor.InvokeOnStop(g.componentMetadata)
	}
}

func (g *Generator[T]) notifyRestart() {
	for _, sensor := range g.snapshotSensors() {
		sensor.InvokeOnRestart(g.componentMetadata)
	}
}

func (g *Generator[T]) snapshotLoggers() []types.Logger {
	g.loggersLock.Lock()
	loggers := append([]types.Logger(nil), g.loggers...)
	g.loggersLock.Unlock()
	return loggers
}

func (g *Generator[T]) snapshotSensors() []types.Sensor[T] {
	return append([]types.Sensor[T](nil), g.sensors...)
}
