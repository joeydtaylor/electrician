// Package wire provides a modular framework designed to build flexible and efficient data processing pipelines
// within the Electrician toolkit. It enables dynamic control of data flow through various processing stages,
// ensuring high throughput and adaptability under different operational requirements.
//
// The Wire component serves as a core processing unit for data ingestion, transformation, routing, and egress.
// It integrates with external components such as circuit breakers (for resilience), generators (for data production),
// and transformers (for data manipulation). Its design emphasizes configurability and control via runtime options,
// allowing the Wire to be tailored to a wide range of data processing scenarios.
//
// Key features of the Wire component include:
// - Configurable input and output channels for dynamic data routing.
// - Concurrent processing capabilities with adjustable buffering and concurrency controls.
// - Robust error handling via a dedicated error channel and circuit breaker integration.
// - Extensive logging and monitoring using attached loggers and sensors.
// - Customizable data transformation and encoding using user-defined functions.
//
// This package provides the necessary tools to build resilient, high-performance data processing pipelines
// capable of handling both streaming and batch loads with complex transformation requirements.

package wire

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Wire represents a data processing pipeline.
// It encapsulates the core functionality for receiving, transforming, and routing data elements.
// Wire is designed to be highly configurable and adaptable to a variety of processing scenarios.
type Wire[T any] struct {
	componentMetadata    types.ComponentMetadata                                 // Metadata for the wire, including ID and type.
	inChan               chan T                                                  // Input channel for receiving elements.
	OutputChan           chan T                                                  // Output channel for emitting processed elements.
	bufferMutex          sync.Mutex                                              // Protects access to the output buffer and encoder.
	OutputBuffer         *bytes.Buffer                                           // Buffer for storing processed data.
	encoder              types.Encoder[T]                                        // Encoder for serializing processed elements.
	loggers              []types.Logger                                          // Loggers for recording events and errors.
	loggersLock          sync.Mutex                                              // Protects access to the loggers slice.
	generators           []types.Generator[T]                                    // Generators for autonomously producing elements.
	transformations      []types.Transformer[T]                                  // Functions applied to elements for transformation.
	sensors              []types.Sensor[T]                                       // Sensors for monitoring the wire's operation.
	sensorLock           sync.Mutex                                              // Protects access to the sensors slice.
	ctx                  context.Context                                         // Context for managing the wire's lifecycle.
	cancel               context.CancelFunc                                      // Function to cancel the wire's context.
	wg                   sync.WaitGroup                                          // Manages concurrent operations.
	completeSignal       chan struct{}                                           // Signals the completion of all processing.
	errorChan            chan types.ElementError[T]                              // Channel for reporting processing errors.
	closeErrorChanOnce   sync.Once                                               // Ensures the error channel is closed only once.
	closeOutputChanOnce  sync.Once                                               // Ensures the output channel is closed only once.
	closeInputChanOnce   sync.Once                                               // Ensures the input channel is closed only once.
	cbLock               sync.Mutex                                              // Protects access to the circuit breaker.
	terminateOnce        sync.Once                                               // Ensures termination logic is executed only once.
	CircuitBreaker       types.CircuitBreaker[T]                                 // Circuit breaker for fault tolerance.
	controlChan          chan bool                                               // Channel for controlling circuit breaker state.
	maxBufferSize        int                                                     // Maximum capacity for input and output channel buffers.
	maxConcurrency       int                                                     // Maximum number of concurrent processing operations.
	concurrencySem       chan struct{}                                           // Semaphore to limit concurrent operations.
	started              int32                                                   // Atomic flag indicating whether the wire has been started.
	insulatorFunc        func(ctx context.Context, elem T, err error) (T, error) // Function for error recovery (insulation).
	retryThreshold       int                                                     // Maximum number of insulator retry attempts.
	retryInterval        time.Duration                                           // Time interval between insulator retries.
	surgeProtector       types.SurgeProtector[T]                                 // Manages rate limiting during data surges.
	queueRetryMaxAttempt int                                                     // Maximum retry attempts for queue operations.
	isClosed             bool                                                    // Indicates if the wire's channels have been closed.
	closeLock            sync.Mutex                                              // Protects access to the isClosed flag.
}

// NewWire creates a new Wire instance configured with the provided options.
// It initializes default values and then applies any configuration options passed in.
// Parameters:
//   - ctx: The context that governs the wire's lifecycle.
//   - options: A variadic list of option functions to configure the wire.
//
// Returns:
//   - types.Wire[T]: A newly created and configured Wire instance.
func NewWire[T any](ctx context.Context, options ...types.Option[types.Wire[T]]) types.Wire[T] {

	// Create a cancellable context for the wire.
	ctx, cancel := context.WithCancel(ctx)
	w := &Wire[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(), // Generate a unique identifier for the wire.
			Type: "WIRE",
		},
		transformations: []types.Transformer[T]{}, // Initialize an empty list of transformations.
		ctx:             ctx,                      // Set the wire's context.
		cancel:          cancel,                   // Assign the cancel function.
		completeSignal:  make(chan struct{}),      // Channel to signal completion.
		sensors:         []types.Sensor[T]{},      // Initialize an empty list of sensors.
		OutputBuffer:    &bytes.Buffer{},          // Initialize the output buffer.
		controlChan:     make(chan bool, 1),       // Buffered channel for circuit breaker control.
		maxBufferSize:   1000,                     // Default maximum buffer size.
		maxConcurrency:  10000,                    // Default maximum concurrency.
	}

	// Apply any provided configuration options.
	for _, opt := range options {
		opt(w)
	}

	// Initialize the concurrency semaphore and channels.
	w.concurrencySem = make(chan struct{}, w.maxConcurrency)
	w.SetInputChannel(make(chan T, w.maxBufferSize))
	w.SetOutputChannel(make(chan T, w.maxBufferSize))
	w.SetErrorChannel(make(chan types.ElementError[T], w.maxBufferSize))
	w.OutputBuffer = &bytes.Buffer{}

	// Start the circuit breaker ticker if a circuit breaker is configured.
	if w.CircuitBreaker != nil {
		go w.startCircuitBreakerTicker()
	}

	return w // Return the newly created Wire instance.
}
