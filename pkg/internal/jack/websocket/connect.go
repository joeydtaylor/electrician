package websocket

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *serverAdapter[T]) ConnectLogger(loggers ...types.Logger) {
	s.requireNotFrozen("ConnectLogger")
	s.loggersLock.Lock()
	defer s.loggersLock.Unlock()
	s.loggers = append(s.loggers, loggers...)
}

func (s *serverAdapter[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	s.requireNotFrozen("ConnectSensor")
	s.sensorsLock.Lock()
	defer s.sensorsLock.Unlock()
	s.sensors = append(s.sensors, sensors...)
}

func (s *serverAdapter[T]) ConnectOutput(wires ...types.Wire[T]) {
	s.requireNotFrozen("ConnectOutput")
	s.configLock.Lock()
	defer s.configLock.Unlock()
	s.outputWires = append(s.outputWires, wires...)
}
