package builder

import (
	"context"
	"time"

	websocketJack "github.com/joeydtaylor/electrician/pkg/internal/jack/websocket"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NewWebSocketServer creates a new WebSocket server with the specified options.
func NewWebSocketServer[T any](ctx context.Context, options ...types.Option[types.WebSocketServer[T]]) types.WebSocketServer[T] {
	return websocketJack.NewWebSocketServer[T](ctx, options...)
}

// WebSocketServerWithLogger attaches one or more loggers to the server.
func WebSocketServerWithLogger[T any](loggers ...types.Logger) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithLogger[T](loggers...)
}

// WebSocketServerWithSensor attaches one or more sensors to the server.
func WebSocketServerWithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithSensor[T](sensors...)
}

// WebSocketServerWithAddress sets the listen address (e.g., ":8080").
func WebSocketServerWithAddress[T any](address string) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithAddress[T](address)
}

// WebSocketServerWithEndpoint sets the WebSocket path (e.g., "/ws").
func WebSocketServerWithEndpoint[T any](endpoint string) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithEndpoint[T](endpoint)
}

// WebSocketServerWithHeader adds a response header to the handshake.
func WebSocketServerWithHeader[T any](key, value string) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithHeader[T](key, value)
}

// WebSocketServerWithAllowedOrigins sets acceptable Origin patterns.
func WebSocketServerWithAllowedOrigins[T any](origins ...string) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithAllowedOrigins[T](origins...)
}

// WebSocketServerWithTokenQueryParam configures a query parameter to extract bearer tokens.
func WebSocketServerWithTokenQueryParam[T any](name string) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithTokenQueryParam[T](name)
}

// WebSocketServerWithMessageFormat sets the encoding format.
func WebSocketServerWithMessageFormat[T any](format string) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithMessageFormat[T](format)
}

// WebSocketServerWithReadLimit sets the maximum inbound message size.
func WebSocketServerWithReadLimit[T any](limit int64) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithReadLimit[T](limit)
}

// WebSocketServerWithWriteTimeout sets the write timeout.
func WebSocketServerWithWriteTimeout[T any](timeout time.Duration) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithWriteTimeout[T](timeout)
}

// WebSocketServerWithIdleTimeout sets the read idle timeout.
func WebSocketServerWithIdleTimeout[T any](timeout time.Duration) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithIdleTimeout[T](timeout)
}

// WebSocketServerWithSendBuffer sets the outbound buffer size per connection.
func WebSocketServerWithSendBuffer[T any](size int) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithSendBuffer[T](size)
}

// WebSocketServerWithMaxConnections caps concurrent connections.
func WebSocketServerWithMaxConnections[T any](max int) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithMaxConnections[T](max)
}

// WebSocketServerWithTLS configures TLS for inbound connections.
func WebSocketServerWithTLS[T any](cfg types.TLSConfig) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithTLS[T](cfg)
}

// WebSocketServerWithAuthenticationOptions configures OAuth2 authentication options.
func WebSocketServerWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithAuthenticationOptions[T](opts)
}

// WebSocketServerWithStaticHeaders enforces constant header key/value pairs.
func WebSocketServerWithStaticHeaders[T any](headers map[string]string) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithStaticHeaders[T](headers)
}

// WebSocketServerWithDynamicAuthValidator registers a per-request validation callback.
func WebSocketServerWithDynamicAuthValidator[T any](fn func(ctx context.Context, headers map[string]string) error) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithDynamicAuthValidator[T](fn)
}

// WebSocketServerWithAuthRequired toggles strict auth enforcement.
func WebSocketServerWithAuthRequired[T any](required bool) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithAuthRequired[T](required)
}

// WebSocketServerWithOutputWire connects output wires for broadcast.
func WebSocketServerWithOutputWire[T any](wires ...types.Wire[T]) types.Option[types.WebSocketServer[T]] {
	return websocketJack.WithOutputWire[T](wires...)
}

// NewWebSocketOAuth2JWTOptions builds JWT validation settings for resource servers.
func NewWebSocketOAuth2JWTOptions(
	issuer string,
	jwksURI string,
	audiences []string,
	scopes []string,
	jwksCacheSeconds int32,
) *relay.OAuth2Options {
	return websocketJack.NewOAuth2JWTOptions(issuer, jwksURI, audiences, scopes, jwksCacheSeconds)
}

// NewWebSocketOAuth2IntrospectionOptions builds RFC 7662 introspection settings.
func NewWebSocketOAuth2IntrospectionOptions(
	introspectionURL string,
	authType string,
	clientID string,
	clientSecret string,
	bearerToken string,
	introspectionCacheSeconds int32,
) *relay.OAuth2Options {
	return websocketJack.NewOAuth2IntrospectionOptions(introspectionURL, authType, clientID, clientSecret, bearerToken, introspectionCacheSeconds)
}

// NewWebSocketOAuth2Forwarding toggles forwarding of inbound bearer tokens.
func NewWebSocketOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return websocketJack.NewOAuth2Forwarding(forward, forwardMetadataKey)
}

// NewWebSocketMergeOAuth2Options merges non-zero fields from src into dst.
func NewWebSocketMergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	return websocketJack.MergeOAuth2Options(dst, src)
}

// NewWebSocketAuthenticationOptionsOAuth2 composes AuthenticationOptions for OAuth2.
func NewWebSocketAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return websocketJack.NewAuthenticationOptionsOAuth2(oauth)
}

// NewWebSocketAuthenticationOptionsNone composes a disabled auth options object.
func NewWebSocketAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return websocketJack.NewAuthenticationOptionsNone()
}
