package websocket

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func WithLogger[T any](loggers ...types.Logger) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.ConnectLogger(loggers...) }
}

func WithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.ConnectSensor(sensors...) }
}

func WithAddress[T any](address string) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetAddress(address) }
}

func WithEndpoint[T any](path string) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetEndpoint(path) }
}

func WithHeader[T any](key, value string) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.AddHeader(key, value) }
}

func WithAllowedOrigins[T any](origins ...string) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetAllowedOrigins(origins...) }
}

func WithTokenQueryParam[T any](name string) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetTokenQueryParam(name) }
}

func WithMessageFormat[T any](format string) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetMessageFormat(format) }
}

func WithReadLimit[T any](limit int64) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetReadLimit(limit) }
}

func WithWriteTimeout[T any](timeout time.Duration) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetWriteTimeout(timeout) }
}

func WithIdleTimeout[T any](timeout time.Duration) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetIdleTimeout(timeout) }
}

func WithSendBuffer[T any](size int) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetSendBuffer(size) }
}

func WithMaxConnections[T any](max int) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetMaxConnections(max) }
}

func WithTLS[T any](cfg types.TLSConfig) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetTLSConfig(cfg) }
}

func WithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetAuthenticationOptions(opts) }
}

func WithStaticHeaders[T any](headers map[string]string) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetStaticHeaders(headers) }
}

func WithDynamicAuthValidator[T any](fn func(ctx context.Context, headers map[string]string) error) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetDynamicAuthValidator(fn) }
}

func WithAuthRequired[T any](required bool) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.SetAuthRequired(required) }
}

func WithOutputWire[T any](wires ...types.Wire[T]) types.Option[types.WebSocketServer[T]] {
	return func(s types.WebSocketServer[T]) { s.ConnectOutput(wires...) }
}
