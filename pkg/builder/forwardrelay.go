package builder

import (
	"context"

	fr "github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// CompressionAlgorithm defines the types of compression available.
type CompressionAlgorithm = relay.CompressionAlgorithm

// EncryptionSuite defines the types of encryption available
// (mirroring your proto's relay.EncryptionSuite values).
type EncryptionSuite = relay.EncryptionSuite

// Constants for compression algorithms.
const (
	COMPRESS_NONE    CompressionAlgorithm = fr.COMPRESS_NONE
	COMPRESS_DEFLATE CompressionAlgorithm = fr.COMPRESS_DEFLATE
	COMPRESS_SNAPPY  CompressionAlgorithm = fr.COMPRESS_SNAPPY
	COMPRESS_ZSTD    CompressionAlgorithm = fr.COMPRESS_ZSTD
	COMPRESS_BROTLI  CompressionAlgorithm = fr.COMPRESS_BROTLI
	COMPRESS_LZ4     CompressionAlgorithm = fr.COMPRESS_LZ4

	ENCRYPTION_NONE    EncryptionSuite = 0
	ENCRYPTION_AES_GCM EncryptionSuite = 1
)

// ForwardRelayWithTarget sets the outbound target(s) for the ForwardRelay.
func ForwardRelayWithTarget[T any](t ...string) types.Option[types.ForwardRelay[T]] {
	return fr.WithTarget[T](t...)
}

// ForwardRelayWithComponentMetadata adds component metadata overrides.
func ForwardRelayWithComponentMetadata[T any](name string, id string) types.Option[types.ForwardRelay[T]] {
	return fr.WithComponentMetadata[T](name, id)
}

// ForwardRelayWithInput sets the input conduit(s) for the ForwardRelay.
func ForwardRelayWithInput[T any](input ...types.Receiver[T]) types.Option[types.ForwardRelay[T]] {
	return fr.WithInput[T](input...)
}

// ForwardRelayWithLogger adds one or more loggers to the ForwardRelay.
func ForwardRelayWithLogger[T any](logger ...types.Logger) types.Option[types.ForwardRelay[T]] {
	return fr.WithLogger[T](logger...)
}

// ForwardRelayWithPerformanceOptions sets performance-related options for the ForwardRelay.
func ForwardRelayWithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[types.ForwardRelay[T]] {
	return fr.WithPerformanceOptions[T](perfOptions)
}

// ForwardRelayWithSecurityOptions sets content-encryption options for the ForwardRelay.
func ForwardRelayWithSecurityOptions[T any](secOpts *relay.SecurityOptions, encryptionKey string) types.Option[types.ForwardRelay[T]] {
	return fr.WithSecurityOptions[T](secOpts, encryptionKey)
}

// ForwardRelayWithTLSConfig sets the TLS configuration for the ForwardRelay.
func ForwardRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ForwardRelay[T]] {
	return fr.WithTLSConfig[T](config)
}

// -------------------- New auth-related wrappers --------------------

// ForwardRelayWithAuthenticationOptions provides auth hints (mirrors proto MessageMetadata.authentication).
// Optional; receivers may ignore; server config may override.
func ForwardRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.ForwardRelay[T]] {
	return fr.WithAuthenticationOptions[T](opts)
}

// ForwardRelayWithOAuthBearer enables per-RPC Bearer injection using the provided token source.
// Pass nil to disable. Implementations must enforce TLS when enabled.
func ForwardRelayWithOAuthBearer[T any](ts types.OAuth2TokenSource) types.Option[types.ForwardRelay[T]] {
	return fr.WithOAuth2[T](ts)
}

// ForwardRelayWithStaticHeaders sets constant metadata headers added to every RPC.
func ForwardRelayWithStaticHeaders[T any](headers map[string]string) types.Option[types.ForwardRelay[T]] {
	return fr.WithStaticHeaders[T](headers)
}

// ForwardRelayWithDynamicHeaders sets a callback to compute per-request headers from context.
func ForwardRelayWithDynamicHeaders[T any](fn func(ctx context.Context) map[string]string) types.Option[types.ForwardRelay[T]] {
	return fr.WithDynamicHeaders[T](fn)
}

// -------------------- Constructors --------------------

