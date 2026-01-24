package forwardrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithTarget configures outbound targets.
func WithTarget[T any](targets ...string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetTargets(targets...) }
}

// WithLogger attaches loggers.
func WithLogger[T any](logger ...types.Logger) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.ConnectLogger(logger...) }
}

// WithInput attaches input receivers.
func WithInput[T any](input ...types.Receiver[T]) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.ConnectInput(input...) }
}

// WithTLSConfig sets TLS configuration.
func WithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetTLSConfig(config) }
}

// WithPerformanceOptions sets performance options.
func WithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetPerformanceOptions(perfOptions) }
}

// WithSecurityOptions sets content-encryption options and key.
func WithSecurityOptions[T any](secOptions *relay.SecurityOptions, encryptionKey string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetSecurityOptions(secOptions, encryptionKey) }
}

// WithComponentMetadata sets name and ID metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetComponentMetadata(name, id) }
}

// WithAuthenticationOptions provides auth hints (mirrors proto MessageMetadata.authentication).
func WithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetAuthenticationOptions(opts) }
}

// WithOAuth2 enables per-RPC Bearer injection using the provided token source.
func WithOAuth2[T any](ts types.OAuth2TokenSource) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetOAuth2(ts) }
}

// WithStaticHeaders sets constant metadata headers added to every RPC.
func WithStaticHeaders[T any](headers map[string]string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetStaticHeaders(headers) }
}

// WithDynamicHeaders sets a callback to compute per-request headers from context.
func WithDynamicHeaders[T any](fn func(ctx context.Context) map[string]string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetDynamicHeaders(fn) }
}

// WithAuthRequired toggles whether a valid bearer token is required when OAuth2 is configured.
func WithAuthRequired[T any](required bool) types.Option[types.ForwardRelay[T]] {
	return func(c types.ForwardRelay[T]) {
		if fr, ok := c.(*ForwardRelay[T]); ok {
			fr.authRequired = required
		}
	}
}

// ForwardRelayWithAuthRequired keeps compatibility with earlier naming.
func ForwardRelayWithAuthRequired[T any](required bool) types.Option[types.ForwardRelay[T]] {
	return WithAuthRequired[T](required)
}
