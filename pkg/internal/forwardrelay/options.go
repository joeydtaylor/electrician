package forwardrelay

import (
	"context"
	"os"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/protobuf/proto"
)

// WithTarget configures one or more outbound target addresses.
func WithTarget[T any](targets ...string) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetTargets(targets...) }
}

// WithLogger attaches one or more loggers.
func WithLogger[T any](logger ...types.Logger) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.ConnectLogger(logger...) }
}

// WithInput attaches one or more inputs.
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

// -------------------- New auth-related options --------------------

// WithAuthenticationOptions provides auth hints (mirrors proto MessageMetadata.authentication).
// Optional; receivers may ignore; server config may override.
func WithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.ForwardRelay[T]] {
	return func(fr types.ForwardRelay[T]) { fr.SetAuthenticationOptions(opts) }
}

// WithOAuth2 enables per-RPC Bearer injection using the provided token source.
// Pass nil to disable. Implementations must enforce TLS when enabled.
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

// ============================================================================
// OAuth2 / Auth builders (kept close to forwardrelay; builder package stays thin)
// ============================================================================

// NewOAuth2JWTOptions builds JWT validation settings for resource servers.
func NewOAuth2JWTOptions(
	issuer string,
	jwksURI string,
	audiences []string,
	scopes []string,
	jwksCacheSeconds int32,
) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		AcceptJwt:             true,
		AcceptIntrospection:   false,
		Issuer:                issuer,
		JwksUri:               jwksURI,
		RequiredAudience:      cloneStrings(audiences),
		RequiredScopes:        cloneStrings(scopes),
		JwksCacheSeconds:      jwksCacheSeconds,
		ForwardBearerToken:    false,
		ForwardMetadataKey:    "",
		IntrospectionUrl:      "",
		IntrospectionAuthType: "",
	}
}

// NewOAuth2IntrospectionOptions builds RFC 7662 introspection settings.
func NewOAuth2IntrospectionOptions(
	introspectionURL string,
	authType string, // "basic" | "bearer" | "none"
	clientID string,
	clientSecret string,
	bearerToken string,
	introspectionCacheSeconds int32,
) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		AcceptJwt:                 false,
		AcceptIntrospection:       true,
		IntrospectionUrl:          introspectionURL,
		IntrospectionAuthType:     authType,
		IntrospectionClientId:     clientID,
		IntrospectionClientSecret: clientSecret,
		IntrospectionBearerToken:  bearerToken,
		IntrospectionCacheSeconds: introspectionCacheSeconds,
	}
}

// NewOAuth2Forwarding toggles forwarding of the inbound bearer token to downstream services.
func NewOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		ForwardBearerToken: forward,
		ForwardMetadataKey: forwardMetadataKey,
	}
}

// MergeOAuth2Options merges non-zero fields from src into dst (shallow merge).
func MergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	if dst == nil {
		return cloneOAuth2(src)
	}
	if src == nil {
		return dst
	}
	if src.AcceptJwt {
		dst.AcceptJwt = true
	}
	if src.AcceptIntrospection {
		dst.AcceptIntrospection = true
	}
	if src.Issuer != "" {
		dst.Issuer = src.Issuer
	}
	if src.JwksUri != "" {
		dst.JwksUri = src.JwksUri
	}
	if len(src.RequiredAudience) > 0 {
		dst.RequiredAudience = cloneStrings(src.RequiredAudience)
	}
	if len(src.RequiredScopes) > 0 {
		dst.RequiredScopes = cloneStrings(src.RequiredScopes)
	}
	if src.IntrospectionUrl != "" {
		dst.IntrospectionUrl = src.IntrospectionUrl
	}
	if src.IntrospectionAuthType != "" {
		dst.IntrospectionAuthType = src.IntrospectionAuthType
	}
	if src.IntrospectionClientId != "" {
		dst.IntrospectionClientId = src.IntrospectionClientId
	}
	if src.IntrospectionClientSecret != "" {
		dst.IntrospectionClientSecret = src.IntrospectionClientSecret
	}
	if src.IntrospectionBearerToken != "" {
		dst.IntrospectionBearerToken = src.IntrospectionBearerToken
	}
	if src.JwksCacheSeconds != 0 {
		dst.JwksCacheSeconds = src.JwksCacheSeconds
	}
	if src.IntrospectionCacheSeconds != 0 {
		dst.IntrospectionCacheSeconds = src.IntrospectionCacheSeconds
	}
	if src.ForwardMetadataKey != "" || src.ForwardBearerToken {
		dst.ForwardBearerToken = src.ForwardBearerToken
		dst.ForwardMetadataKey = src.ForwardMetadataKey
	}
	return dst
}

// NewAuthenticationOptionsOAuth2 builds a ready-to-use AuthenticationOptions for OAuth2.
func NewAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_OAUTH2,
		Oauth2:  oauth,
	}
}

// NewAuthenticationOptionsMTLS builds AuthenticationOptions for mTLS-only expectations.
func NewAuthenticationOptionsMTLS(allowedPrincipals []string, trustDomain string) *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_MUTUAL_TLS,
		Mtls: &relay.MTLSOptions{
			AllowedPrincipals: cloneStrings(allowedPrincipals),
			TrustDomain:       trustDomain,
		},
	}
}

// NewAuthenticationOptionsNone builds a disabled auth options object.
func NewAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: false,
		Mode:    relay.AuthMode_AUTH_NONE,
	}
}

// -------------------- Token source helpers (client-side) --------------------

type staticTokenSource struct{ tok string }

func (s staticTokenSource) AccessToken(_ context.Context) (string, error) { return s.tok, nil }

// NewStaticBearerTokenSource returns a token source that always returns the provided token.
func NewStaticBearerTokenSource(token string) types.OAuth2TokenSource {
	return staticTokenSource{tok: token}
}

type envTokenSource struct{ key string }

func (e envTokenSource) AccessToken(_ context.Context) (string, error) {
	return os.Getenv(e.key), nil
}

// NewEnvBearerTokenSource returns a token source that pulls the token from an env var at call time.
func NewEnvBearerTokenSource(envVar string) types.OAuth2TokenSource {
	return envTokenSource{key: envVar}
}

// -------------------- small internals --------------------

func cloneStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

// cloneOAuth2 avoids copying proto internal mutexes by using proto.Clone.
func cloneOAuth2(in *relay.OAuth2Options) *relay.OAuth2Options {
	if in == nil {
		return nil
	}
	return proto.Clone(in).(*relay.OAuth2Options)
}
