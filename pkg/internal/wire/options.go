package wire

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithCircuitBreaker attaches a circuit breaker to the wire.
//
// Prefer calling this during construction (NewWire options) before Start.
func WithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectCircuitBreaker(cb)
	}
}

// WithConcurrencyControl sets channel buffer size and worker concurrency.
//
// bufferSize controls the internal in/out/error channel capacities.
// maxRoutines controls the number of processing goroutines (workers).
func WithConcurrencyControl[T any](bufferSize int, maxRoutines int) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetConcurrencyControl(bufferSize, maxRoutines)
	}
}

// WithEncoder attaches an encoder that writes processed elements into the wire output buffer.
func WithEncoder[T any](e types.Encoder[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetEncoder(e)
	}
}

// WithGenerator registers one or more generators.
func WithGenerator[T any](generators ...types.Generator[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectGenerator(generators...)
	}
}

// WithInsulator configures retry-based recovery for transform failures.
func WithInsulator[T any](
	retryFunc func(ctx context.Context, elem T, err error) (T, error),
	threshold int,
	interval time.Duration,
) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetInsulator(retryFunc, threshold, interval)
	}
}

// WithLogger registers one or more loggers.
func WithLogger[T any](loggers ...types.Logger) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectLogger(loggers...)
	}
}

// WithSensor registers one or more sensors.
func WithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectSensor(sensors...)
	}
}

// WithSurgeProtector attaches a surge protector (rate limit / queue behavior).
func WithSurgeProtector[T any](surgeProtector types.SurgeProtector[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectSurgeProtector(surgeProtector)
	}
}

// WithTransformer appends one or more transformers in order.
func WithTransformer[T any](transformers ...types.Transformer[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectTransformer(transformers...)
	}
}

// WithComponentMetadata sets the wire's component metadata (name/id).
func WithComponentMetadata[T any](name string, id string) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetComponentMetadata(name, id)
	}
}

/*
	Aliases (non-breaking). Use these going forward.
*/

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

func WithTransformerFactory[T any](factory func() types.Transformer[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		ww, ok := w.(*Wire[T])
		if !ok {
			panic("wire.WithTransformerFactory: unsupported wire implementation")
		}
		ww.SetTransformerFactory(factory)
	}
}

func WithScratchBytes[T any](size int, fn func([]byte, T) (T, error)) types.Option[types.Wire[T]] {
	if size < 0 {
		size = 0
	}
	return WithTransformerFactory[T](func() types.Transformer[T] {
		buf := make([]byte, size) // allocated once per worker
		return func(v T) (T, error) {
			return fn(buf, v)
		}
	})
}
