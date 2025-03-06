// Package circuitbreaker provides a generic circuit breaker implementation for handling failures
// in a distributed system or any operation that can fail. It prevents a system from performing
// operations that are likely to fail by opening the circuit when a threshold of failures has been reached.
package circuitbreaker

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithNeutralWire returns an option for configuring a CircuitBreaker with a specific ground wire.
// A ground wire is used to manage or buffer operations when the circuit breaker is in the open state,
// allowing these operations to be reprocessed later. This function wraps the logic for connecting
// a ground wire to the circuit breaker, ensuring that operations can be queued during periods of failure.
//
// Parameters:
// neutralWire - The ground wire to connect to the circuit breaker, used to buffer and later execute operations.
//
// Returns:
// An Option function that, when applied, configures the circuit breaker with the specified ground wire.
func WithNeutralWire[T any](neutralWire types.Wire[T]) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.ConnectNeutralWire(neutralWire)
	}
}

func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.ConnectSensor(sensor...)
	}
}

// WithLogger returns an option for attaching a logger to a CircuitBreaker.
// Loggers are crucial for monitoring the behavior and state changes of the circuit breaker,
// providing insights into operations and failures. This function encapsulates the process of
// connecting a logger to a circuit breaker, enhancing its observability and debuggability.
//
// Parameters:
// logger - The logger to attach, which will receive log entries about state changes and error events.
//
// Returns:
// An Option function that, when applied, connects the specified logger to the circuit breaker.
func WithLogger[T any](logger types.Logger) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.ConnectLogger(logger)
	}
}

// WithComponentMetadata configures a CircuitBreaker component with custom metadata, including a name and an identifier.
// This function provides an option to set these metadata properties, which can be used for identification,
// logging, or other purposes where metadata is needed for a CircuitBreaker. It uses the SetComponentMetadata method
// internally to apply these settings. If the CircuitBreaker's configuration is frozen (indicating that the component
// has started operation and its configuration should no longer be changed), attempting to set metadata
// will result in a panic. This ensures the integrity of component configurations during runtime.
//
// Parameters:
//   - name: The name to set for the CircuitBreaker component, used for identification and logging.
//   - id: The unique identifier to set for the CircuitBreaker component, used for unique identification across systems.
//
// Returns:
//
//	A function conforming to types.Option[types.CircuitBreaker[T]], which when called with a CircuitBreaker component,
//	sets the specified name and id in the component's metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.SetComponentMetadata(name, id)
	}
}
