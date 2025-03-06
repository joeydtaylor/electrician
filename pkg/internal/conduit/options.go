// Package conduit provides a conduit component that facilitates the connection and processing of data
// between various components in a system. It allows for the management of multiple wires, circuit breakers,
// generators, loggers, and sensors to create a flexible and customizable data processing pipeline.
package conduit

import (
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithCircuitBreaker sets the circuit breaker for the conduit.
//
// This option connects a circuit breaker to the conduit, enabling control over data flow.
// Parameters:
//   - cb: The circuit breaker to be connected.
func WithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.ConnectCircuitBreaker(cb)
	}
}

// WithConcurrencyControl sets the concurrency control parasensors for the conduit.
//
// This option configures the buffer size and maximum concurrent routines for the conduit's wires.
// Parameters:
//   - bufferSize: The buffer size for each wire's channel.
//   - maxRoutines: The maximum number of concurrent routines for each wire.
func WithConcurrencyControl[T any](bufferSize int, maxRoutines int) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.SetConcurrencyControl(bufferSize, maxRoutines)
	}
}

// WithPlug sets the generator function for the conduit.
//
// This option connects one or more generator functions to the conduit component,
// enabling it to generate data and feed it into the processing pipeline.
// Parameters:
//   - generator: The generator function(s) to be connected.
func WithGenerator[T any](generator types.Generator[T]) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.ConnectGenerator(generator)
	}
}

// WithPlug sets the generator function for the conduit.
//
// This option connects one or more generator functions to the conduit component,
// enabling it to generate data and feed it into the processing pipeline.
// Parameters:
//   - generator: The generator function(s) to be connected.
func WithSurgeProtector[T any](sp types.SurgeProtector[T]) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.ConnectSurgeProtector(sp)
	}
}

// WithLogger adds one or more loggers to the conduit.
//
// This option connects a logger to the conduit component, allowing it to
// log relevant events and information during data processing.
// Parameters:
//   - logger: The logger to be connected.
func WithLogger[T any](logger ...types.Logger) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.ConnectLogger(logger...)
	}
}

// WithSensor adds a sensor to the conduit for monitoring performance and events.
//
// This option connects a sensor to the conduit component, enabling it to
// measure and monitor the flow of data through the processing pipeline.
// Parameters:
//   - sensor: The sensor to be connected.
func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.ConnectSensor(sensor...)
	}
}

// WithWire adds one or more wires to the conduit.
//
// This option connects new wires to the conduit, establishing the processing chain.
// Parameters:
//   - wires: The new wires to be connected.
func WithWire[T any](wires ...types.Wire[T]) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.ConnectWire(wires...)
	}
}

// WithComponentMetadata configures a Conduit component with custom metadata, including a name and an identifier.
//
// This option sets metadata properties, such as name and ID, which are used for identification,
// logging, or other purposes within a Conduit. It applies these settings using the SetComponentMetadata method.
// If the Conduit's configuration is frozen (indicating that the component has started operation and its configuration
// should no longer be changed), attempting to set metadata will result in a panic. This ensures the integrity of
// component configurations during runtime.
//
// Parameters:
//   - name: The name to set for the Conduit component, used for identification and logging.
//   - id: The unique identifier to set for the Conduit component, used for unique identification across systems.
//
// Returns:
//
// A function conforming to types.Option[types.Conduit[T]], which when called with a Conduit component,
// sets the specified name and ID in the component's metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.Conduit[T]] {
	return func(c types.Conduit[T]) {
		c.SetComponentMetadata(name, id)
	}
}
