package generator

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func WithPlug[T any](p ...types.Plug[T]) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectPlug(p...)
	}
}

func WithLogger[T any](l ...types.Logger) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectLogger(l...)
	}
}

func WithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectCircuitBreaker(cb)
	}
}

func WithSensor[T any](s ...types.Sensor[T]) types.Option[types.Generator[T]] {
	return func(g types.Generator[T]) {
		g.ConnectSensor(s...)
	}
}
