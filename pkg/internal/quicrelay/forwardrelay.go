package quicrelay

import (
	"context"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

const quicALPN = "electrician-quic-relay-v1"

// ForwardRelay streams payloads to downstream receiving relays over QUIC.
type ForwardRelay[T any] struct {
	Targets []string
	Input   []types.Receiver[T]

	ctx    context.Context
	cancel context.CancelFunc

	componentMetadata types.ComponentMetadata

	Loggers     []types.Logger
	loggersLock sync.Mutex

	TlsConfig          *types.TLSConfig
	PerformanceOptions *relay.PerformanceOptions
	SecurityOptions    *relay.SecurityOptions
	EncryptionKey      string

	authOptions   *relay.AuthenticationOptions
	tokenSource   types.OAuth2TokenSource
	staticHeaders map[string]string
	dynamicHeaders func(ctx context.Context) map[string]string
	authRequired  bool

	ackMode             relay.AckMode
	ackEveryN           uint32
	maxInFlight         uint32
	omitPayloadMetadata bool
	streamSendBuf       int
	maxFrameBytes       int
	dropOnFull          bool

	isRunning    int32
	configFrozen int32

	streamsMu sync.Mutex
	streams   map[string]*streamSession
	seq       uint64

	passthrough bool
}

// NewForwardRelay constructs a QUIC forward relay with optional configuration.
func NewForwardRelay[T any](ctx context.Context, options ...types.Option[*ForwardRelay[T]]) *ForwardRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	fr := &ForwardRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "QUIC_FORWARD_RELAY",
		},
		Loggers:       make([]types.Logger, 0),
		authRequired:  true,
		streamSendBuf: 8192,
		ackMode:       relay.AckMode_ACK_BATCH,
		ackEveryN:     1024,
		maxInFlight:   8192,
		omitPayloadMetadata: true,
		maxFrameBytes:       defaultMaxFrameBytes,
		dropOnFull:          false,
		PerformanceOptions: &relay.PerformanceOptions{
			UseCompression:       false,
			CompressionAlgorithm: COMPRESS_NONE,
		},
		staticHeaders: make(map[string]string),
	}

	for _, option := range options {
		option(fr)
	}

	return fr
}
