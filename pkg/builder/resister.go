package builder

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/resister"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func NewResister[T any](ctx context.Context, options ...types.Option[types.Resister[T]]) types.Resister[T] {
	return resister.NewResister[T](ctx, options...)
}

func ResisterWithLogger[T any](logger ...types.Logger) types.Option[types.Resister[T]] {
	return resister.WithLogger[T](logger...)
}

func ResisterWithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.Resister[T]] {
	return resister.WithSensor[T](sensor...)
}
