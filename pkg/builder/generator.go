package builder

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/generator"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func NewGenerator[T any](ctx context.Context, options ...types.Option[types.Generator[T]]) types.Generator[T] {
	return generator.NewGenerator[T](ctx, options...)
}

func GeneratorWithPlug[T any](p ...types.Plug[T]) types.Option[types.Generator[T]] {
	return generator.WithPlug[T](p...)
}

func GeneratorWithLogger[T any](l ...types.Logger) types.Option[types.Generator[T]] {
	return generator.WithLogger[T](l...)
}

func GeneratorWithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Generator[T]] {
	return generator.WithCircuitBreaker[T](cb)
}

func GeneratorWithSensor[T any](s ...types.Sensor[T]) types.Option[types.Generator[T]] {
	return generator.WithSensor[T](s...)
}
