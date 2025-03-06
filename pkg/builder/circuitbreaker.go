package builder

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/circuitbreaker"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// CircuitBreakerWithComponentMetadata adds component metadata overrides.
func CircuitBreakerWithComponentMetadata[T any](name string, id string) types.Option[types.CircuitBreaker[T]] {
	return circuitbreaker.WithComponentMetadata[T](name, id)
}

func CircuitBreakerWithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.CircuitBreaker[T]] {
	return circuitbreaker.WithSensor[T](sensor...)
}

// CircuitBreakerWithNeutralWire attaches a ground wire to the CircuitBreaker.
func CircuitBreakerWithNeutralWire[T any](neutralWire types.Wire[T]) types.Option[types.CircuitBreaker[T]] {
	return circuitbreaker.WithNeutralWire[T](neutralWire)
}

// CircuitBreakerWithLogger adds a logger to the CircuitBreaker for logging its activities.
func CircuitBreakerWithLogger[T any](logger types.Logger) types.Option[types.CircuitBreaker[T]] {
	return circuitbreaker.WithLogger[T](logger)
}

// NewCircuitBreaker creates a new CircuitBreaker with the specified error threshold, time window, and additional options.
func NewCircuitBreaker[T any](ctx context.Context, errorThreshold int, timeWindow time.Duration, options ...types.Option[types.CircuitBreaker[T]]) types.CircuitBreaker[T] {
	return circuitbreaker.NewCircuitBreaker[T](ctx, errorThreshold, timeWindow, options...)
}
