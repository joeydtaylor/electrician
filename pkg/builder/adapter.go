package builder

import (
	"context"
	"io"
	"time"

	httpClientAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/httpclient"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

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
