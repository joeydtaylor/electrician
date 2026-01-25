package conduit

import (
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (c *Conduit[T]) ensureDefaults() {
	if c.MaxBufferSize <= 0 {
		c.MaxBufferSize = 1000
	}
	if c.MaxConcurrency <= 0 {
		c.MaxConcurrency = 1
	}
	if c.InputChan == nil {
		c.InputChan = make(chan T, c.MaxBufferSize)
	}
	if c.OutputChan == nil {
		c.OutputChan = make(chan T, c.MaxBufferSize)
	}
}

func (c *Conduit[T]) snapshotWires() []types.Wire[T] {
	c.configLock.Lock()
	wires := append([]types.Wire[T](nil), c.wires...)
	c.configLock.Unlock()
	return wires
}

func (c *Conduit[T]) snapshotSensors() []types.Sensor[T] {
	c.configLock.Lock()
	sensors := append([]types.Sensor[T](nil), c.sensors...)
	c.configLock.Unlock()
	return sensors
}

func (c *Conduit[T]) snapshotGenerators() []types.Generator[T] {
	c.configLock.Lock()
	generators := append([]types.Generator[T](nil), c.generators...)
	c.configLock.Unlock()
	return generators
}

func (c *Conduit[T]) snapshotCircuitBreaker() types.CircuitBreaker[T] {
	c.cbLock.Lock()
	cb := c.CircuitBreaker
	c.cbLock.Unlock()
	return cb
}

func (c *Conduit[T]) snapshotSurgeProtector() types.SurgeProtector[T] {
	c.surgeProtectorLock.Lock()
	sp := c.surgeProtector
	c.surgeProtectorLock.Unlock()
	return sp
}

func (c *Conduit[T]) attachWire(w types.Wire[T]) {
	if w == nil {
		return
	}

	if loggers := c.snapshotLoggers(); len(loggers) != 0 {
		w.ConnectLogger(loggers...)
	}
	if sensors := c.snapshotSensors(); len(sensors) != 0 {
		w.ConnectSensor(sensors...)
	}
	if cb := c.snapshotCircuitBreaker(); cb != nil {
		w.ConnectCircuitBreaker(cb)
	}
	if gens := c.snapshotGenerators(); len(gens) != 0 {
		w.ConnectGenerator(gens...)
	}
	if sp := c.snapshotSurgeProtector(); sp != nil {
		w.ConnectSurgeProtector(sp)
	}

	w.SetConcurrencyControl(c.MaxBufferSize, c.MaxConcurrency)
}

func (c *Conduit[T]) chainWires() {
	wires := c.snapshotWires()
	var prev types.Wire[T]

	for _, w := range wires {
		if w == nil {
			continue
		}
		if prev != nil {
			w.SetInputChannel(prev.GetOutputChannel())
		}
		prev = w
	}

	if prev != nil && c.OutputChan != nil {
		prev.SetOutputChannel(c.OutputChan)
	}
}

func (c *Conduit[T]) lastWire() types.Wire[T] {
	wires := c.snapshotWires()
	for i := len(wires) - 1; i >= 0; i-- {
		if wires[i] != nil {
			return wires[i]
		}
	}
	return nil
}
