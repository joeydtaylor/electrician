// Package forwardrelay implements a gRPC client relay that forwards wrapped payloads to a receiving relay.
package forwardrelay

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ForwardRelay forwards payloads to one or more relay targets.
type ForwardRelay[T any] struct {
	Targets            []string
	Input              []types.Receiver[T]
	ctx                context.Context
	cancel             context.CancelFunc
	componentMetadata  types.ComponentMetadata
	Loggers            []types.Logger
	loggersLock        sync.Mutex
	TlsConfig          *types.TLSConfig
	PerformanceOptions *relay.PerformanceOptions

	SecurityOptions *relay.SecurityOptions
	EncryptionKey   string

	authOptions    *relay.AuthenticationOptions
	tokenSource    types.OAuth2TokenSource
	staticHeaders  map[string]string
	dynamicHeaders func(ctx context.Context) map[string]string

	isRunning            int32
	tlsCredentials       atomic.Value
	tlsCredentialsUpdate sync.Mutex
	configFrozen         int32
	authRequired         bool

	streamsMu     sync.Mutex
	streams       map[string]*streamSession
	seq           uint64
	streamSendBuf int
}

// NewForwardRelay constructs a forward relay with optional configuration.
func NewForwardRelay[T any](ctx context.Context, options ...types.Option[types.ForwardRelay[T]]) types.ForwardRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	fr := &ForwardRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "FORWARD_RELAY",
		},
		Loggers:   make([]types.Logger, 0),
		isRunning: 0,
		PerformanceOptions: &relay.PerformanceOptions{
			UseCompression:       false,
			CompressionAlgorithm: COMPRESS_NONE,
		},
		authRequired: true,
	}

	for _, option := range options {
		option(fr)
	}

	return fr
}
