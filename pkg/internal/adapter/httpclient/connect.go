package httpclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectSensor attaches sensors to observe HTTP events.
func (hp *HTTPClientAdapter[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	if len(sensor) == 0 {
		return
	}

	hp.sensorsLock.Lock()
	hp.sensors = append(hp.sensors, sensor...)
	hp.sensorsLock.Unlock()
}

// ConnectLogger attaches loggers for HTTP adapter events.
func (hp *HTTPClientAdapter[T]) ConnectLogger(l ...types.Logger) {
	if len(l) == 0 {
		return
	}

	hp.loggersLock.Lock()
	hp.loggers = append(hp.loggers, l...)
	hp.loggersLock.Unlock()
}
