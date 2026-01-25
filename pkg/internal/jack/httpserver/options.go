package httpserver

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithLogger attaches one or more loggers to the server.
func WithLogger[T any](loggers ...types.Logger) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.ConnectLogger(loggers...)
	}
}

// WithSensor attaches one or more sensors to the server.
func WithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.ConnectSensor(sensors...)
	}
}

// WithAddress sets the network address on which the server will listen (e.g., ":8080").
func WithAddress[T any](address string) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetAddress(address)
	}
}

// WithServerConfig sets the HTTP method and endpoint.
func WithServerConfig[T any](method, endpoint string) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetServerConfig(method, endpoint)
	}
}

// WithHeader adds a default response header to all responses.
func WithHeader[T any](key, value string) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.AddHeader(key, value)
	}
}

// WithTimeout sets the read/write timeout for incoming requests.
func WithTimeout[T any](timeout time.Duration) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetTimeout(timeout)
	}
}

// WithTLS configures the server to use TLS if tlsCfg.UseTLS is true.
func WithTLS[T any](tlsCfg types.TLSConfig) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetTLSConfig(tlsCfg)
	}
}

// WithAuthenticationOptions configures OAuth2 authentication options.
func WithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetAuthenticationOptions(opts)
	}
}

// WithStaticHeaders enforces constant header key/value pairs on incoming requests.
func WithStaticHeaders[T any](headers map[string]string) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetStaticHeaders(headers)
	}
}

// WithDynamicAuthValidator registers a per-request validation callback.
func WithDynamicAuthValidator[T any](fn func(ctx context.Context, headers map[string]string) error) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetDynamicAuthValidator(fn)
	}
}

// WithAuthRequired toggles strict auth enforcement.
func WithAuthRequired[T any](required bool) types.Option[types.HTTPServer[T]] {
	return func(srv types.HTTPServer[T]) {
		srv.SetAuthRequired(required)
	}
}
