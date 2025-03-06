package types

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

// ForwardRelay defines the operations for a Forward Relay, which is responsible for
// dynamically forwarding data from inputs to a ReceivingRelay. This interface
// is crucial for managing data flow and ensuring that data is appropriately routed through
// the system's components based on configurable criteria and operational states.
type ForwardRelay[T any] interface {
	// ConnectInput attaches one or more receivers to the Forward Relay. Receivers are components
	// that can accept data items of type T, which the Forward Relay will forward.
	ConnectInput(...Receiver[T])

	// ConnectLogger attaches one or more loggers to the Forward Relay. These loggers will
	// be used to output logs at various levels, providing insights into the operations and
	// states of the relay.
	ConnectLogger(...Logger)

	// GetAddress returns the network address that the Forward Relay uses for communication.
	// This address could be used for listening to incoming data or sending outgoing data,
	// depending on the relay's configuration and role.
	GetTargets() []string

	// GetComponentMetadata retrieves metadata about the Forward Relay, such as its ID, name,
	// and type. This metadata can be used for identification, logging, and monitoring purposes.
	GetComponentMetadata() ComponentMetadata

	// GetInput returns a slice of Receivers currently connected to the Forward Relay.
	// This method provides insight into which components are receiving data from the relay.
	GetInput() []Receiver[T]

	// IsRunning checks if the Forward Relay is currently active and processing data.
	// This method is useful for monitoring and management purposes to ensure the relay
	// is operational.
	IsRunning() bool

	// NotifyLoggers sends a formatted log message to all attached loggers at a specified
	// log level. This method supports dynamic logging throughout the relay's operation,
	// facilitating detailed and contextual logging.
	NotifyLoggers(level LogLevel, format string, args ...interface{})

	// SetAddress configures the network address for the Forward Relay. This address is
	// used for all network communications handled by the relay.
	SetTargets(...string)

	// SetComponentMetadata sets the metadata for the Forward Relay, such as its name and ID.
	SetComponentMetadata(name string, id string)

	// SetPerformanceOptions configures performance-related settings for the Forward Relay.
	// This includes options that could affect how data is processed and forwarded, such as
	// buffering capacities and concurrency controls.
	SetPerformanceOptions(*relay.PerformanceOptions)

	// SetTLSConfig sets the TLS configuration for secure communication. This method is crucial
	// for configuring security aspects of the relay, ensuring that data transmission is encrypted
	// and secure.
	SetTLSConfig(*TLSConfig)

	// Start initiates the operations of the Forward Relay. This method transitions the relay
	// from a stopped or uninitiated state into an active state where it can begin processing
	// and forwarding data.
	Start(context.Context) error

	// Submit takes a data item of type T and attempts to forward it through the relay. This method
	// is essential for injecting data into the relay for forwarding. It returns an error if the
	// submission fails, allowing for error handling and recovery.
	Submit(ctx context.Context, item T) error

	// Terminate stops the operations of the Forward Relay. This method is used to cleanly shutdown
	// the relay, ensuring that all resources are released and no data is left unprocessed.
	Stop()
}
