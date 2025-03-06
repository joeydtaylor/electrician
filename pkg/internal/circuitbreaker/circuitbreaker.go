// Package circuitbreaker provides a generic circuit breaker implementation for handling failures
// in a distributed system or any operation that can fail. It prevents a system from performing
// operations that are likely to fail by opening the circuit when a threshold of failures has been reached.
package circuitbreaker

import (
	"context"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// CircuitBreaker represents a mechanism to stop the flow of operations when a predefined threshold of errors is reached.
// It prevents potential system overloads or cascading failures by temporarily halting operations.
// This allows the system to stabilize before resuming normal operations.
type CircuitBreaker[T any] struct {
	lastErrorTime     int64
	debouncePeriod    int64                   // in nanoseconds
	componentMetadata types.ComponentMetadata // Metadata provides identifying information for the circuit breaker, primarily used for logging and monitoring.
	ctx               context.Context         // Context controls the lifecycle of the circuit breaker, allowing it to be gracefully shutdown or cancelled.
	cancel            context.CancelFunc      // CancelFunc is a function that can be called to stop all operations and go routines internally managed by the circuit breaker.
	loggers           []types.Logger          // Loggers are used for outputting information about circuit breaker operations and state changes.
	loggersLock       sync.Mutex              // LoggersLock ensures that changes to the loggers slice are thread-safe.
	errorThreshold    int32                   // ErrorThreshold specifies the number of errors required to trip the circuit breaker, transitioning its state from closed to open.
	timeWindow        int64                   // TimeWindow is the duration (in nanoseconds) within which error count is evaluated against the threshold to determine if tripping should occur.
	errorCount        int32                   // ErrorCount tracks the number of errors that have occurred within the current time window.
	lastTripped       int64                   // LastTripped records the last timestamp (in nanoseconds) when the circuit breaker transitioned to the open state.
	allowed           int32                   // Allowed indicates whether the circuit breaker is in the closed state (1) and permitting operations, or in the open state (0) and blocking operations.
	neutralWires      []types.Wire[T]         // GroundWires are optional components that can be used to buffer operations when the circuit breaker is open, allowing them to be reprocessed later.
	resetNotifyChan   chan struct{}           // ResetNotifyChan is a channel that emits a signal when the circuit breaker resets, useful for triggering external actions.
	sensors           []types.Sensor[T]
	sensorLock        sync.Mutex
	allowLock         sync.Mutex
	tripLock          sync.Mutex
	resetLock         sync.Mutex
	configLock        sync.Mutex
	errorLock         sync.Mutex
}

// NewCircuitBreaker creates a new instance of a generic circuit breaker for operations of type T.
// This function initializes the circuit breaker with specified error threshold and time window, along with any additional options provided.
// The error threshold dictates the number of failures required before tripping the circuit, while the time window sets the period over which errors are counted.
//
// Parameters:
// ctx - The parent context for creating a new context specific to this circuit breaker. It controls the lifecycle of the circuit breaker.
// errorThreshold - The number of errors after which the circuit breaker will open, preventing further operations.
// timeWindow - The duration to consider when counting errors before potentially tripping the circuit.
// options - A variadic list of functions that apply configurations to the circuit breaker, such as setting loggers, ground wires, or other behavioral modifications.
//
// Returns:
// A fully initialized instance of a CircuitBreaker that monitors operations and manages state based on the predefined rules.
func NewCircuitBreaker[T any](ctx context.Context, errorThreshold int, timeWindow time.Duration, options ...types.Option[types.CircuitBreaker[T]]) types.CircuitBreaker[T] {
	ctx, cancel := context.WithCancel(ctx)
	cb := &CircuitBreaker[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID: utils.GenerateUniqueHash(), // Generates a unique identifier for each instance for tracking and logging purposes.
		},
		errorThreshold:  int32(errorThreshold),
		timeWindow:      int64(timeWindow),
		allowed:         1,
		sensors:         make([]types.Sensor[T], 0),
		neutralWires:    make([]types.Wire[T], 0),
		resetNotifyChan: make(chan struct{}), // Initializes the channel for notifying external components on reset events.

	}

	for _, option := range options {
		option(cb) // Applies additional configurations provided by the options.
	}

	cb.notifyCircuitBreakerCreation()

	return cb
}
