package builder

import (
	"context"
	"time"

	httpServerJack "github.com/joeydtaylor/electrician/pkg/internal/jack/httpserver"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NewHTTPServer creates a new HTTPServer with the specified options.
func NewHTTPServer[T any](ctx context.Context, options ...types.Option[types.HTTPServer[T]]) types.HTTPServer[T] {
	return httpServerJack.NewHTTPServer[T](ctx, options...)
}

// HTTPServerWithLogger attaches one or more loggers to the server.
func HTTPServerWithLogger[T any](loggers ...types.Logger) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithLogger[T](loggers...)
}

// HTTPServerWithSensor attaches one or more sensors to the server.
func HTTPServerWithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithSensor[T](sensors...)
}

// HTTPServerWithAddress sets the IP/port on which the server will listen (e.g., ":8080").
func HTTPServerWithAddress[T any](address string) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithAddress[T](address)
}

// HTTPServerWithServerConfig sets the HTTP method and endpoint (e.g. "POST" and "/webhook").
func HTTPServerWithServerConfig[T any](method, endpoint string) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithServerConfig[T](method, endpoint)
}

// HTTPServerWithHeader adds a default response header to all successful server responses.
func HTTPServerWithHeader[T any](key, value string) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithHeader[T](key, value)
}

// HTTPServerWithTimeout sets the read/write timeout for incoming requests.
func HTTPServerWithTimeout[T any](timeout time.Duration) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithTimeout[T](timeout)
}

// HTTPServerWithTLS configures the server to use TLS if tlsCfg.UseTLS == true.
// Otherwise, it reverts to plain HTTP (no TLS).
func HTTPServerWithTLS[T any](tlsCfg types.TLSConfig) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithTLS[T](tlsCfg)
}
