package plug

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func WithAdapterFunc[T any](pf ...types.AdapterFunc[T]) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.AddAdapterFunc(pf...)
	}
}

func WithAdapter[T any](pc ...types.Adapter[T]) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.ConnectAdapter(pc...)
	}
}

func WithLogger[T any](l ...types.Logger) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.ConnectLogger(l...)
	}
}

func WithSensor[T any](s ...types.Sensor[T]) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.ConnectSensor(s...)
	}
}
