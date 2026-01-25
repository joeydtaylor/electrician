package conduit

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectSurgeProtector attaches a surge protector to the conduit.
func (c *Conduit[T]) ConnectSurgeProtector(protector types.SurgeProtector[T]) {
	c.requireNotStarted("ConnectSurgeProtector")

	c.surgeProtectorLock.Lock()
	c.surgeProtector = protector
	c.surgeProtectorLock.Unlock()

	for _, w := range c.snapshotWires() {
		if w == nil {
			continue
		}
		w.ConnectSurgeProtector(protector)
	}
}

// ConnectCircuitBreaker attaches a circuit breaker to the conduit.
func (c *Conduit[T]) ConnectCircuitBreaker(cb types.CircuitBreaker[T]) {
	c.requireNotStarted("ConnectCircuitBreaker")

	c.cbLock.Lock()
	c.CircuitBreaker = cb
	c.cbLock.Unlock()

	for _, w := range c.snapshotWires() {
		if w == nil {
			continue
		}
		w.ConnectCircuitBreaker(cb)
	}
}

// ConnectConduit links a downstream conduit.
func (c *Conduit[T]) ConnectConduit(next types.Conduit[T]) {
	c.requireNotStarted("ConnectConduit")

	c.configLock.Lock()
	c.NextConduit = next
	c.configLock.Unlock()
}

// ConnectGenerator registers generators for the conduit and existing wires.
func (c *Conduit[T]) ConnectGenerator(generators ...types.Generator[T]) {
	c.requireNotStarted("ConnectGenerator")

	if len(generators) == 0 {
		return
	}

	n := 0
	for _, g := range generators {
		if g != nil {
			generators[n] = g
			n++
		}
	}
	if n == 0 {
		return
	}
	generators = generators[:n]

	c.configLock.Lock()
	c.generators = append(c.generators, generators...)
	c.configLock.Unlock()

	for _, w := range c.snapshotWires() {
		if w == nil {
			continue
		}
		w.ConnectGenerator(generators...)
	}
}

// ConnectLogger registers loggers for conduit events and existing wires.
func (c *Conduit[T]) ConnectLogger(loggers ...types.Logger) {
	c.requireNotStarted("ConnectLogger")

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

	c.loggersLock.Lock()
	c.loggers = append(c.loggers, loggers...)
	c.loggersLock.Unlock()

	for _, w := range c.snapshotWires() {
		if w == nil {
			continue
		}
		w.ConnectLogger(loggers...)
	}
}

// ConnectSensor registers sensors for conduit events and existing wires.
func (c *Conduit[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	c.requireNotStarted("ConnectSensor")

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

	c.configLock.Lock()
	c.sensors = append(c.sensors, sensors...)
	c.configLock.Unlock()

	for _, w := range c.snapshotWires() {
		if w == nil {
			continue
		}
		w.ConnectSensor(sensors...)
	}
}

// ConnectWire appends wires to the conduit pipeline.
func (c *Conduit[T]) ConnectWire(newWires ...types.Wire[T]) {
	c.requireNotStarted("ConnectWire")

	if len(newWires) == 0 {
		return
	}

	n := 0
	for _, w := range newWires {
		if w != nil {
			newWires[n] = w
			n++
		}
	}
	if n == 0 {
		return
	}
	newWires = newWires[:n]

	c.configLock.Lock()
	c.wires = append(c.wires, newWires...)
	c.configLock.Unlock()

	for _, w := range newWires {
		c.attachWire(w)
	}

	c.chainWires()
}
