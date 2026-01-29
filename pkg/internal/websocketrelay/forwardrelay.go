package websocketrelay

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ForwardRelay streams payloads to downstream receiving relays over WebSocket.
type ForwardRelay[T any] struct {
	Targets []string
	Input   []types.Receiver[T]

	ctx    context.Context
	cancel context.CancelFunc

	componentMetadata types.ComponentMetadata

	Loggers     []types.Logger
	loggersLock sync.Mutex

	TlsConfig *types.TLSConfig

	PerformanceOptions *relay.PerformanceOptions
	SecurityOptions    *relay.SecurityOptions
	EncryptionKey      string

	authOptions    *relay.AuthenticationOptions
	tokenSource    types.OAuth2TokenSource
	staticHeaders  map[string]string
	dynamicHeaders func(ctx context.Context) map[string]string
	authRequired   bool

	ackMode             relay.AckMode
	ackEveryN           uint32
	maxInFlight         uint32
	omitPayloadMetadata bool
	streamSendBuf       int
	maxMessageBytes     int
	dropOnFull          bool

	isRunning    int32
	configFrozen int32

	streamsMu sync.Mutex
	streams   map[string]*wsSession
	seq       uint64

	passthrough   bool
	payloadFormat string
	payloadType   string
}

// NewForwardRelay constructs a WebSocket forward relay with optional configuration.
func NewForwardRelay[T any](ctx context.Context, options ...types.Option[*ForwardRelay[T]]) *ForwardRelay[T] {
	ctx, cancel := context.WithCancel(ctx)
	fr := &ForwardRelay[T]{
		ctx:    ctx,
		cancel: cancel,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "WEBSOCKET_FORWARD_RELAY",
		},
		Loggers:             make([]types.Logger, 0),
		authRequired:        true,
		streamSendBuf:       8192,
		ackMode:             relay.AckMode_ACK_BATCH,
		ackEveryN:           1024,
		maxInFlight:         8192,
		omitPayloadMetadata: true,
		maxMessageBytes:     defaultMaxMessageBytes,
		dropOnFull:          false,
		payloadFormat:       "json",
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

// IsRunning reports whether the relay is running.
func (fr *ForwardRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&fr.isRunning) == 1
}

// Start begins relay operations.
func (fr *ForwardRelay[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&fr.configFrozen, 1)
	atomic.StoreInt32(&fr.isRunning, 1)
	for _, input := range fr.Input {
		go func(r types.Receiver[T]) {
			_ = r.Start(ctx)
			for {
				select {
				case <-ctx.Done():
					return
				case item, ok := <-r.GetOutputChannel():
					if !ok {
						return
					}
					_ = fr.Submit(ctx, item)
				}
			}
		}(input)
	}
	return nil
}

// Stop halts relay operations and closes resources.
func (fr *ForwardRelay[T]) Stop() {
	fr.logKV(types.InfoLevel, "WebSocket forward relay stopping",
		"event", "Stop",
		"result", "SUCCESS",
	)
	fr.cancel()
	fr.closeAllStreams("shutdown")
	atomic.StoreInt32(&fr.isRunning, 0)
}
