package builder

import (
	"context"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	httpClientAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/httpclient"
	s3ClientAdapter "github.com/joeydtaylor/electrician/pkg/internal/adapter/s3client"

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

// NewS3ClientAdapter creates a new S3 client adapter (read + write capable).
func NewS3ClientAdapter[T any](ctx context.Context, options ...types.S3ClientOption[T]) types.S3ClientAdapter[T] {
	return s3ClientAdapter.NewS3ClientAdapter[T](ctx, options...)
}

func S3ClientAdapterWithS3ClientDeps[T any](deps types.S3ClientDeps) types.S3ClientOption[T] {
	return s3ClientAdapter.WithS3ClientDeps[T](deps)
}

func S3ClientAdapterWithWriterConfig[T any](cfg types.S3WriterConfig) types.S3ClientOption[T] {
	return s3ClientAdapter.WithWriterConfig[T](cfg)
}

func S3ClientAdapterWithReaderConfig[T any](cfg types.S3ReaderConfig) types.S3ClientOption[T] {
	return s3ClientAdapter.WithReaderConfig[T](cfg)
}

func S3ClientAdapterWithSensor[T any](sensor ...types.Sensor[T]) types.S3ClientOption[T] {
	return s3ClientAdapter.WithSensor[T](sensor...)
}

func S3ClientAdapterWithLogger[T any](l ...types.Logger) types.S3ClientOption[T] {
	return s3ClientAdapter.WithLogger[T](l...)
}

func S3ClientAdapterWithBatchSettings[T any](maxRecords, maxBytes int, maxAge time.Duration) types.S3ClientOption[T] {
	return s3ClientAdapter.WithBatchSettings[T](maxRecords, maxBytes, maxAge)
}

func S3ClientAdapterWithFormat[T any](format, compression string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithFormat[T](format, compression)
}

func S3ClientAdapterWithSSE[T any](mode, kmsKey string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithSSE[T](mode, kmsKey)
}

func S3ClientAdapterWithReaderListSettings[T any](prefix, startAfter string, pageSize int32, pollEvery time.Duration) types.S3ClientOption[T] {
	return s3ClientAdapter.WithReaderListSettings[T](prefix, startAfter, pageSize, pollEvery)
}

// Inject AWS client + bucket without exposing internal types.
func S3ClientAdapterWithClientAndBucket[T any](cli *s3.Client, bucket string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithS3ClientDeps[T](types.S3ClientDeps{
		Client:         cli,
		Bucket:         bucket,
		ForcePathStyle: true, // good default for LocalStack/MinIO; callers can override via a separate option if needed
	})
}

// S3ClientAdapterWithWriterPrefixTemplate sets the writer's prefix template without exposing internal types.
func S3ClientAdapterWithWriterPrefixTemplate[T any](prefix string) types.S3ClientOption[T] {
	return s3ClientAdapter.WithWriterConfig[T](types.S3WriterConfig{
		PrefixTemplate: prefix,
	})
}
