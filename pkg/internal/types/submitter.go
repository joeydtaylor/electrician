package types

import "context"

// Submitter defines the operations required for components that submit data within the system.
// This interface is designed to handle outgoing data operations, ensuring that data elements
// can be submitted to other parts of the system or external services efficiently and reliably.
type Submitter[T any] interface {
	// Submit sends an element of type T through the component. This method handles the logistics
	// of data submission, including error handling and possibly retries, depending on the implementation.
	// Context is used to manage timeouts and cancellations.
	Submit(ctx context.Context, elem T) error

	// ConnectGenerator attaches a data generator to the Wire, which is responsible for producing data
	// items that the Wire will process.
	ConnectGenerator(...Generator[T])

	GetGenerators() []Generator[T]

	// ConnectLogger attaches one or more Logger instances to the Submitter. These loggers
	// are used to output logs that are useful for monitoring the submission process and diagnosing issues.
	ConnectLogger(...Logger)
	Restart(ctx context.Context) error

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

	// Terminate halts the operations of the Submitter, ensuring a graceful shutdown and cleanup of resources.
	// This method is critical for preventing data loss and ensuring that the system remains consistent.
	Stop() error

	// IsStarted checks if the Submitter has started its operations. This method is useful for system checks
	// and ensuring that components are operational as expected.
	IsStarted() bool

	// Start activates the Submitter, transitioning it from an inactive state to an active state where
	// it can begin processing and submitting data.
	Start(context.Context) error
}
