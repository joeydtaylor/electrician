package httpserver

import (
	"context"
	"crypto/tls"
	"net/http"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// httpServerAdapter implements types.HTTPServer[T].
type httpServerAdapter[T any] struct {
	componentMetadata types.ComponentMetadata

	// Basic config
	address  string // e.g., ":8080"
	method   string // e.g., "POST"
	endpoint string // e.g., "/webhook"
	headers  map[string]string

	timeout   time.Duration
	tlsConfig *tls.Config // If set, we serve TLS using this config.

	// Logging & Sensors
	loggers     []types.Logger
	sensors     []types.Sensor[T]
	loggersLock sync.Mutex
	sensorsLock sync.Mutex

	// Underlying http.Server
	server   *http.Server
	serverMu sync.Mutex
}

// NewHTTPServer constructs a new HTTP server adapter with sensible defaults.
func NewHTTPServer[T any](
	ctx context.Context,
	options ...types.Option[types.HTTPServer[T]],
) types.HTTPServer[T] {

	// Create an instance with default values
	h := &httpServerAdapter[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(),
			Type: "HTTP_SERVER",
		},
		timeout: 30 * time.Second,
		headers: make(map[string]string),
	}

	// Apply functional options
	for _, opt := range options {
		opt(h)
	}
	return h
}
