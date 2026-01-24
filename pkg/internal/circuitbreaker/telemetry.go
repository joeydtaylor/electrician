package circuitbreaker

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// NotifyLoggers emits a log entry to all configured loggers.
func (cb *CircuitBreaker[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	loggers := cb.snapshotLoggers()
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

func (cb *CircuitBreaker[T]) notifyCircuitBreakerCreation(sensors []types.Sensor[T]) {
	if len(sensors) == 0 {
		return
	}
	for _, s := range sensors {
		if s == nil {
			continue
		}
		for _, meter := range s.GetMeters() {
			meter.IncrementCount(types.MetricCircuitBreakerCount)
		}
	}
}

func (cb *CircuitBreaker[T]) notifyAllow() {
	metadata := cb.snapshotMetadata()
	for _, s := range cb.snapshotSensors() {
		s.InvokeOnCircuitBreakerAllow(metadata)
	}
}

func (cb *CircuitBreaker[T]) notifyRecordError(time int64) {
	metadata := cb.snapshotMetadata()
	for _, s := range cb.snapshotSensors() {
		s.InvokeOnCircuitBreakerRecordError(metadata, time)
	}
}

func (cb *CircuitBreaker[T]) notifyTrip(time int64, resetTime int64) {
	metadata := cb.snapshotMetadata()
	for _, s := range cb.snapshotSensors() {
		s.InvokeOnCircuitBreakerTrip(metadata, time, resetTime)
	}
}

func (cb *CircuitBreaker[T]) notifyReset(time int64) {
	metadata := cb.snapshotMetadata()
	for _, s := range cb.snapshotSensors() {
		s.InvokeOnCircuitBreakerReset(metadata, time)
	}
}
