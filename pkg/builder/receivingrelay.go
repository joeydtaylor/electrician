package builder

import (
	"context"

	rr "github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
)

// NewReceivingRelay creates a new ReceivingRelay with specified options.
func NewReceivingRelay[T any](ctx context.Context, options ...types.Option[types.ReceivingRelay[T]]) types.ReceivingRelay[T] {
	return rr.NewReceivingRelay[T](ctx, options...)
}

// NewTlsServerConfig creates a new TLS configuration with configurable TLS versioning.
func NewTlsServerConfig(useTls bool, certFile string, keyFile string, caFile string, subjectAlternativeName string, minTLSVersion uint16, maxTLSVersion uint16) *types.TLSConfig {
	return &types.TLSConfig{
		UseTLS:                 useTls,
		CertFile:               certFile,
		KeyFile:                keyFile,
		CAFile:                 caFile,
		SubjectAlternativeName: subjectAlternativeName,
		MinTLSVersion:          minTLSVersion,
		MaxTLSVersion:          maxTLSVersion,
	}
}

// ReceivingRelayWithAddress sets the network address for the ReceivingRelay.
func ReceivingRelayWithAddress[T any](address string) types.Option[types.ReceivingRelay[T]] {
	return rr.WithAddress[T](address)
}

// ReceivingRelayWithBufferSize sets the buffer size of the data channel for the ReceivingRelay.
func ReceivingRelayWithBufferSize[T any](bufferSize uint32) types.Option[types.ReceivingRelay[T]] {
	return rr.WithBufferSize[T](bufferSize)
}

// ReceivingRelayWithComponentMetadata adds component metadata overrides.
func ReceivingRelayWithComponentMetadata[T any](name string, id string) types.Option[types.ReceivingRelay[T]] {
	return rr.WithComponentMetadata[T](name, id)
}

// ReceivingRelayWithLogger attaches a logger to the ReceivingRelay.
func ReceivingRelayWithLogger[T any](logger ...types.Logger) types.Option[types.ReceivingRelay[T]] {
	return rr.WithLogger[T](logger...)
}

// ReceivingRelayWithOutput connects output conduits to the ReceivingRelay.
func ReceivingRelayWithOutput[T any](output ...types.Submitter[T]) types.Option[types.ReceivingRelay[T]] {
	return rr.WithOutput[T](output...)
}

// ReceivingRelayWithTLSConfig sets the TLS configuration for the ReceivingRelay.
func ReceivingRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ReceivingRelay[T]] {
	return rr.WithTLSConfig[T](config)
}

// ReceivingRelayWithPassthrough forwards raw WrappedPayload values without unwrap/decrypt.
func ReceivingRelayWithPassthrough[T any](enabled bool) types.Option[types.ReceivingRelay[T]] {
	return rr.WithPassthrough[T](enabled)
}

// ReceivingRelayWithGRPCWebConfig configures gRPC-Web CORS and transport behavior.
func ReceivingRelayWithGRPCWebConfig[T any](config *types.GRPCWebConfig) types.Option[types.ReceivingRelay[T]] {
	return rr.WithGRPCWebConfig[T](config)
}

// ReceivingRelayWithDecryptionKey sets the decryption key for the ReceivingRelay.
func ReceivingRelayWithDecryptionKey[T any](key string) types.Option[types.ReceivingRelay[T]] {
	return rr.WithDecryptionKey[T](key)
}

// -------------------- New auth-related wrappers --------------------

// ReceivingRelayWithAuthenticationOptions provides expected auth mode/parameters (proto hints).
func ReceivingRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.ReceivingRelay[T]] {
	return rr.WithAuthenticationOptions[T](opts)
}

// ReceivingRelayWithStaticHeaders enforces constant metadata keys/values.
func ReceivingRelayWithStaticHeaders[T any](headers map[string]string) types.Option[types.ReceivingRelay[T]] {
	return rr.WithStaticHeaders[T](headers)
}

