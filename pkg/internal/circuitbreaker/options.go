package circuitbreaker

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithNeutralWire adds a neutral wire to the circuit breaker.
func WithNeutralWire[T any](neutralWire types.Wire[T]) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.ConnectNeutralWire(neutralWire)
	}
}

// WithSensor registers sensors for circuit breaker events.
func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.ConnectSensor(sensor...)
	}
}

// WithLogger attaches a logger to the circuit breaker.
func WithLogger[T any](logger types.Logger) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.ConnectLogger(logger)
	}
}

// WithComponentMetadata configures component metadata overrides.
func WithComponentMetadata[T any](name string, id string) types.Option[types.CircuitBreaker[T]] {
	return func(cb types.CircuitBreaker[T]) {
		cb.SetComponentMetadata(name, id)
	}
}
