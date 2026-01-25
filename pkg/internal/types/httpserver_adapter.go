// httpserver_adapter.go
package types

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

// HTTPServerRequest wraps the incoming data for a server-side handler.
type HTTPServerRequest[T any] struct {
	Body    T                   // Decoded request body
	Raw     []byte              // For binary/unstructured data
	Headers map[string][]string // Incoming HTTP headers
	Method  string              // e.g., "POST"
	Path    string              // e.g., "/webhook"
}

type HTTPServerResponse struct {
	StatusCode int
	Headers    map[string]string
	Body       []byte // Raw bytes to be sent directly to the client
}

// HTTPServerError encapsulates server-side errors with an HTTP status code.
type HTTPServerError struct {
	StatusCode int
	Err        error
	Message    string
}

func (e *HTTPServerError) Error() string {
	return fmt.Sprintf("HTTP %d: %s, %v", e.StatusCode, e.Message, e.Err)
}

// HTTPServer is the inverse of HTTPClientAdapter. It listens on a
// specified address/endpoint for incoming requests, decodes them into T,
// and passes them into a pipeline via a submitFunc or similar mechanism.
//
// For example, it could act as a webhook receiver: “When data arrives at
// /my-webhook, parse JSON into T, then call submitFunc(ctx, T) to feed the pipeline.”
type HTTPServer[T any] interface {

	// Serve starts the HTTP server. Once running, it should listen for incoming
	// requests on the configured method/endpoint. For each valid request:
	//
	//  1. Parse or decode the request body into T.
	//  2. Create an HTTPServerRequest[T] from it.
	//  3. Invoke submitFunc(ctx, requestStruct) to feed that data into the pipeline.
	//  4. Use the returned HTTPServerResponse to send custom responses to the client.
	//
	// If any error occurs (e.g., parsing fails), the server can return an HTTP error code
	// back to the client and optionally notify sensors/loggers.
	Serve(ctx context.Context, submitFunc func(ctx context.Context, req T) (HTTPServerResponse, error)) error

	// ConnectLogger attaches one or more Logger instances that will receive log messages
	// about server start, incoming request handling, errors, etc.
	ConnectLogger(...Logger)

	// ConnectSensor attaches one or more Sensor instances that will be notified
	// of significant lifecycle events (on request start, on error, on request complete, etc.).
	ConnectSensor(...Sensor[T])

	// SetAddress sets the IP/port or socket on which the server will listen
	// (e.g., ":8080" or "127.0.0.1:8443").
	SetAddress(address string)

	// SetServerConfig sets the route/method combination this server will handle.
	// In a simple webhook scenario, you might only need one method and one path.
	// E.g., SetServerConfig("POST", "/my-webhook").
	SetServerConfig(method, endpoint string)

	// AddHeader could be used to configure default response headers or
	// server-level headers (e.g., "Server: MyCustomServer/1.0").
	AddHeader(key, value string)

	// GetComponentMetadata retrieves metadata about the server (its ID, name, type).
	// This is useful for logging, monitoring, or identifying the adapter in a pipeline.
	GetComponentMetadata() ComponentMetadata

	// SetComponentMetadata sets server metadata for identification (e.g., name, ID).
	SetComponentMetadata(name string, id string)

	// NotifyLoggers is a helper function to send formatted log messages
	// to all connected loggers. Observers can filter by LogLevel.
	NotifyLoggers(level LogLevel, format string, args ...interface{})

	// SetTimeout sets a read/write timeout for inbound requests, to avoid
	// hanging connections or slowloris attacks.
	SetTimeout(timeout time.Duration)

	// SetTLSConfig configures the server to use TLS for inbound connections
	// according to the provided TLSConfig. If UseTLS == false, the server
	// should revert to plain HTTP (no TLS).
	SetTLSConfig(tlsCfg TLSConfig)

	// SetAuthenticationOptions configures OAuth2 authentication options.
	SetAuthenticationOptions(opts *relay.AuthenticationOptions)

	// SetStaticHeaders enforces constant header key/value pairs on incoming requests.
	SetStaticHeaders(headers map[string]string)

	// SetDynamicAuthValidator registers a per-request validation callback.
	SetDynamicAuthValidator(fn func(ctx context.Context, headers map[string]string) error)

	// SetAuthRequired toggles strict auth enforcement.
	SetAuthRequired(required bool)
}
