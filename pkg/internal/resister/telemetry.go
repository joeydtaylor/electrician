package resister

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers emits a log entry to all configured loggers.
func (r *Resister[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := r.snapshotLoggers()
	if len(loggers) == 0 {
		return
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

func (r *Resister[T]) snapshotLoggers() []types.Logger {
	r.loggersLock.Lock()
	loggers := append([]types.Logger(nil), r.loggers...)
	r.loggersLock.Unlock()
	return loggers
}

func (r *Resister[T]) snapshotSensors() []types.Sensor[T] {
	r.sensorLock.Lock()
	sensors := append([]types.Sensor[T](nil), r.sensors...)
	r.sensorLock.Unlock()
	return sensors
}

func (r *Resister[T]) snapshotMetadata() types.ComponentMetadata {
	r.metadataLock.Lock()
	metadata := r.componentMetadata
	r.metadataLock.Unlock()
	return metadata
}

func (r *Resister[T]) notifyResisterQueued(elem T) {
	metadata := r.snapshotMetadata()
	for _, sensor := range r.snapshotSensors() {
		sensor.InvokeOnResisterQueued(metadata, elem)
	}
}

func (r *Resister[T]) notifyResisterRequeued(elem T) {
	metadata := r.snapshotMetadata()
	for _, sensor := range r.snapshotSensors() {
		sensor.InvokeOnResisterRequeued(metadata, elem)
	}
}

func (r *Resister[T]) notifyResisterDequeued(elem T) {
	metadata := r.snapshotMetadata()
	for _, sensor := range r.snapshotSensors() {
		sensor.InvokeOnResisterDequeued(metadata, elem)
	}
}

func (r *Resister[T]) notifyResisterEmpty() {
	metadata := r.snapshotMetadata()
	for _, sensor := range r.snapshotSensors() {
		sensor.InvokeOnResisterEmpty(metadata)
	}
}
