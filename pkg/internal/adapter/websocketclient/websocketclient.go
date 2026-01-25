package websocketclient

import (
	"context"
	"crypto/tls"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// WebSocketClientAdapter connects to a WebSocket endpoint and streams messages in/out.
type WebSocketClientAdapter[T any] struct {
	componentMetadata types.ComponentMetadata

	ctx context.Context

	configLock   sync.Mutex
	url          string
	headers      map[string]string
	format       string
	readLimit    int64
	writeTimeout time.Duration
	idleTimeout  time.Duration
	tlsConfig    *tls.Config
	tlsConfigErr error

	sensorsLock sync.Mutex
	sensors     []types.Sensor[T]

	loggersLock sync.Mutex
	loggers     []types.Logger
}

// NewWebSocketClientAdapter constructs a WebSocket client adapter with defaults.
func NewWebSocketClientAdapter[T any](ctx context.Context, options ...types.Option[types.WebSocketClientAdapter[T]]) types.WebSocketClientAdapter[T] {
	if ctx == nil {
		ctx = context.Background()
	}

	client := &WebSocketClientAdapter[T]{
		ctx: ctx,
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "WEBSOCKET_CLIENT",
		},
		headers:      make(map[string]string),
		format:       types.WebSocketFormatJSON,
		readLimit:    1 << 20,
		writeTimeout: 5 * time.Second,
	}

	for _, opt := range options {
		if opt == nil {
			continue
		}
		opt(client)
	}

	return client
}
