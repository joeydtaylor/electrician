// Package forwardrelay implements a relay component for forwarding data to another relay.
// It facilitates the transmission of data between relay nodes in a distributed system,
// enabling efficient and reliable communication between different components or systems.
//
// The ForwardRelay struct defines the properties and methods of the forward relay:
// - Address: Address of the receiving relay to which data will be forwarded.
// - Input: List of receivers for receiving data to be forwarded by the relay.
// - ctx: Context for managing cancellation and synchronization of relay operations.
// - cancel: Function to cancel the context associated with the relay.
// - componentMetadata: Metadata of the forward relay component, including identifiers and type information.
// - Loggers: List of loggers attached to the forward relay for logging events and messages.
// - loggersLock: Mutex for protecting access to the loggers slice.
// - TlsConfig: TLS configuration for secure communication with the receiving relay.
// - PerformanceOptions: Options for configuring the performance of the relay, such as compression settings.
// - isRunning: Atomic flag indicating whether the relay is currently running.
// - tlsCredentials: Atomic value for storing TLS credentials used for secure communication.
// - tlsCredentialsUpdate: Mutex for controlling access to TLS credentials during updates.
//
// NewForwardRelay creates a new instance of a ForwardRelay with the provided context and optional configuration options.
// It initializes the forward relay with the specified address, input receivers, loggers, TLS configuration, and other settings.
// The function returns a types.ForwardRelay[T] interface, allowing users to interact with the forward relay through a generic interface.
//
// Parameters:
//   - ctx: Context for managing cancellation and synchronization of relay operations.
//   - options: Optional configuration options for customizing the behavior of the forward relay.
//
// Returns:
//   - types.ForwardRelay[T]: An instance of the ForwardRelay component configured according to the provided options.
package forwardrelay

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ForwardRelay represents a relay for forwarding data to another relay.
// It manages the transmission of data between relay nodes, providing reliability and efficiency
// in distributed systems communication.
type ForwardRelay[T any] struct {
	Targets            []string                  // Addresses of the receiving relay(s).
	Input              []types.Receiver[T]       // List of receivers for incoming data.
	ctx                context.Context           // Context for managing relay operations.
	cancel             context.CancelFunc        // Function to cancel relay operations.
	componentMetadata  types.ComponentMetadata   // Metadata of the relay component.
	Loggers            []types.Logger            // List of loggers for event logging.
	loggersLock        *sync.Mutex               // Mutex for loggers slice access control.
	TlsConfig          *types.TLSConfig          // TLS configuration for secure communication.
	PerformanceOptions *relay.PerformanceOptions // Performance options for relay operations.

	// ----- Security (content encryption) -----
	SecurityOptions *relay.SecurityOptions // Security options (e.g. AES-GCM).
	EncryptionKey   string                 // Encryption key for secure communication.

	// ----- Authentication (transport/application layer) -----
	authOptions    *relay.AuthenticationOptions                // Auth hints to send in proto metadata.
	tokenSource    types.OAuth2TokenSource                     // Optional OAuth2 per-RPC token source.
	staticHeaders  map[string]string                           // Constant gRPC metadata headers.
	dynamicHeaders func(ctx context.Context) map[string]string // Per-request metadata callback.

	// ----- Runtime State -----
	isRunning            int32        // Atomic flag indicating running state.
	tlsCredentials       atomic.Value // Atomic value for TLS credentials.
	tlsCredentialsUpdate sync.Mutex   // Mutex for TLS credentials update.
	configFrozen         int32        // Indicates whether the relay's configuration has been frozen.
	authRequired         bool
}

// NewForwardRelay creates a new forward relay instance with the specified context and options.
// It initializes the relay with the provided context and applies optional configuration options.
// The function returns a types.ForwardRelay[T] interface for interacting with the forward relay.
func NewForwardRelay[T any](ctx context.Context, options ...types.Option[types.ForwardRelay[T]]) types.ForwardRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	fr := &ForwardRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "FORWARD_RELAY",
		},
		Loggers:     make([]types.Logger, 0),
		loggersLock: new(sync.Mutex),
		isRunning:   0,
		PerformanceOptions: &relay.PerformanceOptions{
			UseCompression:       false,
			CompressionAlgorithm: COMPRESS_NONE,
		},
		SecurityOptions: nil, // No security by default
		EncryptionKey:   "",  // No key by default
		authRequired:    true,
	}

	// Apply any user-provided options
	for _, option := range options {
		option(fr)
	}

	return fr
}
