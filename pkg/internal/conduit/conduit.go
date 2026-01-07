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
	"runtime"
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
	started            int32 // Indicates whether the conduit has been started.
	MaxBufferSize      int   // Maximum buffer size for channels.
	MaxConcurrency     int   // Maximum number of concurrent routines.
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
	childCtx, cancel := context.WithCancel(ctx)

	c := &Conduit[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "CONDUIT",
		},
		wires:          make([]types.Wire[T], 0),
		generators:     make([]types.Generator[T], 0),
		ctx:            childCtx,
		cancel:         cancel,
		completeSignal: sync.WaitGroup{},
		started:        0,
		MaxBufferSize:  1024,
		MaxConcurrency: runtime.GOMAXPROCS(0),
	}

	for _, option := range options {
		option(c)
	}

	// Ensure non-nil channels to avoid nil-channel goroutine stalls.
	if c.MaxBufferSize <= 0 {
		c.MaxBufferSize = 1000
	}
	if c.MaxConcurrency <= 0 {
		c.MaxConcurrency = 1
	}
	if c.InputChan == nil {
		c.InputChan = make(chan T, c.MaxBufferSize)
	}
	if c.OutputChan == nil {
		c.OutputChan = make(chan T, c.MaxBufferSize)
	}

	// Configure and chain wires.
	if len(c.wires) > 0 {
		for i, w := range c.wires {
			if w == nil {
				continue
			}

			// Attach loggers/sensors/cb/surge/generators as your current behavior intends.
			if c.loggers != nil {
				for _, l := range c.loggers {
					if l != nil {
						w.ConnectLogger(l)
					}
				}
			}
			if c.sensors != nil {
				for _, s := range c.sensors {
					if s != nil {
						w.ConnectSensor(s)
					}
				}
			}
			if c.CircuitBreaker != nil {
				w.ConnectCircuitBreaker(c.CircuitBreaker)
			}
			if len(c.generators) != 0 {
				w.ConnectGenerator(c.generators...)
			}
			if c.surgeProtector != nil {
				w.ConnectSurgeProtector(c.surgeProtector)
			}

			// Concurrency control: buffer size + max routines.
			w.SetConcurrencyControl(c.MaxBufferSize, c.MaxConcurrency)

			// Chain: previous wire output feeds next wire input.
			if i > 0 && c.wires[i-1] != nil {
				w.SetInputChannel(c.wires[i-1].GetOutputChannel())
			}
		}

		// Force the last wire to emit to the conduit OutputChan.
		last := c.wires[len(c.wires)-1]
		if last != nil {
			last.SetOutputChannel(c.OutputChan)
		}
	}

	return c
}
