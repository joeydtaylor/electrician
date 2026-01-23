package generator

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectCircuitBreaker attaches a circuit breaker to the generator.
// Panics if called after Start.
func (g *Generator[T]) ConnectCircuitBreaker(cb types.CircuitBreaker[T]) {
	g.requireNotStarted("ConnectCircuitBreaker")

	g.cbLock.Lock()
	g.CircuitBreaker = cb
	g.cbLock.Unlock()

	g.NotifyLoggers(types.DebugLevel, "ConnectCircuitBreaker: connected circuit breaker")
}

// ConnectLogger registers loggers for the generator.
// Panics if called after Start.
func (g *Generator[T]) ConnectLogger(loggers ...types.Logger) {
	g.requireNotStarted("ConnectLogger")

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

	g.loggersLock.Lock()
	g.loggers = append(g.loggers, loggers...)
	g.loggersLock.Unlock()
}

// ConnectSensor registers sensors for the generator.
// Panics if called after Start.
func (g *Generator[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	g.requireNotStarted("ConnectSensor")

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

	g.sensors = append(g.sensors, sensors...)
	g.attachSensorsToPlugs(g.plugs)

	for _, s := range sensors {
		g.NotifyLoggers(types.DebugLevel, "ConnectSensor: connected sensor %s", s.GetComponentMetadata())
	}
}

// ConnectPlug registers plugs for the generator.
// Panics if called after Start.
func (g *Generator[T]) ConnectPlug(plugs ...types.Plug[T]) {
	g.requireNotStarted("ConnectPlug")

	if len(plugs) == 0 {
		return
	}

	n := 0
	for _, p := range plugs {
		if p != nil {
			plugs[n] = p
			n++
		}
	}
	if n == 0 {
		return
	}
	plugs = plugs[:n]

	g.configLock.Lock()
	g.plugs = append(g.plugs, plugs...)
	g.configLock.Unlock()

	g.attachSensorsToPlugs(plugs)
}

// ConnectToComponent registers downstream submitters.
// Panics if called after Start.
func (g *Generator[T]) ConnectToComponent(submitters ...types.Submitter[T]) {
	g.requireNotStarted("ConnectToComponent")

	if len(submitters) == 0 {
		return
	}

	n := 0
	for _, s := range submitters {
		if s != nil {
			submitters[n] = s
			n++
		}
	}
	if n == 0 {
		return
	}
	submitters = submitters[:n]

	g.configLock.Lock()
	g.connectedComponents = append(g.connectedComponents, submitters...)
	g.configLock.Unlock()
}

func (g *Generator[T]) attachSensorsToPlugs(plugs []types.Plug[T]) {
	if len(plugs) == 0 || len(g.sensors) == 0 {
		return
	}
	for _, p := range plugs {
		if p == nil {
			continue
		}
		p.ConnectSensor(g.sensors...)
	}
}
