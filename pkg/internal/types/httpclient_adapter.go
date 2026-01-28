package types

import (
	"context"
	"fmt"
	"io"
	"time"
)

// WrappedResponse is a generic wrapper for various types of HTTP response payloads.
type WrappedResponse[T any] struct {
	Data       T                   `json:"data"`
	Raw        []byte              `json:"raw,omitempty"` // Raw data for binary or non-struct responses
	Error      error               `json:"error,omitempty"`
	Headers    map[string][]string `json:"headers"`
	StatusCode int
}

// NewWrappedResponse creates a new instance of WrappedResponse with provided data.
func NewWrappedResponse[T any](data T, raw []byte, err error, headers map[string][]string) *WrappedResponse[T] {
	return &WrappedResponse[T]{
		Data:    data,
		Raw:     raw,
		Error:   err,
		Headers: headers,
	}
}

type HttpResponse[T any] struct {
	StatusCode int
	Body       T
}

type HTTPError struct {
	StatusCode int
	Err        error
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s, %v", e.StatusCode, e.Message, e.Err)
}

type HTTPClientAdapter[T any] interface {
	SetTlsPinnedCertificate(certPath string)
	SetBasicAuth(username, password string)
	SetMaxRetries(retries int)
	Serve(ctx context.Context, submitFunc func(ctx context.Context, elem T) error) error
	Fetch() (HttpResponse[T], error)
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	SetInterval(interval time.Duration)
	SetRequestConfig(method, endpoint string, body io.Reader)
	AddHeader(key, value string)
	// GetComponentMetadata retrieves metadata about the Forward Relay, such as its ID, name,
	// and type. This metadata can be used for identification, logging, and monitoring purposes.
	GetComponentMetadata() ComponentMetadata

	// SetComponentMetadata sets the metadata for the Forward Relay, such as its name and ID.
	SetComponentMetadata(name string, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
	GetTokenFromOAuthServer(clientID, clientSecret, tokenURL string, audience string, scopes ...string) (string, error)
	SetTimeout(timeout time.Duration)
	SetOAuth2Config(clientID, clientSecret, tokenURL string, audience string, scopes ...string)
}
