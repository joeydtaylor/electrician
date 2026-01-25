package httpserver

import (
	"time"

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
