// Package conduit provides a conduit component that facilitates the connection and processing of data
// between various components in a system. It enables the creation of robust and flexible data processing pipelines,
// supporting complex data flows and interactions between different parts of a software system.
//
// Conduits act as central hubs within these pipelines, orchestrating the flow of data through multiple wires,
// each representing a specific processing step or transformation. By managing wires, circuit breakers, generators,
// loggers, and sensors, the conduit package empowers developers to build scalable, customizable, and observable data
// processing architectures.
//
// The key features of the conduit package include:
//  1. Flexible Data Processing: Conduits allow developers to connect and manage multiple wires, enabling the
//     implementation of complex data processing logic with ease. Wires can be customized to perform various tasks,
//     such as filtering, transformation, aggregation, and more.
//  2. Circuit Breaker Integration: The package supports the integration of circuit breakers, enhancing the reliability
//     of data processing pipelines by providing mechanisms to handle failures and control the flow of data under
//     exceptional conditions.
//  3. Observability and Monitoring: Conduits support the integration of sensors for monitoring data flow and performance.
//     Developers can track key metrics, such as throughput, latency, and error rates, to gain insights into the behavior
//     of their pipelines and optimize performance.
//  4. Logging and Debugging: The package includes built-in support for logging, allowing developers to capture and
//     analyze events and errors occurring within the pipeline. This facilitates debugging and troubleshooting,
//     improving the overall reliability and maintainability of the system.
//
// Overall, the conduit package serves as a fundamental building block for designing resilient and efficient data
// processing systems, providing developers with the tools and abstractions needed to tackle complex data challenges
// and build robust software architectures.
package conduit

