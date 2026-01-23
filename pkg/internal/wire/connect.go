package wire

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ConnectCircuitBreaker attaches a circuit breaker to the wire.
// Panics if called after Start.
func (w *Wire[T]) ConnectCircuitBreaker(cb types.CircuitBreaker[T]) {
	w.requireNotStarted("ConnectCircuitBreaker")

	w.cbLock.Lock()
	w.CircuitBreaker = cb
	w.cbLock.Unlock()

	if cb == nil {
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectCircuitBreaker",
			"component", w.componentMetadata,
			"event", "ConnectCircuitBreaker",
			"result", "SUCCESS",
			"circuitBreaker", nil,
		)
		return
	}

	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectCircuitBreaker",
		"component", w.componentMetadata,
		"event", "ConnectCircuitBreaker",
		"result", "SUCCESS",
		"circuitBreakerComponentMetadata", cb.GetComponentMetadata(),
		"circuitBreakerAddress", cb,
	)
}

// ConnectGenerator registers generators for the wire.
// Panics if called after Start.
func (w *Wire[T]) ConnectGenerator(generators ...types.Generator[T]) {
	w.requireNotStarted("ConnectGenerator")

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

	w.generators = append(w.generators, generators...)
	for _, g := range generators {
		g.ConnectToComponent(w)
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectGenerator",
			"component", w.componentMetadata,
			"event", "ConnectGenerator",
			"result", "SUCCESS",
			"generatorComponentMetadata", g.GetComponentMetadata(),
			"generatorAddress", g,
		)
	}
}

// ConnectLogger registers loggers for the wire.
// Panics if called after Start.
func (w *Wire[T]) ConnectLogger(loggers ...types.Logger) {
	w.requireNotStarted("ConnectLogger")

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

	w.loggersLock.Lock()
	w.loggers = append(w.loggers, loggers...)
	w.loggersLock.Unlock()

	atomic.AddInt32(&w.loggerCount, int32(n))

	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectLogger",
		"component", w.componentMetadata,
		"event", "ConnectLogger",
		"result", "SUCCESS",
		"count", n,
	)
}

// ConnectSensor registers sensors for the wire.
// Panics if called after Start.
func (w *Wire[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	w.requireNotStarted("ConnectSensor")

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

	w.sensorLock.Lock()
	w.sensors = append(w.sensors, sensors...)
	w.sensorLock.Unlock()

	atomic.AddInt32(&w.sensorCount, int32(n))

	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectSensor",
		"component", w.componentMetadata,
		"event", "ConnectSensor",
		"result", "SUCCESS",
		"count", n,
	)
}

// ConnectSurgeProtector attaches a surge protector to the wire.
// Panics if called after Start.
func (w *Wire[T]) ConnectSurgeProtector(protector types.SurgeProtector[T]) {
	w.requireNotStarted("ConnectSurgeProtector")

	w.surgeProtector = protector
	if protector == nil {
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectSurgeProtector",
			"component", w.componentMetadata,
			"event", "ConnectSurgeProtector",
			"result", "SUCCESS",
			"surgeProtector", nil,
		)
		return
	}

	_, _, maxRetryAttempts, _ := protector.GetRateLimit()
	w.queueRetryMaxAttempt = maxRetryAttempts

	protector.ConnectComponent(w)

	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectSurgeProtector",
		"component", w.componentMetadata,
		"event", "ConnectSurgeProtector",
		"result", "SUCCESS",
		"surgeProtectorComponentMetadata", protector.GetComponentMetadata(),
		"surgeProtectorAddress", protector,
	)
}

// ConnectTransformer registers transformers for the wire.
// Panics if called after Start.
func (w *Wire[T]) ConnectTransformer(transformers ...types.Transformer[T]) {
	w.requireNotStarted("ConnectTransformer")

	if len(transformers) == 0 {
		return
	}
	if w.transformerFactory != nil {
		panic("wire: cannot use ConnectTransformer when TransformerFactory is set; wrap transforms inside the factory closure instead")
	}

	n := 0
	for _, t := range transformers {
		if t != nil {
			transformers[n] = t
			n++
		}
	}
	if n == 0 {
		return
	}
	transformers = transformers[:n]

	w.transformations = append(w.transformations, transformers...)
	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectTransformer",
		"component", w.componentMetadata,
		"event", "ConnectTransformer",
		"result", "SUCCESS",
		"count", n,
	)
}
