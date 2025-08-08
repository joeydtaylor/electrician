package types

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

// OAuth2TokenSource supplies access tokens for per-RPC Bearer auth.
// Implementations may cache/refresh (e.g., client_credentials against aegis-auth).
type OAuth2TokenSource interface {
	AccessToken(ctx context.Context) (string, error)
}

// ForwardRelay defines the operations for a Forward Relay (client side).
// Additions are strictly additive and optional; existing callers remain valid.
type ForwardRelay[T any] interface {
	// --- Existing API (unchanged) ---

	ConnectInput(...Receiver[T])
	ConnectLogger(...Logger)

	GetTargets() []string
	GetComponentMetadata() ComponentMetadata
	GetInput() []Receiver[T]
	IsRunning() bool
	NotifyLoggers(level LogLevel, format string, args ...interface{})

	SetTargets(...string)
	SetComponentMetadata(name string, id string)
	SetPerformanceOptions(*relay.PerformanceOptions)
	SetSecurityOptions(secOptions *relay.SecurityOptions, encryptionKey string)
	SetTLSConfig(*TLSConfig)

	Start(context.Context) error
	Submit(ctx context.Context, item T) error
	Stop()

	// --- Optional auth/metadata hooks (for OAuth2 over gRPC) ---

	// SetAuthenticationOptions provides auth hints to the receiving resource server.
	// This mirrors proto MessageMetadata.authentication and is safe to omit.
	SetAuthenticationOptions(*relay.AuthenticationOptions)

	// SetOAuth2 enables per-RPC Bearer injection using the provided token source.
	// When enabled, the implementation MUST:
	//   - Require TLS on the underlying transport.
	//   - Add `authorization: Bearer <token>` to outgoing gRPC metadata.
	// Disable by passing nil.
	SetOAuth2(ts OAuth2TokenSource)

	// SetStaticHeaders adds/overrides constant metadata headers on every RPC
	// (e.g., correlation tags, tenant hints). Values here are merged after
	// per-request headers and before the Authorization header.
	SetStaticHeaders(map[string]string)

	// SetDynamicHeaders installs a callback to compute per-request headers.
	// If provided, it runs for each Submit() and can inspect ctx to produce
	// keys like "x-trace-id" or additional auth forwarding keys when needed.
	// Return nil to add nothing for that call.
	SetDynamicHeaders(func(ctx context.Context) map[string]string)
}
