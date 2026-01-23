package wire

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithCircuitBreaker attaches a circuit breaker.
func WithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectCircuitBreaker(cb)
	}
}

// WithConcurrencyControl configures buffer size and worker count.
func WithConcurrencyControl[T any](bufferSize int, maxRoutines int) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetConcurrencyControl(bufferSize, maxRoutines)
	}
}

// WithEncoder configures the output encoder.
func WithEncoder[T any](e types.Encoder[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetEncoder(e)
	}
}

// WithGenerator registers generators.
func WithGenerator[T any](generators ...types.Generator[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectGenerator(generators...)
	}
}

// WithInsulator configures retry behavior for transform errors.
func WithInsulator[T any](
	retryFunc func(ctx context.Context, elem T, err error) (T, error),
	threshold int,
	interval time.Duration,
) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetInsulator(retryFunc, threshold, interval)
	}
}

// WithLogger registers loggers.
func WithLogger[T any](loggers ...types.Logger) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectLogger(loggers...)
	}
}

// WithSensor registers sensors.
func WithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectSensor(sensors...)
	}
}

// WithSurgeProtector attaches a surge protector.
func WithSurgeProtector[T any](surgeProtector types.SurgeProtector[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectSurgeProtector(surgeProtector)
	}
}

// WithTransformer registers transformers.
func WithTransformer[T any](transformers ...types.Transformer[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectTransformer(transformers...)
	}
}

// WithComponentMetadata sets the wire metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetComponentMetadata(name, id)
	}
}

// WithBreaker is an alias for WithCircuitBreaker.
func WithBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Wire[T]] {
	return WithCircuitBreaker[T](cb)
}

// WithWorkers is an alias for WithConcurrencyControl.
func WithWorkers[T any](bufferSize int, workerCount int) types.Option[types.Wire[T]] {
	return WithConcurrencyControl[T](bufferSize, workerCount)
}

// WithTransformers is an alias for WithTransformer.
func WithTransformers[T any](transformers ...types.Transformer[T]) types.Option[types.Wire[T]] {
	return WithTransformer[T](transformers...)
}

// WithGenerators is an alias for WithGenerator.
func WithGenerators[T any](generators ...types.Generator[T]) types.Option[types.Wire[T]] {
	return WithGenerator[T](generators...)
}

// WithLoggers is an alias for WithLogger.
func WithLoggers[T any](loggers ...types.Logger) types.Option[types.Wire[T]] {
	return WithLogger[T](loggers...)
}

// WithSensors is an alias for WithSensor.
func WithSensors[T any](sensors ...types.Sensor[T]) types.Option[types.Wire[T]] {
	return WithSensor[T](sensors...)
}

// WithRetryPolicy is an alias for WithInsulator.
func WithRetryPolicy[T any](
	retryFunc func(ctx context.Context, elem T, err error) (T, error),
	maxAttempts int,
	interval time.Duration,
) types.Option[types.Wire[T]] {
	return WithInsulator[T](retryFunc, maxAttempts, interval)
}

// WithTransformerFactory registers a per-worker transformer factory.
func WithTransformerFactory[T any](factory func() types.Transformer[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		ww, ok := w.(*Wire[T])
		if !ok {
			panic("wire.WithTransformerFactory: unsupported wire implementation")
		}
		ww.SetTransformerFactory(factory)
	}
}

// WithScratchBytes allocates per-worker scratch space for a transformer.
func WithScratchBytes[T any](size int, fn func([]byte, T) (T, error)) types.Option[types.Wire[T]] {
	if size < 0 {
		size = 0
	}
	return WithTransformerFactory[T](func() types.Transformer[T] {
		buf := make([]byte, size)
		return func(v T) (T, error) {
			return fn(buf, v)
		}
	})
}
