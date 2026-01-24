package circuitbreaker

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (cb *CircuitBreaker[T]) snapshotLoggers() []types.Logger {
	cb.configLock.Lock()
	loggers := append([]types.Logger(nil), cb.loggers...)
	cb.configLock.Unlock()
	return loggers
}

func (cb *CircuitBreaker[T]) snapshotSensors() []types.Sensor[T] {
	cb.configLock.Lock()
	sensors := append([]types.Sensor[T](nil), cb.sensors...)
	cb.configLock.Unlock()
	return sensors
}

func (cb *CircuitBreaker[T]) snapshotMetadata() types.ComponentMetadata {
	cb.configLock.Lock()
	metadata := cb.componentMetadata
	cb.configLock.Unlock()
	return metadata
}

func (cb *CircuitBreaker[T]) signalReset() {
	select {
	case cb.resetNotifyChan <- struct{}{}:
	default:
	}
}
