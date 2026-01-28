package types

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

// OAuth2TokenSource unchanged...
type OAuth2TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

type ForwardRelay[T any] interface {
	// --- Existing API (unchanged) ---

	ConnectInput(...Receiver[T])
	ConnectLogger(...Logger)

	GetTargets() []string
	GetComponentMetadata() ComponentMetadata
	GetInput() []Receiver[T]
	IsRunning() bool
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	SetTargets(...string)
	SetComponentMetadata(name string, id string)
	SetPerformanceOptions(*relay.PerformanceOptions)
	SetSecurityOptions(secOptions *relay.SecurityOptions, encryptionKey string)
	SetTLSConfig(*TLSConfig)
	// SetPassthrough enables forwarding pre-wrapped payloads without modification.
	SetPassthrough(enabled bool)

	Start(context.Context) error
	Submit(ctx context.Context, item T) error
	Stop()

	// --- Optional auth/metadata hooks (for OAuth2 over gRPC) ---

	SetAuthenticationOptions(*relay.AuthenticationOptions)
	SetOAuth2(ts OAuth2TokenSource)
	SetStaticHeaders(map[string]string)
	SetDynamicHeaders(func(ctx context.Context) map[string]string)

	// --- NEW: auth gate control ---

	// SetAuthRequired toggles whether a valid Bearer token is REQUIRED
	// before any RPC is attempted when OAuth2 is configured.
	// If true (default), Submit() will fail fast if a token cannot be obtained.
	// If false, Submit() proceeds without Authorization (useful for dev).
	SetAuthRequired(required bool)

	// GetAuthRequired reports the current setting.
	GetAuthRequired() bool
}
