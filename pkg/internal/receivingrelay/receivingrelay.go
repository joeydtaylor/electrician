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
	relay.UnimplementedRelayServiceServer // Embed unimplemented gRPC server

	// ---- Lifecycle ----
	ctx    context.Context
	cancel context.CancelFunc

	// ---- Outbound data flow ----
	Outputs []types.Submitter[T]

	// ---- Identity / config ----
	componentMetadata types.ComponentMetadata
	Address           string

	// ---- Data channel ----
	DataCh chan T

	// ---- Logging ----
	Loggers     []types.Logger
	loggersLock *sync.Mutex

	// ---- TLS / runtime ----
	TlsConfig    *types.TLSConfig
	isRunning    int32 // set atomically elsewhere
	configFrozen int32

	// ---- Content decryption (AES-GCM) ----
	DecryptionKey string

	// ---- Authentication / Authorization (optional) ----
	authOptions *relay.AuthenticationOptions

	// Interceptors used to enforce/observe auth on inbound RPCs.
	authUnary  grpc.UnaryServerInterceptor
	authStream grpc.StreamServerInterceptor

	// Static metadata expectations (tenant/correlation tags, etc.)
	staticHeaders map[string]string

	// Per-request validator hook (runs early with parsed metadata map).
	dynamicAuthValidator func(ctx context.Context, md map[string]string) error

	// Strictness toggle: when false, auth failures can be logged and allowed (best-effort).
	authRequired bool
}

// NewReceivingRelay initializes and returns a new instance of ReceivingRelay with configuration options applied.
func NewReceivingRelay[T any](ctx context.Context, options ...types.Option[types.ReceivingRelay[T]]) types.ReceivingRelay[T] {
	ctx, cancel := context.WithCancel(ctx)

	rr := &ReceivingRelay[T]{
		ctx:    ctx,
		cancel: cancel,

		// Buffered a bit so slow outputs don't immediately stall handlers; tune later if needed.
		DataCh: make(chan T, 1024),

		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "RECEIVING_RELAY",
		},

		Loggers:              make([]types.Logger, 0),
		loggersLock:          new(sync.Mutex),
		DecryptionKey:        "",
		staticHeaders:        make(map[string]string),
		authRequired:         true, // default to strict; can be relaxed via option
		authUnary:            nil,
		authStream:           nil,
		authOptions:          nil,
		TlsConfig:            nil,
		Outputs:              nil,
		Address:              "",
		isRunning:            0,
		configFrozen:         0,
		dynamicAuthValidator: nil,
	}

	// Apply provided options
	for _, option := range options {
		option(rr)
	}

	// If Outputs are configured, forward DataCh to each output.
	if len(rr.Outputs) > 0 {
		go func() {
			for data := range rr.DataCh {
				for _, out := range rr.Outputs {
					// Intentionally ignore errors here; handlers should log on failure.
					_ = out.Submit(ctx, data)
				}
			}
		}()
	}

	return rr
}
