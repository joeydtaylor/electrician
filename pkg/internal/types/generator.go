package types

import "context"

type Generator[T any] interface {
	// ConnectCircuitBreaker attaches a CircuitBreaker to the Wire, providing a mechanism to halt
	// processing when error conditions meet certain thresholds, thereby protecting the system from failures.
	ConnectCircuitBreaker(CircuitBreaker[T])
	Start(context.Context) error
	Stop() error
	IsStarted() bool
	Restart(ctx context.Context) error
	ConnectPlug(...Plug[T])
	NotifyLoggers(level LogLevel, format string, args ...interface{})
	ConnectLogger(...Logger)
	ConnectToComponent(submitters ...Submitter[T])
	// GetComponentMetadata retrieves the metadata associated with the Conduit, such as its unique identifier,
	// name, and type. This information can be crucial for system management and monitoring.
	GetComponentMetadata() ComponentMetadata
	// SetComponentMetadata sets specific metadata fields for the Conduit, such as its name and identifier.
	// This is particularly useful when you need to reconfigure or rename the Conduit dynamically during runtime.
	SetComponentMetadata(name string, id string)

	// ConnectSensor attaches one or more sensors to the Conduit. Sensors are used to monitor and measure various
	// aspects of the data flow, such as throughput and performance metrics.
	ConnectSensor(...Sensor[T])
}
