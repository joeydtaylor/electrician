package resister

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithLogger creates an option to add a logger to a Sensor.
//
// Parameters:
//   - logger: One or more logger instances to be added to the Sensor for logging.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	connects the specified logger(s) to the Sensor.
func WithLogger[T any](logger ...types.Logger) types.Option[types.Resister[T]] {
	return func(r types.Resister[T]) {
		r.ConnectLogger(logger...)
	}
}

func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.Resister[T]] {
	return func(r types.Resister[T]) {
		r.ConnectSensor(sensor...)
	}
}
