package types

import "context"

// Receiver defines the behaviors required for components capable of receiving data
// within the system. Receivers can be standalone or part of larger constructs like
// pipelines or processing nodes. This interface facilitates generic handling of data
// types through channel communication and supports integration with other components
// such as loggers and circuit breakers.
type Receiver[T any] interface {

	// ConnectCircuitBreaker attaches a CircuitBreaker to the Receiver. This is used to manage
	// the flow of incoming data based on the health of the system or the Receiver itself,
	// preventing overload and ensuring resilience.
	ConnectCircuitBreaker(cb CircuitBreaker[T])

	// ConnectLogger attaches one or more Logger instances to the Receiver. These loggers
	// are used for outputting logs related to the data receiving process, facilitating
	// debugging and operational monitoring.
	ConnectLogger(...Logger)

	// GetComponentMetadata retrieves the metadata associated with the Receiver. This metadata
	// typically includes identifiers like name and ID that may be used for logging, monitoring,
	// and managing configurations dynamically.
	GetComponentMetadata() ComponentMetadata

	// GetOutputChannel returns a channel of type T through which data items are output.
	// This channel is used to pass data to subsequent processing stages or components.
	GetOutputChannel() chan T

	// IsStarted checks if the Receiver has begun processing data. This method can be used to
	// monitor the state of the Receiver, ensuring that it is active when expected.
	IsStarted() bool

	// NotifyLoggers sends a formatted log message to all attached loggers at a specified
	// log level. This method is used for dynamic logging throughout the Receiver's operation,
	// enhancing traceability and diagnostics.
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	// SetComponentMetadata sets specific metadata fields such as name and ID for the Receiver.
	// Modifying these properties can be crucial for reconfiguring components during runtime or
	// for administrative purposes.
	SetComponentMetadata(name string, id string)

	// Start initiates the data receiving process. This method transitions the Receiver from
	// an inactive or prepared state to an active state, ready to accept and process incoming data.
	Start(context.Context) error
}