// NewForwardRelay creates a new ForwardRelay with specified options.
func NewForwardRelay[T any](ctx context.Context, options ...types.Option[types.ForwardRelay[T]]) types.ForwardRelay[T] {
	return fr.NewForwardRelay[T](ctx, options...)
}

// NewPerformanceOptions creates a new PerformanceOptions.
func NewPerformanceOptions(useCompression bool, compressionAlgorithm relay.CompressionAlgorithm) *relay.PerformanceOptions {
	return &relay.PerformanceOptions{
		UseCompression:       useCompression,
		CompressionAlgorithm: compressionAlgorithm,
	}
}

// NewSecurityOptions creates a new SecurityOptions object indicating whether encryption is enabled and which suite to use.
func NewSecurityOptions(enabled bool, suite EncryptionSuite) *relay.SecurityOptions {
	return &relay.SecurityOptions{
		Enabled: enabled,
		Suite:   suite,
	}
}

// NewTlsClientConfig creates a new TLSConfig with configurable TLS versioning.
func NewTlsClientConfig(useTls bool, certFile string, keyFile string, caFile string, minTLSVersion uint16, maxTLSVersion uint16) *types.TLSConfig {
	return &types.TLSConfig{
		UseTLS:        useTls,
		CertFile:      certFile,
		KeyFile:       keyFile,
		CAFile:        caFile,
		MinTLSVersion: minTLSVersion,
		MaxTLSVersion: maxTLSVersion,
	}
}

// -------------------- OAuth2 translators (thin wrappers over forwardrelay) --------------------

// NewForwardRelayOAuth2JWTOptions builds JWT validation settings for resource servers.
func NewForwardRelayOAuth2JWTOptions(
	issuer string,
	jwksURI string,
	audiences []string,
	scopes []string,
	jwksCacheSeconds int32,
) *relay.OAuth2Options {
	return fr.NewOAuth2JWTOptions(issuer, jwksURI, audiences, scopes, jwksCacheSeconds)
}

// NewForwardRelayOAuth2IntrospectionOptions builds RFC 7662 introspection settings.
func NewForwardRelayOAuth2IntrospectionOptions(
	introspectionURL string,
	authType string, // "basic" | "bearer" | "none"
	clientID string,
	clientSecret string,
	bearerToken string,
	introspectionCacheSeconds int32,
) *relay.OAuth2Options {
	return fr.NewOAuth2IntrospectionOptions(introspectionURL, authType, clientID, clientSecret, bearerToken, introspectionCacheSeconds)
}

// NewForwardRelayOAuth2Forwarding toggles forwarding of the inbound bearer token to downstream services.
func NewForwardRelayOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return fr.NewOAuth2Forwarding(forward, forwardMetadataKey)
}

// NewForwardRelayMergeOAuth2Options merges non-zero fields from src into dst (shallow merge).
func NewForwardRelayMergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	return fr.MergeOAuth2Options(dst, src)
}

// NewForwardRelayAuthenticationOptionsOAuth2 builds AuthenticationOptions for OAuth2.
func NewForwardRelayAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return fr.NewAuthenticationOptionsOAuth2(oauth)
}

// NewForwardRelayAuthenticationOptionsMTLS builds AuthenticationOptions for mTLS-only expectations.
func NewForwardRelayAuthenticationOptionsMTLS(allowedPrincipals []string, trustDomain string) *relay.AuthenticationOptions {
	return fr.NewAuthenticationOptionsMTLS(allowedPrincipals, trustDomain)
}

// NewForwardRelayAuthenticationOptionsNone builds a disabled auth options object.
func NewForwardRelayAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return fr.NewAuthenticationOptionsNone()
}

// NewForwardRelayStaticBearerTokenSource returns a token source that always returns the provided token.
func NewForwardRelayStaticBearerTokenSource(token string) types.OAuth2TokenSource {
	return fr.NewStaticBearerTokenSource(token)
}

// NewForwardRelayEnvBearerTokenSource returns a token source that pulls the token from an env var at call time.
func NewForwardRelayEnvBearerTokenSource(envVar string) types.OAuth2TokenSource {
	return fr.NewEnvBearerTokenSource(envVar)
}
