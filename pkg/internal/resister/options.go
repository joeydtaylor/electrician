package resister

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithLogger registers loggers for resister events.
func WithLogger[T any](logger ...types.Logger) types.Option[types.Resister[T]] {
	return func(r types.Resister[T]) {
		r.ConnectLogger(logger...)
	}
}

// WithSensor registers sensors for resister events.
func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.Resister[T]] {
	return func(r types.Resister[T]) {
		r.ConnectSensor(sensor...)
	}
}
