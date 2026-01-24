package kafkaclient

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (a *KafkaClient[T]) snapshotLoggers() []types.Logger {
	a.loggersLock.Lock()
	defer a.loggersLock.Unlock()

	if len(a.loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(a.loggers))
	copy(loggers, a.loggers)
	return loggers
}

func (a *KafkaClient[T]) snapshotSensors() []types.Sensor[T] {
	a.sensorLock.Lock()
	defer a.sensorLock.Unlock()

	if len(a.sensors) == 0 {
		return nil
	}

	sensors := make([]types.Sensor[T], len(a.sensors))
	copy(sensors, a.sensors)
	return sensors
}

func (a *KafkaClient[T]) fanIn(ctx context.Context, dst chan<- T, src <-chan T) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-src:
			if !ok {
				return
			}
			dst <- v
		}
	}
}
