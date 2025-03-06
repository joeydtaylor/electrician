package builder

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/plug"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func NewPlug[T any](ctx context.Context, options ...types.Option[types.Plug[T]]) types.Plug[T] {
	return plug.NewPlug[T](ctx, options...)
}

func PlugWithAdapterFunc[T any](pf ...types.AdapterFunc[T]) types.Option[types.Plug[T]] {
	return plug.WithAdapterFunc[T](pf...)
}

func PlugWithLogger[T any](l ...types.Logger) types.Option[types.Plug[T]] {
	return plug.WithLogger[T](l...)
}

func PlugWithAdapter[T any](pc ...types.Adapter[T]) types.Option[types.Plug[T]] {
	return plug.WithAdapter[T](pc...)
}

func PlugWithSensor[T any](s ...types.Sensor[T]) types.Option[types.Plug[T]] {
	return plug.WithSensor[T](s...)
}
