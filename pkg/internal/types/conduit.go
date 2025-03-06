package types

import (
	"bytes"
	"context"
)

// Conduit represents the orchestration layer for data flow management through a series of processing stages,
// types.y known as wires. It ensures that data elements are processed in a structured and efficient manner,
// coordinating the flow between different components within the system.
type Conduit[T any] interface {
	// ConnectCircuitBreaker attaches a CircuitBreaker to the Conduit. This is used to manage the flow of
	// operations based on error conditions detected during the data processing stages.
	ConnectCircuitBreaker(CircuitBreaker[T])

	// ConnectConduit attaches another Conduit downstream. This allows for complex data processing pipelines
	// where output from one Conduit can be directly fed into another.
	ConnectConduit(Conduit[T])

	GetGenerators() []Generator[T]

	ConnectSurgeProtector(SurgeProtector[T])

	// ConnectGenerator attaches a data generator function to the Conduit. This generator is responsible for
	// producing data items that are processed through the Conduit's stages.
	ConnectGenerator(...Generator[T])

	// ConnectLogger attaches one or more loggers to the Conduit. These loggers are used to record operational
	// events and states within the Conduit, facilitating monitoring and debugging.
	ConnectLogger(...Logger)

	// ConnectSensor attaches one or more sensors to the Conduit. Sensors are used to monitor and measure various
	// aspects of the data flow, such as throughput and performance metrics.
	ConnectSensor(...Sensor[T])

	// ConnectWire connects one or more Wire components to the Conduit. Wires are the processing units within
	// the Conduit that handle data transformation, filtering, or any other required processing.
	ConnectWire(...Wire[T])

	// GetComponentMetadata retrieves the metadata associated with the Conduit, such as its unique identifier,
	// name, and type. This information can be crucial for system management and monitoring.
	GetComponentMetadata() ComponentMetadata

	// GetCircuitBreaker retrieves the CircuitBreaker associated with the Conduit, if any. This allows for
	// checking and managing the state of flow control based on processing errors.
	GetCircuitBreaker() CircuitBreaker[T]

	// GetInputChannel returns the input channel from which the Conduit receives data items to be processed.
	GetInputChannel() chan T

	// GetOutputChannel returns the output channel through which the Conduit sends processed data items.
	GetOutputChannel() chan T

	// IsStarted checks whether the Conduit is currently active and processing data.
	IsStarted() bool

	// Load retrieves the entire contents of the output buffer as a raw byte slice. This is often used for
	// debugging or when batch processing of data is required after a series of transformations.
	Load() *bytes.Buffer

	// LoadAsJSONArray retrieves the contents of the output buffer and formats them as a JSON array. This
	// method is particularly useful when the data needs to be exported or analyzed in a structured format.
	LoadAsJSONArray() ([]byte, error)

	// NotifyLoggers sends a formatted log message to all attached loggers at the specified log level. This
	// method supports dynamic logging and is essential for operational transparency.
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	// SetComponentMetadata sets specific metadata fields for the Conduit, such as its name and identifier.
	// This is particularly useful when you need to reconfigure or rename the Conduit dynamically during runtime.
	SetComponentMetadata(name string, id string)

	// SetConcurrencyControl sets parasensors related to the concurrency of operations within the Conduit,
	// such as buffer sizes and maximum number of concurrent processing routines. This is crucial for tuning
	// performance and handling load efficiently.
	SetConcurrencyControl(bufferSize int, maxRoutines int)

	// Start activates the Conduit, making it ready to receive and process data according to its configuration.
	Start(context.Context) error
	Restart(ctx context.Context) error

	// Submit takes a data item of type T and processes it through the Conduit. If successful, the item
	// is passed through the connected stages; otherwise, an error is returned.
	Submit(ctx context.Context, elem T) error

	// Terminate stops the operations of the Conduit, ensuring a clean and orderly shutdown of all processing
	// activities and resource deallocation.
	Stop() error
}
