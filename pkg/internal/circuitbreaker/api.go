// Package circuitbreaker provides a generic circuit breaker implementation for handling failures
// in a distributed system or any operation that can fail. It prevents a system from performing
// operations that are likely to fail by opening the circuit when a threshold of failures has been reached.
// The circuit breaker can be configured with various parasensors to control its behavior, such as the
// error threshold, time window for resetting, and associated ground wires.
//
// The main functionality of the package is encapsulated in the CircuitBreaker type, which tracks the
// error count, manages the state of the circuit, and provides methods for controlling the flow of operations.
// The CircuitBreaker type supports methods like Allow, RecordError, Trip, and Reset, allowing users to
// interact with and manage the circuit breaker's behavior programmatically.
//
// The circuit breaker can be connected to ground wires using the ConnectGroundWire method, enabling it
// to control the flow of items through the wires based on its state. Additionally, loggers can be attached
// to the circuit breaker using the ConnectLogger method, facilitating event logging and observability.
//
// Users can retrieve metadata about the circuit breaker using the GetComponentMetadata method, which
// provides information such as the unique identifier of the circuit breaker. This metadata can be useful
// for identification and logging purposes within a distributed system.
//
// The package also offers a mechanism for notifying external components when the circuit breaker resets.
// The NotifyOnReset method returns a channel that is notified when the circuit breaker resets, enabling
// asynchronous reactions to reset events and enhancing resilience by allowing dependent components to adapt
// to changes in circuit breaker status.
//
// Overall, the circuitbreaker package provides a robust and configurable solution for managing failures
// and controlling the flow of operations within distributed systems, enhancing reliability and resilience.
package circuitbreaker

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Allow checks if the operation is allowed by the circuit breaker.
// It automatically resets the circuit breaker based on the elapsed time since the last trip.
// This method uses atomic operations for thread safety.
func (cb *CircuitBreaker[T]) Allow() bool {
	cb.allowLock.Lock() // Ensure exclusive access to check and possibly modify state
	defer cb.allowLock.Unlock()

	now := time.Now().UnixNano()
	elapsed := now - atomic.LoadInt64(&cb.lastTripped)

	if atomic.LoadInt32(&cb.allowed) == 0 && elapsed > cb.timeWindow {
		cb.Reset() // Safe to call Reset here as we hold the allowLock
	}

	return atomic.LoadInt32(&cb.allowed) == 1
}

// SetGroundWire associates a groundWire with the circuit breaker.
// This allows the circuit breaker to control the flow of items through the groundWire, logging the association.
func (cb *CircuitBreaker[T]) ConnectNeutralWire(groundWire ...types.Wire[T]) {
	cb.neutralWires = append(cb.neutralWires, groundWire...)
	for _, gw := range groundWire {
		cb.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectGroundWire, target: %v => Connected groundWire", cb.componentMetadata, gw.GetComponentMetadata())
	}
}