// ReceivingRelayWithDynamicAuthValidator registers a per-request validation callback.
func ReceivingRelayWithDynamicAuthValidator[T any](fn func(ctx context.Context, md map[string]string) error) types.Option[types.ReceivingRelay[T]] {
	return rr.WithDynamicAuthValidator[T](fn)
}

// ReceivingRelayWithAuthInterceptors installs unary and stream interceptors
// used to enforce/observe auth on inbound RPCs. Pass nil for either to skip it.
func ReceivingRelayWithAuthInterceptors[T any](
	unary grpc.UnaryServerInterceptor,
	stream grpc.StreamServerInterceptor,
) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetAuthInterceptors(unary, stream)
	}
}

// ReceivingRelayWithAuthRequired toggles strict auth enforcement. When false,
// failed/absent credentials are logged and allowed (best-effort).
func ReceivingRelayWithAuthRequired[T any](required bool) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetAuthRequired(required)
	}
}

// -------------------- OAuth2 translators (prefixed to avoid collision) --------------------

// NewReceivingRelayOAuth2JWTOptions builds JWT validation settings for resource servers.
func NewReceivingRelayOAuth2JWTOptions(
	issuer string,
	jwksURI string,
	audiences []string,
	scopes []string,
	jwksCacheSeconds int32,
) *relay.OAuth2Options {
	return rr.NewOAuth2JWTOptions(issuer, jwksURI, audiences, scopes, jwksCacheSeconds)
}

// NewReceivingRelayOAuth2IntrospectionOptions builds RFC 7662 introspection settings.
func NewReceivingRelayOAuth2IntrospectionOptions(
	introspectionURL string,
	authType string, // "basic" | "bearer" | "none"
	clientID string,
	clientSecret string,
	bearerToken string,
	introspectionCacheSeconds int32,
) *relay.OAuth2Options {
	return rr.NewOAuth2IntrospectionOptions(introspectionURL, authType, clientID, clientSecret, bearerToken, introspectionCacheSeconds)
}

// NewReceivingRelayOAuth2Forwarding toggles forwarding of the inbound bearer token to downstream services.
func NewReceivingRelayOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return rr.NewOAuth2Forwarding(forward, forwardMetadataKey)
}

// NewReceivingRelayMergeOAuth2Options merges non-zero fields from src into dst.
func NewReceivingRelayMergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	return rr.MergeOAuth2Options(dst, src)
}

// NewReceivingRelayAuthenticationOptionsOAuth2 composes AuthenticationOptions for OAuth2.
func NewReceivingRelayAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return rr.NewAuthenticationOptionsOAuth2(oauth)
}

// NewReceivingRelayAuthenticationOptionsMTLS composes AuthenticationOptions for mTLS-only.
func NewReceivingRelayAuthenticationOptionsMTLS(allowedPrincipals []string, trustDomain string) *relay.AuthenticationOptions {
	return rr.NewAuthenticationOptionsMTLS(allowedPrincipals, trustDomain)
}

// NewReceivingRelayAuthenticationOptionsNone composes a disabled auth options object.
func NewReceivingRelayAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return rr.NewAuthenticationOptionsNone()
}

// NewGRPCWebConfigAllowAllOrigins returns a gRPC-Web config that accepts any origin.
// This only affects browser CORS behavior; auth and TLS are still enforced.
func NewGRPCWebConfigAllowAllOrigins() *types.GRPCWebConfig {
	return &types.GRPCWebConfig{AllowAllOrigins: true}
}

// NewGRPCWebConfigAllowAll returns a gRPC-Web config that accepts any origin.
// Deprecated: use NewGRPCWebConfigAllowAllOrigins for clarity.
func NewGRPCWebConfigAllowAll() *types.GRPCWebConfig {
	return NewGRPCWebConfigAllowAllOrigins()
}
