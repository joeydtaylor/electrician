package quicrelay

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"github.com/quic-go/quic-go"
)

// ReceivingRelay implements a QUIC relay server and forwards decoded items to submitters.
type ReceivingRelay[T any] struct {
	ctx    context.Context
	cancel context.CancelFunc

	Outputs []types.Submitter[T]

	componentMetadata types.ComponentMetadata
	Address           string

	DataCh chan T

	Loggers     []types.Logger
	loggersLock sync.Mutex

	TlsConfig *types.TLSConfig

	isRunning    int32
	configFrozen int32

	DecryptionKey string

	authOptions *relay.AuthenticationOptions

	staticHeaders map[string]string
	dynamicAuthValidator func(ctx context.Context, md map[string]string) error
	authRequired         bool

	passthrough bool

	listener *quic.Listener
	listenerMu sync.Mutex

	outputsOnce sync.Once

	maxFrameBytes int
}

// NewReceivingRelay initializes a QUIC receiving relay with optional configuration.
func NewReceivingRelay[T any](ctx context.Context, options ...types.Option[*ReceivingRelay[T]]) *ReceivingRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	rr := &ReceivingRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		DataCh: make(chan T, 1024),
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "QUIC_RECEIVING_RELAY",
		},
		Loggers:       make([]types.Logger, 0),
		DecryptionKey: "",
		staticHeaders: make(map[string]string),
		authRequired:  true,
		maxFrameBytes: defaultMaxFrameBytes,
	}

	for _, option := range options {
		option(rr)
	}

	return rr
}
