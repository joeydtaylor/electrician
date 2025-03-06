package types

// ResisterInterface defines the operations available for a Resister.
type Resister[T any] interface {
	Len() int                       // Returns the length of the queue.
	Push(element *Element[T]) error // Pushes an element into the queue, potentially increasing its priority.
	Pop() *Element[T]               // Pops the highest priority element from the queue.
	ConnectLogger(...Logger)

	ConnectSensor(...Sensor[T])

	// NotifyLoggers sends a formatted log message to all attached loggers at a specified log level.
	// This method is crucial for dynamic logging throughout the Submitter's operations, enhancing
	// operational visibility and traceability.
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	// GetComponentMetadata retrieves the metadata associated with the Submitter, including identifiers
	// such as name and ID, which are useful for logging, monitoring, and managing configurations dynamically.
	GetComponentMetadata() ComponentMetadata

	// SetComponentMetadata configures specific metadata fields such as name and ID for the Submitter.
	// Adjusting these properties can be essential for administrative tasks or runtime reconfiguration.
	SetComponentMetadata(name string, id string)
}
