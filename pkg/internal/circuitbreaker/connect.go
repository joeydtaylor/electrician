package circuitbreaker

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectNeutralWire attaches wires that can receive work when the breaker is open.
func (cb *CircuitBreaker[T]) ConnectNeutralWire(wires ...types.Wire[T]) {
	if len(wires) == 0 {
		return
	}

	n := 0
	for _, w := range wires {
		if w != nil {
			wires[n] = w
			n++
		}
	}
	if n == 0 {
		return
	}
	wires = wires[:n]

	cb.configLock.Lock()
	cb.neutralWires = append(cb.neutralWires, wires...)
	cb.configLock.Unlock()

	component := cb.snapshotMetadata()
	for _, w := range wires {
		cb.NotifyLoggers(types.DebugLevel, "ConnectNeutralWire: connected neutral wire", "component", component, "wire", w.GetComponentMetadata())
	}
}

// ConnectSensor registers sensors for circuit breaker events.
func (cb *CircuitBreaker[T]) ConnectSensor(sensors ...types.Sensor[T]) {
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

	cb.configLock.Lock()
	cb.sensors = append(cb.sensors, sensors...)
	cb.configLock.Unlock()

	cb.notifyCircuitBreakerCreation(sensors)

	component := cb.snapshotMetadata()
	for _, s := range sensors {
		cb.NotifyLoggers(types.DebugLevel, "ConnectSensor: connected sensor", "component", component, "sensor", s.GetComponentMetadata())
	}
}

// ConnectLogger attaches a logger to the circuit breaker.
func (cb *CircuitBreaker[T]) ConnectLogger(logger types.Logger) {
	if logger == nil {
		return
	}

	cb.configLock.Lock()
	cb.loggers = append(cb.loggers, logger)
	cb.configLock.Unlock()

	cb.NotifyLoggers(types.DebugLevel, "ConnectLogger: connected logger", "component", cb.snapshotMetadata(), "logger", logger)
}
