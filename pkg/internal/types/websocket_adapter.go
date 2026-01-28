package types

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

const (
	WebSocketFormatJSON   = "json"
	WebSocketFormatBinary = "binary"
	WebSocketFormatProto  = "proto"
	WebSocketFormatText   = "text"
)

// WebSocketServer accepts inbound WebSocket connections and streams messages into a pipeline.
// It can also broadcast outbound messages to all connected clients.
type WebSocketServer[T any] interface {
	// Serve starts the WebSocket server and dispatches inbound messages to submit.
	Serve(ctx context.Context, submit func(context.Context, T) error) error

	// Broadcast sends a message to all active connections.
	Broadcast(ctx context.Context, msg T) error

	// ConnectOutput wires connect one or more wires whose output should be broadcast.
	ConnectOutput(...Wire[T])

	// Configuration
	SetAddress(address string)
	SetEndpoint(path string)
	AddHeader(key, value string)
	SetAllowedOrigins(origins ...string)
	SetTokenQueryParam(name string)
	SetMessageFormat(format string)
	SetReadLimit(limit int64)
	SetWriteTimeout(timeout time.Duration)
	SetIdleTimeout(timeout time.Duration)
	SetSendBuffer(size int)
	SetMaxConnections(max int)
	SetTLSConfig(tlsCfg TLSConfig)

	// Auth
	SetAuthenticationOptions(opts *relay.AuthenticationOptions)
	SetStaticHeaders(headers map[string]string)
	SetDynamicAuthValidator(fn func(ctx context.Context, headers map[string]string) error)
	SetAuthRequired(required bool)

	// Observability
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name string, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
}

// WebSocketClientAdapter connects to a WebSocket endpoint and streams messages in/out.
type WebSocketClientAdapter[T any] interface {
	// Serve reads from the connection and dispatches decoded messages to submit.
	Serve(ctx context.Context, submit func(context.Context, T) error) error

	// ServeWriter encodes and writes messages from in to the connection.
	ServeWriter(ctx context.Context, in <-chan T) error

	// ServeDuplex runs both read and write loops on a single connection.
	ServeDuplex(ctx context.Context, in <-chan T, submit func(context.Context, T) error) error

	// Configuration
	SetURL(url string)
	SetHeaders(headers map[string]string)
	AddHeader(key, value string)
	SetMessageFormat(format string)
	SetReadLimit(limit int64)
	SetWriteTimeout(timeout time.Duration)
	SetIdleTimeout(timeout time.Duration)
	SetTLSConfig(tlsCfg TLSConfig)

	// Observability
	ConnectLogger(...Logger)
	ConnectSensor(...Sensor[T])
	GetComponentMetadata() ComponentMetadata
	SetComponentMetadata(name string, id string)
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})
}
