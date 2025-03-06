package httpclient

import (
	"io"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithTLSPinning enables TLS pinning and sets the expected certificate.
func WithTLSPinning[T any](certPath string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetTlsPinnedCertificate(certPath)
	}
}

// WithTimeout sets the timeout for HTTP requests.
func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.ConnectSensor(sensor...)
	}
}

// WithRequestConfig allows setting the request configuration.
func WithLogger[T any](l ...types.Logger) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.ConnectLogger(l...)
	}
}

// WithRequestConfig allows setting the request configuration.
func WithRequestConfig[T any](method, endpoint string, body io.Reader) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetRequestConfig(method, endpoint, body)
	}
}

// WithInterval sets the interval for making requests.
func WithInterval[T any](interval time.Duration) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetInterval(interval)
	}
}

// WithInterval sets the interval for making requests.
func WithMaxRetries[T any](maxRetries int) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetMaxRetries(maxRetries)
	}
}

// WithHeader adds a header to the request.
func WithHeader[T any](key, value string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.AddHeader(key, value)
	}
}

func WithOAuth2ClientCredentials[T any](clientID, clientSecret, tokenURL string, audience string, scopes ...string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetOAuth2Config(clientID, clientSecret, tokenURL, audience, scopes...)
		token, err := hp.GetTokenFromOAuthServer(clientID, clientSecret, tokenURL, audience, scopes...)
		if err != nil {
			hp.NotifyLoggers(types.ErrorLevel, "Failed to set OAuth2 credentials: %v", err)
			return
		}
		hp.AddHeader("Authorization", "Bearer "+token)
	}
}

// WithTimeout sets the timeout for HTTP requests.
func WithTimeout[T any](timeout time.Duration) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetTimeout(timeout)
	}
}

// WithContentType sets the Content-Type header for outgoing requests.
func WithContentType[T any](contentType string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.AddHeader("Content-Type", contentType)
	}
}

// WithUserAgent sets the User-Agent header for outgoing requests.
func WithUserAgent[T any](userAgent string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.AddHeader("User-Agent", userAgent)
	}
}
