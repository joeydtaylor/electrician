package builder

import (
	"context"
	"time"

	wsAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/websocketclient"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NewWebSocketClientAdapter creates a new WebSocket client adapter.
func NewWebSocketClientAdapter[T any](ctx context.Context, options ...types.Option[types.WebSocketClientAdapter[T]]) types.WebSocketClientAdapter[T] {
	return wsAdapter.NewWebSocketClientAdapter[T](ctx, options...)
}

// WebSocketClientAdapterWithLogger attaches loggers to the client.
func WebSocketClientAdapterWithLogger[T any](loggers ...types.Logger) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithLogger[T](loggers...)
}

// WebSocketClientAdapterWithSensor attaches sensors to the client.
func WebSocketClientAdapterWithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithSensor[T](sensors...)
}

// WebSocketClientAdapterWithURL sets the WebSocket URL.
func WebSocketClientAdapterWithURL[T any](url string) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithURL[T](url)
}

// WebSocketClientAdapterWithHeaders sets request headers.
func WebSocketClientAdapterWithHeaders[T any](headers map[string]string) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithHeaders[T](headers)
}

// WebSocketClientAdapterWithHeader adds a single header.
func WebSocketClientAdapterWithHeader[T any](key, value string) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithHeader[T](key, value)
}

// WebSocketClientAdapterWithMessageFormat sets the message encoding format.
func WebSocketClientAdapterWithMessageFormat[T any](format string) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithMessageFormat[T](format)
}

// WebSocketClientAdapterWithReadLimit sets the maximum inbound message size.
func WebSocketClientAdapterWithReadLimit[T any](limit int64) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithReadLimit[T](limit)
}

// WebSocketClientAdapterWithWriteTimeout sets the write timeout.
func WebSocketClientAdapterWithWriteTimeout[T any](timeout time.Duration) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithWriteTimeout[T](timeout)
}

// WebSocketClientAdapterWithIdleTimeout sets the read idle timeout.
func WebSocketClientAdapterWithIdleTimeout[T any](timeout time.Duration) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithIdleTimeout[T](timeout)
}

// WebSocketClientAdapterWithTLS configures TLS for outbound connections.
func WebSocketClientAdapterWithTLS[T any](cfg types.TLSConfig) types.Option[types.WebSocketClientAdapter[T]] {
	return wsAdapter.WithTLS[T](cfg)
}
