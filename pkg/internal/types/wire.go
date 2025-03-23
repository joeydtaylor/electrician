package types

import (
	"bytes"
	"context"
	"time"
)

type Wire[T any] interface {
	// ConnectCircuitBreaker attaches a CircuitBreaker to the Wire, providing a mechanism to halt
	// processing when error conditions meet certain thresholds, thereby protecting the system from failures.
	ConnectCircuitBreaker(CircuitBreaker[T])

	// ConnectGenerator attaches a data generator to the Wire, which is responsible for producing data
	// items that the Wire will process.
	ConnectGenerator(...Generator[T])

	// ConnectLogger attaches one or more loggers to the Wire for logging events and operations within the Wire.
	ConnectLogger(...Logger)

	// ConnectSensor attaches one or more sensors to the Wire, which are used to measure and monitor various
	// aspects of data processing like throughput and error rates.
	ConnectSensor(...Sensor[T])

	ConnectSurgeProtector(SurgeProtector[T])

	// ConnectTransformer attaches a transformation function to the Wire, which is used to modify or
	// process each data item flowing through the Wire.
	ConnectTransformer(...Transformer[T])

	// GetComponentMetadata retrieves the metadata associated with the Wire, providing information like
	// its identifier, name, and operational type.
	GetComponentMetadata() ComponentMetadata

	GetGenerators() []Generator[T]

	// GetCircuitBreaker retrieves the CircuitBreaker attached to the Wire, allowing for status checks
	// and control over the circuit-breaking functionality.
	GetCircuitBreaker() CircuitBreaker[T]

	// GetInputChannel returns the input channel from which the Wire receives data items.
	GetInputChannel() chan T

	// GetOutputBuffer retrieves a buffer where the Wire's output data is temporarily stored, typically
	// used for debugging or batch processing scenarios.
	GetOutputBuffer() *bytes.Buffer

	// GetOutputChannel returns the output channel through which the Wire sends processed data items.
	GetOutputChannel() chan T

	// IsStarted indicates whether the Wire has been started and is currently operational.
	IsStarted() bool

	// Load retrieves the contents of the output buffer as a raw byte slice, suitable for when data
	// needs to be accessed in a non-structured form.
	Load() *bytes.Buffer

	SetSemaphore(sem *chan struct{})

	// LoadAsJSONArray retrieves the contents of the output buffer and returns it as a JSON array,
	// useful for exporting the data in a structured format.
	LoadAsJSONArray() ([]byte, error)

	// NotifyLoggers sends a formatted log message to all attached loggers at a specified log level.
	// This method is critical for dynamic logging throughout the Wire's operations.
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	// SetComponentMetadata sets the metadata for the Wire, such as its name and unique identifier.
	// This can be essential for reconfiguring or identifying the Wire during runtime.
	SetComponentMetadata(name string, id string)

	// SetConcurrencyControl configures settings related to the concurrency of operations within
	// the Wire, such as buffer sizes and the maximum number of concurrent routines.
	SetConcurrencyControl(bufferSize int, maxRoutines int)

	// SetEncoder configures an encoder that the Wire uses to encode data items before passing them
	// to the output channel or processing them further.
	SetEncoder(e Encoder[T])

	// SetInputChannel sets the input channel for receiving data items into the Wire.
	SetInputChannel(chan T)

	SetInsulator(retryFunc func(ctx context.Context, elem T, err error) (T, error), threshold int, interval time.Duration)

	// SetOutputChannel sets the output channel for sending processed data items from the Wire.
	SetOutputChannel(chan T)

	// Start initiates the operations of the Wire, making it ready to receive and process data.
	Start(context.Context) error
	Restart(ctx context.Context) error

	// Submit accepts a data item and processes it according to the Wire's configuration. This method
	// may involve transformations, checks by a circuit breaker, and other processing steps.
	Submit(ctx context.Context, elem T) error

	// Terminate stops the Wire's operations, ensuring a clean shutdown and proper resource deallocation.
	Stop() error
}