func (cb *CircuitBreaker[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	cb.sensors = append(cb.sensors, sensor...)
	for _, m := range sensor {
		cb.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectSensor, target: %v => Connected sensor", cb.componentMetadata, m.GetComponentMetadata())
	}
}

// ConnectLogger attaches a logger to the circuit breaker.
// It allows logging of circuit breaker events, enriching observability.
func (cb *CircuitBreaker[T]) ConnectLogger(logger types.Logger) {
	if cb.loggers == nil {
		cb.loggers = make([]types.Logger, 0)
	}
	cb.loggers = append(cb.loggers, logger)
	cb.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectLogger, target: %v => Connected logger", cb.componentMetadata, logger)
}

// GetComponentMetadata returns the metadata of the circuit breaker, including its unique identifier.
func (cb *CircuitBreaker[T]) GetComponentMetadata() types.ComponentMetadata {
	cb.configLock.Lock()
	defer cb.configLock.Unlock()
	cb.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: GetComponentMetadata, return: %v => GetComponentMetadata called", cb.componentMetadata, cb.errorCount)
	return cb.componentMetadata
}

// GetGroundWire returns the groundWire associated with the circuit breaker, if any.
// This groundWire is controlled based on the state of the circuit breaker.
func (cb *CircuitBreaker[T]) GetNeutralWires() []types.Wire[T] {
	cb.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: GetGroundWires, return: %v => GetGroundWires called", cb.componentMetadata, cb.neutralWires)
	return cb.neutralWires
}

// NotifyLoggers sends a formatted log message to all attached loggers.
// It supports different log levels and is safe to use with multiple loggers.
func (w *CircuitBreaker[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if w.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range w.loggers {
			if logger == nil {
				continue // Skip if the logger is nil
			}
			w.loggersLock.Lock()
			defer w.loggersLock.Unlock()
			switch level {
			case types.DebugLevel:
				logger.Debug(msg)
			case types.InfoLevel:
				logger.Info(msg)
			case types.WarnLevel:
				logger.Warn(msg)
			case types.ErrorLevel:
				logger.Error(msg)
			case types.DPanicLevel:
				logger.DPanic(msg)
			case types.PanicLevel:
				logger.Panic(msg)
			case types.FatalLevel:
				logger.Fatal(msg)
			}
		}
	}
}

// NotifyOnReset returns a channel that is notified when the circuit breaker resets.
// This provides an asynchronous mechanism for other parts of the system to react to reset events.
// It can be used to perform cleanup, reinitialization, or other recovery actions once the circuit breaker
// allows operations to proceed again. This method supports the reactive programming pattern within
// distributed systems, enhancing resilience by allowing dependent components to adapt to changes
// in circuit breaker status.
func (cb *CircuitBreaker[T]) NotifyOnReset() <-chan struct{} {
	cb.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: NotifyOnReset => Notified Reset Channel!", cb.componentMetadata)
	return cb.resetNotifyChan
}

// RecordError increments the error count and trips the circuit breaker if the error threshold is exceeded.
// It logs the event of exceeding the threshold using atomic operations for thread safety.
func (cb *CircuitBreaker[T]) RecordError() {
	cb.errorLock.Lock() // Ensure exclusive access to check and possibly modify state
	defer cb.errorLock.Unlock()
	now := time.Now().UnixNano()
	last := atomic.LoadInt64(&cb.lastErrorTime)
	cb.notifyError(last)
	if now-last > cb.debouncePeriod {
		atomic.StoreInt64(&cb.lastErrorTime, now)
		newCount := atomic.AddInt32(&cb.errorCount, 1)
		if newCount >= cb.errorThreshold && atomic.LoadInt32(&cb.allowed) == 1 {
			cb.Trip()
		}
		cb.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: RecordError, currentErrorCount: %d, errorThreshold: %d => Recorded Error!", cb.GetComponentMetadata(), newCount, cb.errorThreshold)
	}
}

// Example of how to set the debounce period (perhaps during initialization)
func (cb *CircuitBreaker[T]) SetDebouncePeriod(seconds int) {
	cb.debouncePeriod = int64(seconds) * 1e9 // convert seconds to nanoseconds
}

// Reset switches the circuit breaker to the closed state, allowing operations to be processed again.
// It resets the error count and, if a groundWire is associated, processes any buffered items.
// This method is thread-safe and logs the reset event.
func (cb *CircuitBreaker[T]) Reset() {
	cb.resetLock.Lock()
	defer cb.resetLock.Unlock()
	if atomic.CompareAndSwapInt32(&cb.allowed, 0, 1) {
		atomic.StoreInt32(&cb.errorCount, 0)
		cb.notifyReset(time.Now().UnixNano())
		cb.NotifyLoggers(types.InfoLevel, "component: level: INFO, result: SUCCESS, event: Reset, %s => Circuit breaker reset...", cb.componentMetadata)

	}
}

// SetComponentMetadata sets overrides for fields that are settable such as name and id.
// through which error messages are communicated.
// Parameters:
//   - name: The name to be set.
//   - id: The id to be set.
func (cb *CircuitBreaker[T]) SetComponentMetadata(name string, id string) {
	cb.componentMetadata.Name = name
	cb.componentMetadata.ID = id
}

// Trip switches the circuit breaker to the open state, stopping it from allowing further operations.
// It uses an atomic operation to ensure thread safety. This method logs the tripping event.
func (cb *CircuitBreaker[T]) Trip() {
	cb.tripLock.Lock()
	defer cb.tripLock.Unlock()
	if atomic.CompareAndSwapInt32(&cb.allowed, 1, 0) {

		now := time.Now().UnixNano()
		atomic.StoreInt64(&cb.lastTripped, now)
		nextResetTime := now + cb.timeWindow // Calculate the next reset time

		// Notify the trip with the time when the trip happened and the expected next reset time
		cb.notifyTrip(now, nextResetTime)

		// Log the tripping event along with the next reset time
		cb.NotifyLoggers(types.WarnLevel, "component: %s, level: WARN, result: SUCCESS, event: Trip, errorThreshold: %d, resetsIn: %d => Circuit breaker tripped! Next reset at: %v", cb.componentMetadata, cb.errorThreshold, cb.timeWindow, time.Unix(0, nextResetTime))
	}

}
