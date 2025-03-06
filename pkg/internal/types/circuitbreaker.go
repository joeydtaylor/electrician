package types

// CircuitBreaker defines the interface for managing state transitions in response to failures
// within wires to prevent system overloads and maintain reliability. The interface provides mechanisms
// to stop and resume operations dynamically based on error conditions and system health checks.
type CircuitBreaker[T any] interface {
	ConnectSensor(sensor ...Sensor[T])
	// Allow checks the current state of the circuit breaker to determine if operations can proceed.
	// Returns true if the state is "closed", allowing normal operation, and false if the state is "open",
	// which means operations should be halted to prevent further errors.
	Allow() bool

	SetDebouncePeriod(seconds int)

	// ConnectNeutralWire attaches additional wires that can handle requests when the circuit breaker is tripped.
	// These ground wires are used to safely reroute or queue operations until the circuit breaker resets.
	ConnectNeutralWire(...Wire[T])

	// ConnectLogger attaches a Logger to the circuit breaker for logging state changes and significant events
	// related to circuit breaker operations. This enhances monitoring and debugging capabilities.
	ConnectLogger(Logger)

	// GetComponentMetadata retrieves the metadata associated with the circuit breaker, such as its identifier and type,
	// which can be useful for logging, monitoring, and dynamically managing circuit breaker configurations.
	GetComponentMetadata() ComponentMetadata

	// GetNeutralWires returns a slice of all wires connected as ground wires. These are used when the circuit breaker is open.
	GetNeutralWires() []Wire[T]

	// NotifyLoggers sends a formatted log message to all attached loggers at the specified log level.
	// This method is essential for real-time logging of circuit breaker events, facilitating system monitoring and analysis.
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	// NotifyOnReset provides a channel that emits a notification when the circuit breaker resets.
	// This can be used to trigger recovery or other actions once the circuit breaker allows operations to resume.
	NotifyOnReset() <-chan struct{}

	// RecordError is called to indicate that an error has occurred. This method is responsible for incrementing
	// the error count and potentially tripping the circuit breaker if a predefined threshold is exceeded.
	RecordError()

	// Reset transitions the circuit breaker to a "closed" state, allowing operations to continue.
	// This method is typically called automatically after a defined cooldown period or manually to recover operations.
	Reset()

	// SetComponentMetadata sets specific metadata fields for the circuit breaker, such as name and ID.
	// This is useful for reconfiguring or re-identifying the circuit breaker during runtime.
	SetComponentMetadata(name string, id string)

	// Trip transitions the circuit breaker to an "open" state, halting the flow of operations
	// to prevent further failures and system instability.
	Trip()
}
