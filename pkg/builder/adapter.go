package builder

import (
	"context"
	"io"
	"time"

	httpClientAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/httpclient"
	httpServerAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/httpserver"

	internalLogger "github.com/joeydtaylor/electrician/pkg/internal/internallogger"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type LoggerOption = internalLogger.LoggerOption

type SinkConfig = types.SinkConfig

type SinkType = types.SinkType

type HTTPServerResponse = types.HTTPServerResponse

const (
	FileSink    SinkType = "file"
	StdoutSink  SinkType = "stdout"
	NetworkSink SinkType = "network"
)

func NewLogger(options ...internalLogger.LoggerOption) types.Logger {
	return internalLogger.NewLogger(options...)
}

// WithLevel configures the logger to use the specified log level
func LoggerWithLevel(levelStr string) LoggerOption {
	return internalLogger.LoggerWithLevel(levelStr)
}

// WithDevelopment enables or disables development mode
func LoggerWithDevelopment(dev bool) LoggerOption {
	return internalLogger.LoggerWithDevelopment(dev)
}

// NewHTTPClientAdapter creates a new HTTP plug with a base URL, custom headers, and an interval.
func NewHTTPClientAdapter[T any](ctx context.Context, options ...types.Option[types.HTTPClientAdapter[T]]) types.HTTPClientAdapter[T] {
	return httpClientAdapter.NewHTTPClientAdapter[T](ctx, options...)
}

// WithRequestConfig allows setting the request configuration.
func HTTPClientAdapterWithRequestConfig[T any](method, endpoint string, body io.Reader) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithRequestConfig[T](method, endpoint, body)
}

// WithRequestConfig allows setting the request configuration.
func HTTPClientAdapterWithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithSensor[T](sensor...)
}

// WithInterval sets the interval for making requests.
func HTTPClientAdapterWithInterval[T any](interval time.Duration) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithInterval[T](interval)
}

// WithHeader adds a header to the request.
func HTTPClientAdapterWithHeader[T any](key, value string) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithHeader[T](key, value)
}

// WithHeader adds a header to the request.
func HTTPClientAdapterWithLogger[T any](l ...types.Logger) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithLogger[T](l...)
}

// WithInterval sets the interval for making requests.
func HTTPClientAdapterWithMaxRetries[T any](maxRetries int) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithMaxRetries[T](maxRetries)
}

// WithTimeout sets the timeout for HTTP requests.
func HTTPClientAdapterWithTimeout[T any](timeout time.Duration) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithTimeout[T](timeout)
}

// WithContentType sets the Content-Type header for outgoing requests.
func HTTPClientAdapterWithContentType[T any](contentType string) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithContentType[T](contentType)
}

// WithUserAgent sets the User-Agent header for outgoing requests.
func HTTPClientAdapterWithUserAgent[T any](userAgent string) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithUserAgent[T](userAgent)
}

// WithTLSPinning enables TLS pinning and sets the expected certificate.
func HTTPClientAdapterWithTLSPinning[T any](certPath string) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithTLSPinning[T](certPath)
}

func HTTPClientAdapterWithOAuth2ClientCredentials[T any](clientID, clientSecret, tokenURL string, audience string, scopes ...string) types.Option[types.HTTPClientAdapter[T]] {
	return httpClientAdapter.WithOAuth2ClientCredentials[T](clientID, clientSecret, tokenURL, audience, scopes...)
}

// NewHTTPServerAdapter creates a new HTTPServerAdapter with the specified options.
func NewHTTPServerAdapter[T any](ctx context.Context, options ...types.Option[types.HTTPServerAdapter[T]]) types.HTTPServerAdapter[T] {
	return httpServerAdapter.NewHTTPServerAdapter[T](ctx, options...)
}

// HTTPServerAdapterWithLogger attaches one or more loggers to the server.
func HTTPServerAdapterWithLogger[T any](loggers ...types.Logger) types.Option[types.HTTPServerAdapter[T]] {
	return httpServerAdapter.WithLogger[T](loggers...)
}

// HTTPServerAdapterWithSensor attaches one or more sensors to the server.
func HTTPServerAdapterWithSensor[T any](sensors ...types.Sensor[T]) types.Option[types.HTTPServerAdapter[T]] {
	return httpServerAdapter.WithSensor[T](sensors...)
}

// HTTPServerAdapterWithAddress sets the IP/port on which the server will listen (e.g., ":8080").
func HTTPServerAdapterWithAddress[T any](address string) types.Option[types.HTTPServerAdapter[T]] {
	return httpServerAdapter.WithAddress[T](address)
}

// HTTPServerAdapterWithServerConfig sets the HTTP method and endpoint (e.g. "POST" and "/webhook").
func HTTPServerAdapterWithServerConfig[T any](method, endpoint string) types.Option[types.HTTPServerAdapter[T]] {
	return httpServerAdapter.WithServerConfig[T](method, endpoint)
}

// HTTPServerAdapterWithHeader adds a default response header to all successful server responses.
func HTTPServerAdapterWithHeader[T any](key, value string) types.Option[types.HTTPServerAdapter[T]] {
	return httpServerAdapter.WithHeader[T](key, value)
}

// HTTPServerAdapterWithTimeout sets the read/write timeout for incoming requests.
func HTTPServerAdapterWithTimeout[T any](timeout time.Duration) types.Option[types.HTTPServerAdapter[T]] {
	return httpServerAdapter.WithTimeout[T](timeout)
}

// HTTPServerAdapterWithTLS configures the server to use TLS if tlsCfg.UseTLS == true.
// Otherwise, it reverts to plain HTTP (no TLS).
func HTTPServerAdapterWithTLS[T any](tlsCfg types.TLSConfig) types.Option[types.HTTPServerAdapter[T]] {
	return httpServerAdapter.WithTLS[T](tlsCfg)
}
