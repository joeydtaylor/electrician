package plug

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithAdapterFunc registers adapter funcs for the plug.
func WithAdapterFunc[T any](pf ...types.AdapterFunc[T]) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.AddAdapterFunc(pf...)
	}
}

// WithAdapter registers adapters for the plug.
func WithAdapter[T any](pc ...types.Adapter[T]) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.ConnectAdapter(pc...)
	}
}

// WithLogger registers loggers for the plug.
func WithLogger[T any](l ...types.Logger) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.ConnectLogger(l...)
	}
}

// WithSensor registers sensors for the plug.
func WithSensor[T any](s ...types.Sensor[T]) types.Option[types.Plug[T]] {
	return func(p types.Plug[T]) {
		p.ConnectSensor(s...)
	}
}
