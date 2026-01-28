package builder

import (
	"context"
	"net/http"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/quicrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ---- Forward QUIC relay options ----

func QuicForwardRelayWithTarget[T any](t ...string) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithTarget[T](t...)
}

func QuicForwardRelayWithComponentMetadata[T any](name string, id string) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithComponentMetadata[T](name, id)
}

func QuicForwardRelayWithInput[T any](input ...types.Receiver[T]) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithInput[T](input...)
}

func QuicForwardRelayWithLogger[T any](logger ...types.Logger) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithLogger[T](logger...)
}

func QuicForwardRelayWithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithPerformanceOptions[T](perfOptions)
}

func QuicForwardRelayWithSecurityOptions[T any](secOpts *relay.SecurityOptions, encryptionKey string) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithSecurityOptions[T](secOpts, encryptionKey)
}

func QuicForwardRelayWithPassthrough[T any](enabled bool) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithPassthrough[T](enabled)
}

func QuicForwardRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithTLSConfig[T](config)
}

func QuicForwardRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithAuthenticationOptions[T](opts)
}

func QuicForwardRelayWithOAuthBearer[T any](ts types.OAuth2TokenSource) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithOAuth2[T](ts)
}

func QuicForwardRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithStaticHeaders[T](headers)
}

func QuicForwardRelayWithDynamicHeaders[T any](fn func(ctx context.Context) map[string]string) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithDynamicHeaders[T](fn)
}

func QuicForwardRelayWithAuthRequired[T any](required bool) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithAuthRequired[T](required)
}

func QuicForwardRelayWithAckMode[T any](mode relay.AckMode) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithAckMode[T](mode)
}

func QuicForwardRelayWithAckEveryN[T any](n uint32) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithAckEveryN[T](n)
}

func QuicForwardRelayWithMaxInFlight[T any](n uint32) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithMaxInFlight[T](n)
}

func QuicForwardRelayWithOmitPayloadMetadata[T any](omit bool) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithOmitPayloadMetadata[T](omit)
}

func QuicForwardRelayWithStreamSendBuffer[T any](n int) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithStreamSendBuffer[T](n)
}

func QuicForwardRelayWithMaxFrameBytes[T any](n int) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithMaxFrameBytes[T](n)
}

func QuicForwardRelayWithDropOnFull[T any](drop bool) types.Option[*quicrelay.ForwardRelay[T]] {
	return quicrelay.ForwardRelayWithDropOnFull[T](drop)
}

// NewQuicForwardRelay creates a QUIC forward relay.
func NewQuicForwardRelay[T any](ctx context.Context, options ...types.Option[*quicrelay.ForwardRelay[T]]) *quicrelay.ForwardRelay[T] {
	return quicrelay.NewForwardRelay[T](ctx, options...)
}

// ---- Receiving QUIC relay options ----

func QuicReceivingRelayWithAddress[T any](address string) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithAddress[T](address)
}

func QuicReceivingRelayWithOutput[T any](output ...types.Submitter[T]) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithOutput[T](output...)
}

func QuicReceivingRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithTLSConfig[T](config)
}

func QuicReceivingRelayWithPassthrough[T any](enabled bool) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithPassthrough[T](enabled)
}

func QuicReceivingRelayWithBufferSize[T any](bufferSize uint32) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithBufferSize[T](bufferSize)
}

func QuicReceivingRelayWithLogger[T any](logger ...types.Logger) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithLogger[T](logger...)
}

func QuicReceivingRelayWithComponentMetadata[T any](name string, id string) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithComponentMetadata[T](name, id)
}

func QuicReceivingRelayWithDecryptionKey[T any](key string) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithDecryptionKey[T](key)
}

func QuicReceivingRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithAuthenticationOptions[T](opts)
}

func QuicReceivingRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithStaticHeaders[T](headers)
}

func QuicReceivingRelayWithDynamicAuthValidator[T any](fn func(ctx context.Context, md map[string]string) error) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithDynamicAuthValidator[T](fn)
}

func QuicReceivingRelayWithAuthRequired[T any](required bool) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithAuthRequired[T](required)
}

func QuicReceivingRelayWithMaxFrameBytes[T any](n int) types.Option[*quicrelay.ReceivingRelay[T]] {
	return quicrelay.ReceivingRelayWithMaxFrameBytes[T](n)
}

// NewQuicReceivingRelay creates a QUIC receiving relay.
func NewQuicReceivingRelay[T any](ctx context.Context, options ...types.Option[*quicrelay.ReceivingRelay[T]]) *quicrelay.ReceivingRelay[T] {
	return quicrelay.NewReceivingRelay[T](ctx, options...)
}

// ---- OAuth2 option helpers (reuse receivingrelay helpers) ----

func NewQuicReceivingRelayOAuth2JWTOptions(
	issuer string,
	jwksURI string,
	audiences []string,
	scopes []string,
	jwksCacheSeconds int32,
) *relay.OAuth2Options {
	return receivingrelay.NewOAuth2JWTOptions(issuer, jwksURI, audiences, scopes, jwksCacheSeconds)
}

func NewQuicReceivingRelayOAuth2IntrospectionOptions(
	introspectionURL string,
	authType string,
	clientID string,
	clientSecret string,
	bearerToken string,
	introspectionCacheSeconds int32,
) *relay.OAuth2Options {
	return receivingrelay.NewOAuth2IntrospectionOptions(introspectionURL, authType, clientID, clientSecret, bearerToken, introspectionCacheSeconds)
}

func NewQuicReceivingRelayOAuth2Forwarding(forward bool, forwardMetadataKey string) *relay.OAuth2Options {
	return receivingrelay.NewOAuth2Forwarding(forward, forwardMetadataKey)
}

func NewQuicReceivingRelayMergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	return receivingrelay.MergeOAuth2Options(dst, src)
}

func NewQuicReceivingRelayAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return receivingrelay.NewAuthenticationOptionsOAuth2(oauth)
}

func NewQuicReceivingRelayAuthenticationOptionsMTLS(allowedPrincipals []string, trustDomain string) *relay.AuthenticationOptions {
	return receivingrelay.NewAuthenticationOptionsMTLS(allowedPrincipals, trustDomain)
}

func NewQuicReceivingRelayAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return receivingrelay.NewAuthenticationOptionsNone()
}

// Forward relay OAuth token sources (reuse forwardrelay helpers).

func NewQuicForwardRelayStaticBearerTokenSource(token string) types.OAuth2TokenSource {
	return NewForwardRelayStaticBearerTokenSource(token)
}

func NewQuicForwardRelayEnvBearerTokenSource(envVar string) types.OAuth2TokenSource {
	return NewForwardRelayEnvBearerTokenSource(envVar)
}

func NewQuicForwardRelayRefreshingClientCredentialsSource(
	baseURL string,
	clientID string,
	clientSecret string,
	scopes []string,
	leeway time.Duration,
	httpClient *http.Client,
) types.OAuth2TokenSource {
	return NewForwardRelayRefreshingClientCredentialsSource(baseURL, clientID, clientSecret, scopes, leeway, httpClient)
}
