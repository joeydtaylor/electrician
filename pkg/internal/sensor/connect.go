package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectLogger registers loggers for sensor output.
func (s *Sensor[T]) ConnectLogger(loggers ...types.Logger) {
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

	s.loggersLock.Lock()
	s.loggers = append(s.loggers, loggers...)
	s.loggersLock.Unlock()
}

// ConnectMeter registers meters for sensor metrics.
func (s *Sensor[T]) ConnectMeter(meter ...types.Meter[T]) {
	if len(meter) == 0 {
		return
	}

	n := 0
	for _, m := range meter {
		if m != nil {
			meter[n] = m
			n++
		}
	}
	if n == 0 {
		return
	}
	meter = meter[:n]

	s.metersLock.Lock()
	s.meters = append(s.meters, meter...)
	s.metersLock.Unlock()
}
