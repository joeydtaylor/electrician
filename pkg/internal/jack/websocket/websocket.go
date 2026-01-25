package websocket

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"nhooyr.io/websocket"
)

type serverAdapter[T any] struct {
	componentMetadata types.ComponentMetadata

	address         string
	endpoint        string
	headers         map[string]string
	allowedOrigins  []string
	tokenQueryParam string
	readLimit       int64
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	sendBuffer      int
	maxConnections  int
	format          string

	tlsConfig    *tls.Config
	tlsConfigErr error

	loggers     []types.Logger
	sensors     []types.Sensor[T]
	loggersLock sync.Mutex
	sensorsLock sync.Mutex

	authOptions          *relay.AuthenticationOptions
	staticHeaders        map[string]string
	dynamicAuthValidator func(ctx context.Context, headers map[string]string) error
	authRequired         bool

	outputWires []types.Wire[T]

	configLock   sync.Mutex
	configFrozen int32

	server   *http.Server
	serverMu sync.Mutex

	conns   map[*websocket.Conn]*wsConn
	connsMu sync.Mutex
}

// NewWebSocketServer constructs a new WebSocket server with defaults.
func NewWebSocketServer[T any](ctx context.Context, options ...types.Option[types.WebSocketServer[T]]) types.WebSocketServer[T] {
	_ = ctx
	s := &serverAdapter[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "WEBSOCKET_SERVER",
		},
		address:        ":8080",
		endpoint:       "/ws",
		headers:        make(map[string]string),
		staticHeaders:  make(map[string]string),
		readLimit:      1 << 20,
		writeTimeout:   5 * time.Second,
		sendBuffer:     256,
		maxConnections: 0,
		format:         types.WebSocketFormatJSON,
		authRequired:   true,
		conns:          make(map[*websocket.Conn]*wsConn),
	}

	for _, opt := range options {
		opt(s)
	}

	return s
}

func (s *serverAdapter[T]) isFrozen() bool {
	return atomic.LoadInt32(&s.configFrozen) == 1
}
