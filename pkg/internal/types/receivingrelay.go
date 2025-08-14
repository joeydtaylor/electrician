package types

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"google.golang.org/grpc"
)

// ReceivingRelay defines the operations for a Receiving Relay, which is responsible for
// receiving data from external sources or other components and managing the flow
// of data within the system.
type ReceivingRelay[T any] interface {
	// ---- Connections / plumbing ------------------------------------------------

	// ConnectLogger attaches one or more loggers to the Receiving Relay for operational logging.
	ConnectLogger(...Logger)

	// ConnectOutput connects one or more output submitters to the relay.
	ConnectOutput(...Submitter[T])

	// ---- Getters ----------------------------------------------------------------

	// GetAddress returns the current network address on which the relay is configured.
	GetAddress() string

	// GetComponentMetadata retrieves the metadata of the Receiving Relay.
	GetComponentMetadata() ComponentMetadata

	// ---- Basic configuration -----------------------------------------------------

	// SetDecryptionKey sets the symmetric key used to decrypt incoming AES-GCM payloads.
	SetDecryptionKey(string)

	// IsRunning reports whether the relay is actively processing data.
	IsRunning() bool

	// Listen starts listening on the configured address.
	// If listenForever is true, it will retry every retryInSeconds on failure.
	Listen(listenForever bool, retryInSeconds int) error

	// NotifyLoggers sends a formatted log message to all attached loggers.
	NotifyLoggers(level LogLevel, format string, args ...interface{})

	// Receive handles a single wrapped payload via a unary RPC.
	Receive(ctx context.Context, payload *relay.WrappedPayload) (*relay.StreamAcknowledgment, error)

	// SetComponentMetadata updates the relay's metadata.
	SetComponentMetadata(name string, id string)

	// SetAddress sets the listening address.
	SetAddress(string)

	// SetDataChannel sets (or resizes) the internal data channel buffer.
	SetDataChannel(uint32)

	// SetTLSConfig configures TLS for incoming connections.
	SetTLSConfig(*TLSConfig)

	// Start begins receiving operations.
	Start(context.Context) error

	// StreamReceive handles a server stream of payloads.
	StreamReceive(stream relay.RelayService_StreamReceiveServer) error

	// Stop halts receiving operations and cleans up resources.
	Stop()

	// ---- Optional authentication / authorization hooks --------------------------

	// SetAuthenticationOptions provides server-side expectations for transport/app-layer
	// auth (mirrors relay.AuthenticationOptions).
	SetAuthenticationOptions(*relay.AuthenticationOptions)

	// SetAuthInterceptors installs gRPC interceptors used to enforce/observe auth on inbound RPCs.
	// Either argument may be nil to skip that interceptor type.
	SetAuthInterceptors(unary grpc.UnaryServerInterceptor, stream grpc.StreamServerInterceptor)

	// SetStaticHeaders installs constant metadata keys the server expects (or propagates)
	// for downstream processing (e.g., correlation/tenant tags).
	SetStaticHeaders(map[string]string)

	// SetDynamicAuthValidator registers a callback to perform custom validation on inbound metadata.
	// Return nil to accept; return an error to reject the call with an auth failure.
	SetDynamicAuthValidator(func(ctx context.Context, md map[string]string) error)

	// SetAuthRequired toggles strict enforcement. When true, failed/absent credentials
	// should reject requests; when false, the server may log and accept (best-effort).
	SetAuthRequired(required bool)
}
