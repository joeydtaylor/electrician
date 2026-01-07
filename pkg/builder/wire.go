package builder

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/codec"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/wire"
)

// NewWire creates a new wire with the provided context, output writer, encoder, and configuration options.
func NewWire[T any](ctx context.Context, options ...types.Option[types.Wire[T]]) types.Wire[T] {
	return wire.NewWire[T](ctx, options...)
}

// WireWithCircuitBreaker configures a CircuitBreaker for the Wire.
func WireWithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Wire[T]] {
	return wire.WithCircuitBreaker[T](cb)
}

// WireWithComponentMetadata adds component metadata overrides.
func WireWithComponentMetadata[T any](name string, id string) types.Option[types.Wire[T]] {
	return wire.WithComponentMetadata[T](name, id)
}

// WireWithConcurrencyControl sets the concurrency control for the Wire.
func WireWithConcurrencyControl[T any](bufferSize int, maxRoutines int) types.Option[types.Wire[T]] {
	return wire.WithConcurrencyControl[T](bufferSize, maxRoutines)
}

// WireWithEncoder configures an encoder for the Wire.
func WireWithEncoder[T any](e codec.Encoder[T]) types.Option[types.Wire[T]] {
	return wire.WithEncoder[T](e)
}

// WireWithPlug sets the generator function of the wire.
func WireWithGenerator[T any](generator types.Generator[T]) types.Option[types.Wire[T]] {
	return wire.WithGenerator[T](generator)
}

// WireWithInsulator adds an insulator to the wire.
func WireWithInsulator[T any](retryFunc func(ctx context.Context, elem T, err error) (T, error), threshold int, interval time.Duration) types.Option[types.Wire[T]] {
	return wire.WithInsulator[T](retryFunc, threshold, interval)
}

// WireWithLogger adds one or more loggers to the wire.
func WireWithLogger[T any](logger ...types.Logger) types.Option[types.Wire[T]] {
	return wire.WithLogger[T](logger...)
}

// WireWithSensor adds a sensor to the wire to monitor its performance and events.
func WireWithSensor[T any](sensor types.Sensor[T]) types.Option[types.Wire[T]] {
	return wire.WithSensor[T](sensor)
}

// WithSurgeProtector adds a surge protector to the wire component.
// This option allows adding surge protectors to the wire.
func WireWithSurgeProtector[T any](surgeProtector types.SurgeProtector[T]) types.Option[types.Wire[T]] {
	return wire.WithSurgeProtector[T](surgeProtector)
}

// WireWithTransformer adds a transformation function to the wire.
func WireWithTransformer[T any](transformation ...types.Transformer[T]) types.Option[types.Wire[T]] {
	return wire.WithTransformer[T](transformation...)
}

func WithRetryPolicy[T any](
	retryFunc func(ctx context.Context, elem T, err error) (T, error),
	maxAttempts int,
	interval time.Duration,
) types.Option[types.Wire[T]] {
	return wire.WithInsulator[T](retryFunc, maxAttempts, interval)
}

func WireWithTransformerFactory[T any](factory func() func(T) (T, error)) types.Option[types.Wire[T]] {
	return wire.WithTransformerFactory[T](func() types.Transformer[T] {
		return factory()
	})
}

func WireWithScratchBytes[T any](size int, fn func([]byte, T) (T, error)) types.Option[types.Wire[T]] {
	return wire.WithScratchBytes[T](size, fn)
}
