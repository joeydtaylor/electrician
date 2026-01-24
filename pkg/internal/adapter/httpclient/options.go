package httpclient

import (
	"io"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithContentType sets the Content-Type header for outgoing requests.
func WithContentType[T any](contentType string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.AddHeader("Content-Type", contentType)
	}
}

// WithHeader adds a static header to outgoing requests.
func WithHeader[T any](key, value string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.AddHeader(key, value)
	}
}

// WithInterval sets the polling interval used by Serve.
func WithInterval[T any](interval time.Duration) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetInterval(interval)
	}
}

// WithLogger attaches loggers to the adapter.
func WithLogger[T any](l ...types.Logger) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.ConnectLogger(l...)
	}
}

// WithMaxRetries sets the maximum number of retry attempts.
func WithMaxRetries[T any](maxRetries int) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetMaxRetries(maxRetries)
	}
}

// WithOAuth2ClientCredentials configures OAuth2 client credentials and pre-fetches a token.
func WithOAuth2ClientCredentials[T any](clientID, clientSecret, tokenURL string, audience string, scopes ...string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetOAuth2Config(clientID, clientSecret, tokenURL, audience, scopes...)
		token, err := hp.GetTokenFromOAuthServer(clientID, clientSecret, tokenURL, audience, scopes...)
		if err != nil {
			hp.NotifyLoggers(types.ErrorLevel, "failed to set OAuth2 credentials: %v", err)
			return
		}
		hp.AddHeader("Authorization", "Bearer "+token)
	}
}

// WithRequestConfig sets the request method, endpoint, and body.
func WithRequestConfig[T any](method, endpoint string, body io.Reader) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetRequestConfig(method, endpoint, body)
	}
}

// WithSensor attaches sensors to observe HTTP events.
func WithSensor[T any](sensor ...types.Sensor[T]) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.ConnectSensor(sensor...)
	}
}

// WithTimeout sets the per-request timeout.
func WithTimeout[T any](timeout time.Duration) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetTimeout(timeout)
	}
}

// WithTLSPinning enables TLS pinning with the provided certificate path.
func WithTLSPinning[T any](certPath string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.SetTlsPinnedCertificate(certPath)
	}
}

// WithUserAgent sets the User-Agent header for outgoing requests.
func WithUserAgent[T any](userAgent string) types.Option[types.HTTPClientAdapter[T]] {
	return func(hp types.HTTPClientAdapter[T]) {
		hp.AddHeader("User-Agent", userAgent)
	}
}
