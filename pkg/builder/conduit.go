package builder

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/conduit"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ConduitWithCircuitBreaker sets the circuit breaker for the Conduit.
func ConduitWithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Conduit[T]] {
	return conduit.WithCircuitBreaker[T](cb)
}

// ConduitWithComponentMetadata adds component metadata overrides.
func ConduitWithComponentMetadata[T any](name string, id string) types.Option[types.Conduit[T]] {
	return conduit.WithComponentMetadata[T](name, id)
}

// ConduitWithConcurrencyControl sets the concurrency control for the Conduit.
func ConduitWithConcurrencyControl[T any](bufferSize int, maxRoutines int) types.Option[types.Conduit[T]] {
	return conduit.WithConcurrencyControl[T](bufferSize, maxRoutines)
}

// ConduitWithPlug sets the generator function for the Conduit.
func ConduitWithGenerator[T any](generator types.Generator[T]) types.Option[types.Conduit[T]] {
	return conduit.WithGenerator[T](generator)
}

func ConduitWithSurgeProtector[T any](sp types.SurgeProtector[T]) types.Option[types.Conduit[T]] {
	return conduit.WithSurgeProtector[T](sp)
}

// ConduitWithLogger adds loggers to the Conduit.
func ConduitWithLogger[T any](logger ...types.Logger) types.Option[types.Conduit[T]] {
	return conduit.WithLogger[T](logger...)
}

// ConduitWithSensor attaches a sensor to the Conduit to monitor performance.
func ConduitWithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.Conduit[T]] {
	return conduit.WithSensor[T](sensor...)
}

// ConduitWithWire connects wires to the Conduit.
func ConduitWithWire[T any](wires ...types.Wire[T]) types.Option[types.Conduit[T]] {
	return conduit.WithWire[T](wires...)
}

// NewConduit creates a new Conduit with specified options.
func NewConduit[T any](ctx context.Context, options ...types.Option[types.Conduit[T]]) types.Conduit[T] {
	return conduit.NewConduit(ctx, options...)
}
