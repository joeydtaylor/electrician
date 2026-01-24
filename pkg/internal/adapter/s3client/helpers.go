package s3client

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (a *S3Client[T]) snapshotLoggers() []types.Logger {
	a.loggersLock.Lock()
	defer a.loggersLock.Unlock()

	if len(a.loggers) == 0 {
		return nil
	}

	loggers := make([]types.Logger, len(a.loggers))
	copy(loggers, a.loggers)
	return loggers
}

func (a *S3Client[T]) snapshotSensors() []types.Sensor[T] {
	a.sensorLock.Lock()
	defer a.sensorLock.Unlock()

	if len(a.sensors) == 0 {
		return nil
	}

	sensors := make([]types.Sensor[T], len(a.sensors))
	copy(sensors, a.sensors)
	return sensors
}

func (a *S3Client[T]) fanIn(ctx context.Context, dst chan<- T, src <-chan T) {
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-src:
			if !ok {
				return
			}
			select {
			case dst <- v:
			case <-ctx.Done():
				return
			}
		}
	}
}
