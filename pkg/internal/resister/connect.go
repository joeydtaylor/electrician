package resister

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectLogger registers loggers for resister events.
func (r *Resister[T]) ConnectLogger(loggers ...types.Logger) {
	if len(loggers) == 0 {
		return
	}

	n := 0
	for _, l := range loggers {
		if l != nil {
			loggers[n] = l
			n++
		}
	}
	if n == 0 {
		return
	}
	loggers = loggers[:n]

	r.loggersLock.Lock()
	r.loggers = append(r.loggers, loggers...)
	r.loggersLock.Unlock()
}

// ConnectSensor registers sensors for resister events.
func (r *Resister[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	if len(sensors) == 0 {
		return
	}

	n := 0
	for _, s := range sensors {
		if s != nil {
			sensors[n] = s
			n++
		}
	}
	if n == 0 {
		return
	}
	sensors = sensors[:n]

	r.sensorLock.Lock()
	r.sensors = append(r.sensors, sensors...)
	r.sensorLock.Unlock()
}
