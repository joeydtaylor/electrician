package websocketclient

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithLogger attaches loggers to the client.
func WithLogger[T any](loggers ...types.Logger) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.ConnectLogger(loggers...) }
}

// WithSensor attaches sensors to the client.
func WithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.ConnectSensor(sensors...) }
}

// WithURL sets the WebSocket URL.
func WithURL[T any](url string) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.SetURL(url) }
}

// WithHeaders sets request headers.
func WithHeaders[T any](headers map[string]string) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.SetHeaders(headers) }
}

// WithHeader adds a single header.
func WithHeader[T any](key, value string) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.AddHeader(key, value) }
}

// WithMessageFormat sets the message encoding format.
func WithMessageFormat[T any](format string) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.SetMessageFormat(format) }
}

// WithReadLimit sets the max inbound message size.
func WithReadLimit[T any](limit int64) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.SetReadLimit(limit) }
}

// WithWriteTimeout sets the write timeout.
func WithWriteTimeout[T any](timeout time.Duration) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.SetWriteTimeout(timeout) }
}

// WithIdleTimeout sets the read idle timeout.
func WithIdleTimeout[T any](timeout time.Duration) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.SetIdleTimeout(timeout) }
}

// WithTLS configures TLS for the WebSocket client.
func WithTLS[T any](cfg types.TLSConfig) types.Option[types.WebSocketClientAdapter[T]] {
	return func(c types.WebSocketClientAdapter[T]) { c.SetTLSConfig(cfg) }
}
