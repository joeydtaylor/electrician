package generator

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithPlug registers plugs for the generator.
func WithPlug[T any](p ...types.Plug[T]) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectPlug(p...)
	}
}

// WithLogger registers loggers for the generator.
func WithLogger[T any](l ...types.Logger) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectLogger(l...)
	}
}

// WithCircuitBreaker attaches a circuit breaker to the generator.
func WithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectCircuitBreaker(cb)
	}
}

// WithSensor registers sensors for the generator.
func WithSensor[T any](s ...types.Sensor[T]) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectSensor(s...)
	}
}
