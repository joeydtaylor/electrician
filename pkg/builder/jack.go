package builder

import (
	"context"
	"time"

	httpServerJack "github.com/joeydtaylor/electrician/pkg/internal/jack/httpserver"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
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

// HTTPServerWithAuthenticationOptions configures OAuth2 authentication options.
func HTTPServerWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithAuthenticationOptions[T](opts)
}

// HTTPServerWithStaticHeaders enforces constant header key/value pairs on incoming requests.
func HTTPServerWithStaticHeaders[T any](headers map[string]string) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithStaticHeaders[T](headers)
}

// HTTPServerWithDynamicAuthValidator registers a per-request validation callback.
func HTTPServerWithDynamicAuthValidator[T any](fn func(ctx context.Context, headers map[string]string) error) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithDynamicAuthValidator[T](fn)
}

// HTTPServerWithAuthRequired toggles strict auth enforcement.
func HTTPServerWithAuthRequired[T any](required bool) types.Option[types.HTTPServer[T]] {
	return httpServerJack.WithAuthRequired[T](required)
}

// NewHTTPServerOAuth2JWTOptions builds JWT validation settings for resource servers.
func NewHTTPServerOAuth2JWTOptions(
	issuer string,
	jwksURI string,
	audiences []string,
	scopes []string,
	jwksCacheSeconds int32,
) *relay.OAuth2Options {
	return httpServerJack.NewOAuth2JWTOptions(issuer, jwksURI, audiences, scopes, jwksCacheSeconds)
}

// NewHTTPServerOAuth2IntrospectionOptions builds RFC 7662 introspection settings.
func NewHTTPServerOAuth2IntrospectionOptions(
	introspectionURL string,
	authType string,
	clientID string,
	clientSecret string,
	bearerToken string,
	introspectionCacheSeconds int32,
) *relay.OAuth2Options {
	return httpServerJack.NewOAuth2IntrospectionOptions(introspectionURL, authType, clientID, clientSecret, bearerToken, introspectionCacheSeconds)
}

// NewHTTPServerOAuth2Forwarding toggles forwarding of inbound bearer tokens.
func NewHTTPServerOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return httpServerJack.NewOAuth2Forwarding(forward, forwardMetadataKey)
}

// NewHTTPServerMergeOAuth2Options merges non-zero fields from src into dst.
func NewHTTPServerMergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	return httpServerJack.MergeOAuth2Options(dst, src)
}

// NewHTTPServerAuthenticationOptionsOAuth2 composes AuthenticationOptions for OAuth2.
func NewHTTPServerAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return httpServerJack.NewAuthenticationOptionsOAuth2(oauth)
}

// NewHTTPServerAuthenticationOptionsNone composes a disabled auth options object.
func NewHTTPServerAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return httpServerJack.NewAuthenticationOptionsNone()
}
