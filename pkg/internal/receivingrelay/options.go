// Package receivingrelay offers a collection of functional options that are used to configure instances
// of a ReceivingRelay. These options allow for flexible and dynamic adjustments to the ReceivingRelay's
// behavior and settings during initialization and runtime. The use of options follows the functional
// options pattern, providing a robust way to create configurable objects without an explosion of parasensors
// in constructors or needing to expose the underlying properties directly.

// Each function in this package returns a closure around the option's effect, which can be applied to
// a ReceivingRelay instance. This approach encapsulates the option's logic inside a function that modifies
// the state of a ReceivingRelay, ensuring that all configurations are applied safely and consistently.

// This file includes options for:
// - Setting the network address of the ReceivingRelay.
// - Configuring output conduits to which the ReceivingRelay sends its processed data.
// - Applying TLS configurations to secure the data communication channels.
// - Adjusting the internal buffer size for incoming data streams.
// - Attaching logging mechanisms to provide insight into the ReceivingRelay's operations and health.
// - Setting custom metadata for the ReceivingRelay, aiding in identification and management.

// Utilizing these options enhances the modularity and maintainability of the ReceivingRelay setup,
// enabling developers to customize and extend functionality as required by their specific use cases.
package receivingrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

// WithAddress returns an option for setting the network address of a ReceivingRelay.
func WithAddress[T any](address string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetAddress(address) }
}

// WithOutput configures output conduits.
func WithOutput[T any](output ...types.Submitter[T]) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.ConnectOutput(output...) }
}

// WithTLSConfig sets TLS configuration.
func WithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetTLSConfig(config) }
}

// WithBufferSize adjusts the internal data channel buffer size.
func WithBufferSize[T any](bufferSize uint32) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetDataChannel(bufferSize) }
}

// WithLogger attaches one or more loggers.
func WithLogger[T any](logger ...types.Logger) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.ConnectLogger(logger...) }
}

// WithComponentMetadata sets name and ID metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetComponentMetadata(name, id) }
}

// WithDecryptionKey sets the decryption key (e.g., AES-GCM).
func WithDecryptionKey[T any](key string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetDecryptionKey(key) }
}

// WithAuthenticationOptions configures expected auth mode/parameters.
func WithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetAuthenticationOptions(opts) }
}

// WithStaticHeaders enforces constant metadata on incoming requests.
func WithStaticHeaders[T any](headers map[string]string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetStaticHeaders(headers) }
}

// WithDynamicAuthValidator registers a per-request validation callback.
func WithDynamicAuthValidator[T any](fn func(ctx context.Context, md map[string]string) error) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetDynamicAuthValidator(fn) }
}

// ============================================================================
// OAuth2 / Auth constructors (server-side) — keep builder thin; expose here.
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

// NewOAuth2Forwarding controls forwarding of inbound bearer tokens to downstream services.
func NewOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		ForwardBearerToken: forward,
		ForwardMetadataKey: forwardMetadataKey,
	}
}

// MergeOAuth2Options merges non-zero fields from src into dst (uses proto clone to avoid mutex copy).
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

// NewAuthenticationOptionsOAuth2 composes AuthenticationOptions for OAuth2.
func NewAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_OAUTH2,
		Oauth2:  oauth,
	}
}

// NewAuthenticationOptionsMTLS composes AuthenticationOptions for mTLS-only.
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

// NewAuthenticationOptionsNone composes a disabled auth options object.
func NewAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return &relay.AuthenticationOptions{
		Enabled: false,
		Mode:    relay.AuthMode_AUTH_NONE,
	}
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

// WithAuthInterceptors installs unary and stream interceptors on the server.
// Pass nil for either to skip that interceptor type.
func WithAuthInterceptors[T any](
	unary grpc.UnaryServerInterceptor,
	stream grpc.StreamServerInterceptor,
) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetAuthInterceptors(unary, stream)
	}
}

// WithAuthRequired toggles strict auth enforcement on the receiver.
// When true, missing/invalid auth will reject requests; when false, it may log and allow.
func WithAuthRequired[T any](required bool) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetAuthRequired(required)
	}
}
