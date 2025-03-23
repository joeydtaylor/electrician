// options.go file
package httpserver

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithLogger attaches one or more loggers to the server.
func WithLogger[T any](loggers ...types.Logger) types.Option[types.HTTPServerAdapter[T]] {
	return func(srv types.HTTPServerAdapter[T]) {
		srv.ConnectLogger(loggers...)
	}
}

// WithSensor attaches one or more sensors to the server.
// Sensors can observe request events, errors, etc.
func WithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.HTTPServerAdapter[T]] {
	return func(srv types.HTTPServerAdapter[T]) {
		srv.ConnectSensor(sensors...)
	}
}

// WithAddress sets the network address on which the server will listen (e.g., ":8080").
func WithAddress[T any](address string) types.Option[types.HTTPServerAdapter[T]] {
	return func(srv types.HTTPServerAdapter[T]) {
		srv.SetAddress(address)
	}
}

// WithServerConfig configures the HTTP method (e.g. POST) and endpoint path (e.g. "/webhook") this server handles.
func WithServerConfig[T any](method, endpoint string) types.Option[types.HTTPServerAdapter[T]] {
	return func(srv types.HTTPServerAdapter[T]) {
		srv.SetServerConfig(method, endpoint)
	}
}

// WithHeader adds a default response header to all successful server responses.
func WithHeader[T any](key, value string) types.Option[types.HTTPServerAdapter[T]] {
	return func(srv types.HTTPServerAdapter[T]) {
		srv.AddHeader(key, value)
	}
}

// WithTimeout sets the read/write timeout for incoming requests,
// helping to avoid hanging connections (e.g., slowloris attacks).
func WithTimeout[T any](timeout time.Duration) types.Option[types.HTTPServerAdapter[T]] {
	return func(srv types.HTTPServerAdapter[T]) {
		srv.SetTimeout(timeout)
	}
}