import (
	"bytes"
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Conduit orchestrates the flow of data through a series of wires and manages their lifecycle.
type Conduit[T any] struct {
	componentMetadata  types.ComponentMetadata // Metadata for the conduit component.
	wires              []types.Wire[T]         // Slice of wires for processing data.
	loggers            []types.Logger          // Slice of loggers for logging events.
	loggersLock        sync.Mutex              // Mutex for protecting loggers slice.
	sensors            []types.Sensor[T]       // Slice of sensors for monitoring data flow.
	ctx                context.Context         // Context for managing cancellation.
	cancel             context.CancelFunc      // Function to cancel the context.
	Outputs            []*bytes.Buffer         // Slice of buffers for holding wire outputs.
	generators         []types.Generator[T]    // Slice of generator functions.
	NextConduit        types.Conduit[T]        // Pointer to the next conduit in the processing chain.
	OutputChan         chan T                  // Output channel for the final output from the conduit.
	InputChan          chan T                  // Input channel for the initial input into the conduit.
	completeSignal     sync.WaitGroup          // Wait group for synchronizing processing completion.
	terminateOnce      sync.Once               // Ensures Terminate is executed once.
	CircuitBreaker     types.CircuitBreaker[T] // Circuit Breaker attached to the Conduit.
	cbLock             sync.Mutex
	concurrencySem     chan struct{} // Semaphore for controlling concurrency.
	started            int32         // Indicates whether the conduit has been started.
	MaxBufferSize      int           // Maximum buffer size for channels.
	MaxConcurrency     int           // Maximum number of concurrent routines.
	surgeProtector     types.SurgeProtector[T]
	configLock         sync.Mutex
	surgeProtectorLock sync.Mutex
}

// Conduit orchestrates the flow of data through a series of wires and manages their lifecycle.
// It provides a central component for building data processing pipelines, allowing developers to connect,
// configure, and control the flow of data between different processing stages.
//
// The Conduit struct includes various fields and methods for managing the lifecycle and behavior of the conduit:
// - componentMetadata: Metadata for the conduit component, including identifiers and type information.
// - wires: Slice of wires for processing data within the conduit.
// - loggers: Slice of loggers for logging events and messages related to conduit operations.
// - sensors: Slice of sensors for monitoring data flow and performance metrics.
// - ctx: Context for managing cancellation and synchronization of conduit operations.
// - cancel: Function to cancel the context associated with the conduit.
// - Outputs: Slice of buffers for holding output data from the conduit's wires.
// - generators: Slice of generator functions for generating input data to the conduit.
// - NextConduit: Pointer to the next conduit in the processing chain.
// - OutputChan: Output channel for passing the final output from the conduit.
// - InputChan: Input channel for receiving initial input data into the conduit.
// - completeSignal: Wait group for synchronizing the completion of conduit operations.
// - terminateOnce: Sync.Once for ensuring termination logic is executed only once.
// - CircuitBreaker: Circuit Breaker attached to the conduit for controlling data flow.
// - concurrencySem: Semaphore for controlling concurrency within the conduit.
// - started: Indicator for tracking whether the conduit has been started.
// - MaxBufferSize: Maximum buffer size for channels within the conduit.
// - MaxConcurrency: Maximum number of concurrent routines for processing data.
//
// NewConduit creates a new instance of a Conduit with the provided context and optional configuration options.
// It initializes the conduit with the specified wires, loggers, sensors, circuit breakers, generators, and other
// configuration settings. The function returns a types.Conduit[T] interface, allowing users to interact with
// the conduit through a generic interface.
//
// Parameters:
//   - ctx: Context for managing cancellation and synchronization of conduit operations.
//   - options: Optional configuration options for customizing the behavior of the conduit.
//
// Returns:
//   - types.Conduit[T]: An instance of the Conduit component configured according to the provided options.
func NewConduit[T any](ctx context.Context, options ...types.Option[types.Conduit[T]]) types.Conduit[T] {
	// Create a child context for the conduit and a cancellation function.
	childCtx, cancel := context.WithCancel(ctx)

	// Create the conduit instance.
	conduit := &Conduit[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(), // Generate a unique identifier for the conduit.
			Type: "CONDUIT",
		},
		wires:          make([]types.Wire[T], 0), // Initialize wires slice.
		generators:     make([]types.Generator[T], 0),
		ctx:            childCtx,         // Assign the child context to the conduit.
		cancel:         cancel,           // Assign the cancellation function to the conduit.
		completeSignal: sync.WaitGroup{}, // Initialize the complete signal for synchronization.
		started:        0,                // Initialize started indicator.
		MaxBufferSize:  1000,             // Set default maximum buffer size.
		MaxConcurrency: 100,              // Set default maximum concurrency.
	}

	// Apply provided options to the conduit.
	for _, option := range options {
		option(conduit)
	}

	var outputChan chan T // Declare output channel variable.
	if len(conduit.wires) > 0 {
		// Calculate buffer size and concurrency per wire.
		wireCount := len(conduit.wires)
		conduit.concurrencySem = make(chan struct{}, conduit.MaxConcurrency)
		// Initialize the input channel.

		for i, wire := range conduit.wires {
			// Initialize the concurrency semaphore.

			// Attach loggers, sensors, and circuit breakers.
			if conduit.loggers != nil {
				for _, l := range conduit.loggers {
					wire.ConnectLogger(l)
				}
			}
			if conduit.sensors != nil {
				for _, m := range conduit.sensors {
					wire.ConnectSensor(m)
				}
			}
			if conduit.CircuitBreaker != nil {
				wire.ConnectCircuitBreaker(conduit.CircuitBreaker)
			}
			if len(conduit.generators) != 0 {
				wire.ConnectGenerator(conduit.generators...)
			}
			if conduit.surgeProtector != nil {
				wire.ConnectSurgeProtector(conduit.surgeProtector)
			}

			// Set concurrency control for each wire.
			wire.SetConcurrencyControl(conduit.MaxBufferSize, conduit.MaxConcurrency)
			wire.SetSemaphore(&conduit.concurrencySem)

			if i == 0 {
				// Create input channel for the first wire.
				conduit.InputChan = make(chan T, conduit.MaxBufferSize)
			}

			if i > 0 {
				// Connect output of previous wire to input of current wire.
				prevOutputChan := conduit.wires[i-1].GetOutputChannel()
				conduit.wires[i].SetInputChannel(prevOutputChan)
			}

			if i == wireCount-1 {
				// Create output channel for the last wire.
				outputChan = make(chan T, conduit.MaxBufferSize)
				wire.SetOutputChannel(outputChan)
				conduit.OutputChan = outputChan
			}

			conduit.SetInputChannel(make(chan T, conduit.MaxBufferSize))
		}
	}

	return conduit // Return the created conduit instance.
}
