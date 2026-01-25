package httpserver

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectLogger attaches logger(s).
func (h *httpServerAdapter[T]) ConnectLogger(loggers ...types.Logger) {
	h.requireNotFrozen("ConnectLogger")

	if len(loggers) == 0 {
		return
	}

	n := 0
	for _, logger := range loggers {
		if logger != nil {
			loggers[n] = logger
			n++
		}
	}
	if n == 0 {
		return
	}
	loggers = loggers[:n]

	h.loggersLock.Lock()
	h.loggers = append(h.loggers, loggers...)
	h.loggersLock.Unlock()
}

// ConnectSensor attaches sensor(s).
func (h *httpServerAdapter[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	h.requireNotFrozen("ConnectSensor")

	if len(sensors) == 0 {
		return
	}

	n := 0
	for _, sensor := range sensors {
		if sensor != nil {
			sensors[n] = sensor
			n++
		}
	}
	if n == 0 {
		return
	}
	sensors = sensors[:n]

	h.sensorsLock.Lock()
	h.sensors = append(h.sensors, sensors...)
	h.sensorsLock.Unlock()
}
