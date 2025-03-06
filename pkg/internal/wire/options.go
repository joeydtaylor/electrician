// Package wire offers a set of configurable options that can be applied to the Wire components in the Electrician framework.
// These options allow customization of the wire's behavior and its interaction with other components in the data processing environment.
// This file defines functional options that enhance the flexibility of the Wire by allowing dynamic adjustments to its operational parameters.
//
// The functional options pattern encapsulates settings in functions that modify the state of Wire objects. This approach
// keeps the API flexible and allows configuration adjustments both at instantiation and at runtime, supporting complex
// and scalable data processing pipelines.
//
// Key features provided through these options include:
// - Connecting circuit breakers to manage fault tolerance.
// - Setting concurrency parameters to optimize performance.
// - Defining custom encoders for data serialization.
// - Attaching generator functions for autonomous data production.
// - Incorporating loggers for detailed event and error logging.
// - Adding sensors for performance monitoring.
// - Integrating transformation functions to modify data as it flows through the wire.
//
// Each option is designed to be applied at creation time or prior to operational use, ensuring that configurations
// are tailored to specific processing needs.

package wire

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithCircuitBreaker sets the circuit breaker for the wire component.
// This option connects a circuit breaker to manage failures and prevent cascading issues.
// Returns a types.Option[types.Wire[T]] that, when applied, configures the circuit breaker.
func WithCircuitBreaker[T any](cb types.CircuitBreaker[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectCircuitBreaker(cb)
	}
}

// WithConcurrencyControl configures the wire's buffering and concurrent processing parameters.
// Parameters:
//   - bufferSize: The capacity for buffering incoming elements.
//   - maxRoutines: The maximum number of concurrent processing routines.
//
// Returns a types.Option[types.Wire[T]] that applies these settings.
func WithConcurrencyControl[T any](bufferSize int, maxRoutines int) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetConcurrencyControl(bufferSize, maxRoutines)
	}
}

// WithEncoder sets the encoder for the wire component.
// The encoder serializes processed elements before output.
// Returns a types.Option[types.Wire[T]] that applies the provided encoder.
func WithEncoder[T any](e types.Encoder[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetEncoder(e)
	}
}

// WithGenerator attaches one or more generator functions to the wire.
// Generators autonomously produce elements for processing.
// Returns a types.Option[types.Wire[T]] that connects the provided generators.
func WithGenerator[T any](generator ...types.Generator[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectGenerator(generator...)
	}
}

// WithInsulator sets the insulator function for error recovery.
// The insulator function provides retry logic for failed processing attempts.
// Parameters:
//   - retryFunc: The function to invoke for retrying processing.
//   - threshold: The maximum number of retry attempts.
//   - interval: The duration to wait between retry attempts.
//
// Returns a types.Option[types.Wire[T]] that configures the insulator.
func WithInsulator[T any](retryFunc func(ctx context.Context, elem T, err error) (T, error), threshold int, interval time.Duration) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetInsulator(retryFunc, threshold, interval)
	}
}

// WithLogger attaches one or more logger instances to the wire.
// Loggers record events and errors to aid in debugging and monitoring.
// Returns a types.Option[types.Wire[T]] that connects the provided loggers.
func WithLogger[T any](logger ...types.Logger) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectLogger(logger...)
	}
}

// WithSensor attaches one or more sensor instances to the wire.
// Sensors monitor the wire’s operation and collect metrics.
// Returns a types.Option[types.Wire[T]] that connects the provided sensors.
func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectSensor(sensor...)
	}
}

// WithSurgeProtector sets the surge protector for the wire component.
// The surge protector manages rate limiting and protects against sudden load spikes.
// Returns a types.Option[types.Wire[T]] that configures the surge protector.
func WithSurgeProtector[T any](surgeProtector types.SurgeProtector[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectSurgeProtector(surgeProtector)
	}
}

// WithTransformer attaches one or more transformation functions to the wire.
// Transformers modify or process elements as they flow through the wire.
// Returns a types.Option[types.Wire[T]] that applies the provided transformations.
func WithTransformer[T any](transformation ...types.Transformer[T]) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.ConnectTransformer(transformation...)
	}
}

// WithComponentMetadata sets custom metadata for the wire component.
// Metadata, such as name and ID, can be used for identification and logging.
// Returns a types.Option[types.Wire[T]] that configures the component metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.Wire[T]] {
	return func(w types.Wire[T]) {
		w.SetComponentMetadata(name, id)
	}
}
