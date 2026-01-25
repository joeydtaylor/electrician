package httpserver

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
)

// httpServerAdapter implements types.HTTPServer.
type httpServerAdapter[T any] struct {
	componentMetadata types.ComponentMetadata

	address  string
	method   string
	endpoint string
	headers  map[string]string
	timeout  time.Duration

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

	configLock   sync.Mutex
	configFrozen int32

	server   *http.Server
	serverMu sync.Mutex
}

// NewHTTPServer constructs a new HTTP server adapter with defaults.
func NewHTTPServer[T any](ctx context.Context, options ...types.Option[types.HTTPServer[T]]) types.HTTPServer[T] {
	_ = ctx
	adapter := &httpServerAdapter[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "HTTP_SERVER",
		},
		timeout:       30 * time.Second,
		headers:       make(map[string]string),
		staticHeaders: make(map[string]string),
		authRequired:  true,
	}

	for _, opt := range options {
		opt(adapter)
	}

	return adapter
}

func (h *httpServerAdapter[T]) isFrozen() bool {
	return atomic.LoadInt32(&h.configFrozen) == 1
}
