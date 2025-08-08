package types

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"google.golang.org/grpc"
)

// ReceivingRelay defines the operations for a Receiving Relay, which is responsible for
// receiving data from external sources or other components and managing the flow
// of data within the system. It provides mechanisms to start, manage, and monitor
// the data reception processes.
type ReceivingRelay[T any] interface {

	// ConnectLogger attaches one or more loggers to the Receiving Relay for operational
	// logging. These loggers will record events and states as configured.
	ConnectLogger(...Logger)

	// ConnectOutput connects one or more output submitters to the relay. These submitters
	// are used to forward processed data to other components or systems.
	ConnectOutput(...Submitter[T])

	// GetAddress returns the current network address on which the relay is configured
	// to receive data. This address is crucial for network configurations and routing.
	GetAddress() string

	// GetComponentMetadata retrieves the metadata of the Receiving Relay, such as its ID,
	// name, and type, which are useful for identification and configuration purposes.
	GetComponentMetadata() ComponentMetadata

	SetDecryptionKey(string)

	// IsRunning checks if the Receiving Relay is actively running and processing data.
	// This method is useful for health checks and operational monitoring.
	IsRunning() bool

	// Listen starts listening on the configured network address. It can operate in a
	// continuous mode or retry listening for incoming connections based on the
	// provided parasensors.
	// listenForever determines whether to listen indefinitely.
	// retryInSeconds specifies the interval for retries in case of failures.
	Listen(listenForever bool, retryInSeconds int) error

	// NotifyLoggers sends a formatted log message to all attached loggers at a specified
	// log level. It is used to log significant events and operations within the relay.
	NotifyLoggers(level LogLevel, format string, args ...interface{})

	// Receive handles the reception of a single wrapped payload through a standard
	// context. It processes the payload and returns an acknowledgment or an error
	// if the reception fails.
	Receive(ctx context.Context, payload *relay.WrappedPayload) (*relay.StreamAcknowledgment, error)

	// SetComponentMetadata sets the metadata for the Receiving Relay. This is useful for
	// dynamically configuring or re-identifying the relay during operation.
	SetComponentMetadata(name string, id string)

	// SetAddress configures the network address for the relay. This address is used for
	// all incoming network communications.
	SetAddress(string)

	// SetDataChannel configures the channel used for internal data handling with a specific
	// buffer size. This is important for managing data flow and preventing bottlenecks.
	SetDataChannel(uint32)

	// SetTLSConfig sets the configuration for TLS, ensuring secure communications for
	// the incoming data. This setup is critical for maintaining data integrity and security.
	SetTLSConfig(*TLSConfig)

	// Start initiates the receiving operations of the relay. It prepares the relay
	// to accept incoming data and ensures that it is ready to process and route data
	// as configured.
	Start(context.Context) error

	// StreamReceive handles the reception of streams of data via a gRPC server stream.
	// This method is crucial for handling continuous data flows and streaming scenarios.
	StreamReceive(stream relay.RelayService_StreamReceiveServer) error

	// Terminate stops the receiving operations of the relay. This method ensures that
	// all resources are released, all ongoing data receptions are gracefully shut down,
	// and the system is cleaned up properly.
	Stop()

	// --- New auth-related hooks (optional) ---

	// --- Auth (optional) ---
	SetAuthenticationOptions(*relay.AuthenticationOptions)
	SetAuthInterceptor(grpc.UnaryServerInterceptor) // <-- fix here
	SetStaticHeaders(map[string]string)
	SetDynamicAuthValidator(func(ctx context.Context, md map[string]string) error)
}
