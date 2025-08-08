package receivingrelay

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"google.golang.org/grpc"
)

// ReceivingRelay represents a component in a distributed system that receives data from forward relays
// or similar data sources.
type ReceivingRelay[T any] struct {
	relay.UnimplementedRelayServiceServer // Embedding the unimplemented server

	// Core lifecycle management
	ctx    context.Context
	cancel context.CancelFunc

	// Outbound data flow
	Outputs []types.Submitter[T]

	// Identifying info
	componentMetadata types.ComponentMetadata

	// Where we listen for inbound data
	Address string

	// Channel through which received data is passed to the processing functions
	DataCh chan T

	// Logging
	Loggers     []types.Logger
	loggersLock *sync.Mutex

	// TLS for secure communication
	TlsConfig    *types.TLSConfig
	isRunning    int32 // Atomic, 1 if running
	configFrozen int32

	// Content decryption (AES-GCM) for payloads, if enabled by metadata.
	DecryptionKey string

	// ---------------- Authentication / Authorization ----------------
	// Optional auth hints mirrored from proto MessageMetadata.authentication.
	authOptions *relay.AuthenticationOptions

	// Optional unary interceptor to enforce OAuth2/mTLS auth before handler logic.
	// Typical implementation validates Bearer via JWKS/introspection or checks client cert.
	authUnary grpc.UnaryServerInterceptor

	// Optional constant metadata requirements (e.g., tenant headers).
	staticHeaders map[string]string

	// Optional per-request validator. Runs early with incoming metadata.
	// Return error to reject request (e.g., scope/audience check, header policy).
	dynamicAuthValidator func(ctx context.Context, md map[string]string) error
}

// NewReceivingRelay initializes and returns a new instance of ReceivingRelay with configuration options applied.
func NewReceivingRelay[T any](ctx context.Context, options ...types.Option[types.ReceivingRelay[T]]) types.ReceivingRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	rr := &ReceivingRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		DataCh: make(chan T),

		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "RECEIVING_RELAY",
		},

		Loggers:       make([]types.Logger, 0),
		loggersLock:   new(sync.Mutex),
		DecryptionKey: "",

		// auth fields default to nil/zero; optional features are off unless configured.
	}

	// Apply provided options
	for _, option := range options {
		option(rr)
	}

	// If Outputs are configured, automatically forward data from DataCh to each output
	if rr.Outputs != nil {
		go func() {
			for data := range rr.DataCh {
				for _, output := range rr.Outputs {
					output.Submit(ctx, data)
				}
			}
		}()
	}

	return rr
}
