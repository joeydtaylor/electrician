package builder

import (
	"context"
	"net/http"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/websocketrelay"
)

// ---- Forward relay options ----

func WebSocketForwardRelayWithTarget[T any](t ...string) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetTargets(t...) }
}

func WebSocketForwardRelayWithComponentMetadata[T any](name string, id string) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetComponentMetadata(name, id) }
}

func WebSocketForwardRelayWithInput[T any](input ...types.Receiver[T]) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.ConnectInput(input...) }
}

func WebSocketForwardRelayWithLogger[T any](logger ...types.Logger) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.ConnectLogger(logger...) }
}

func WebSocketForwardRelayWithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetPerformanceOptions(perfOptions) }
}

func WebSocketForwardRelayWithSecurityOptions[T any](secOpts *relay.SecurityOptions, encryptionKey string) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetSecurityOptions(secOpts, encryptionKey) }
}

func WebSocketForwardRelayWithPassthrough[T any](enabled bool) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetPassthrough(enabled) }
}

func WebSocketForwardRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetTLSConfig(config) }
}

func WebSocketForwardRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetAuthenticationOptions(opts) }
}

func WebSocketForwardRelayWithOAuthBearer[T any](ts types.OAuth2TokenSource) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetOAuth2(ts) }
}

func WebSocketForwardRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetStaticHeaders(headers) }
}

func WebSocketForwardRelayWithDynamicHeaders[T any](fn func(ctx context.Context) map[string]string) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetDynamicHeaders(fn) }
}

func WebSocketForwardRelayWithAuthRequired[T any](required bool) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetAuthRequired(required) }
}

func WebSocketForwardRelayWithAckMode[T any](mode relay.AckMode) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetAckMode(mode) }
}

func WebSocketForwardRelayWithAckEveryN[T any](n uint32) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetAckEveryN(n) }
}

func WebSocketForwardRelayWithMaxInFlight[T any](n uint32) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetMaxInFlight(n) }
}

func WebSocketForwardRelayWithOmitPayloadMetadata[T any](omit bool) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetOmitPayloadMetadata(omit) }
}

func WebSocketForwardRelayWithStreamSendBuffer[T any](n int) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetStreamSendBuffer(n) }
}

func WebSocketForwardRelayWithMaxMessageBytes[T any](n int) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetMaxMessageBytes(n) }
}

func WebSocketForwardRelayWithDropOnFull[T any](drop bool) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetDropOnFull(drop) }
}

func WebSocketForwardRelayWithPayloadFormat[T any](format string) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetPayloadFormat(format) }
}

func WebSocketForwardRelayWithPayloadType[T any](payloadType string) types.Option[*websocketrelay.ForwardRelay[T]] {
	return func(fr *websocketrelay.ForwardRelay[T]) { fr.SetPayloadType(payloadType) }
}

// NewWebSocketForwardRelay creates a WebSocket forward relay.
func NewWebSocketForwardRelay[T any](ctx context.Context, options ...types.Option[*websocketrelay.ForwardRelay[T]]) *websocketrelay.ForwardRelay[T] {
	return websocketrelay.NewForwardRelay[T](ctx, options...)
}

// ---- Receiving relay options ----

func WebSocketReceivingRelayWithAddress[T any](address string) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetAddress(address) }
}

func WebSocketReceivingRelayWithPath[T any](path string) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetPath(path) }
}

func WebSocketReceivingRelayWithOutput[T any](output ...types.Submitter[T]) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.ConnectOutput(output...) }
}

func WebSocketReceivingRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetTLSConfig(config) }
}

func WebSocketReceivingRelayWithPassthrough[T any](enabled bool) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetPassthrough(enabled) }
}

func WebSocketReceivingRelayWithBufferSize[T any](bufferSize uint32) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetBufferSize(bufferSize) }
}

func WebSocketReceivingRelayWithLogger[T any](logger ...types.Logger) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.ConnectLogger(logger...) }
}

func WebSocketReceivingRelayWithComponentMetadata[T any](name string, id string) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetComponentMetadata(name, id) }
}

func WebSocketReceivingRelayWithDecryptionKey[T any](key string) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetDecryptionKey(key) }
}

func WebSocketReceivingRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetAuthenticationOptions(opts) }
}

func WebSocketReceivingRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetStaticHeaders(headers) }
}

func WebSocketReceivingRelayWithDynamicAuthValidator[T any](fn func(ctx context.Context, md map[string]string) error) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetDynamicAuthValidator(fn) }
}

func WebSocketReceivingRelayWithAuthRequired[T any](required bool) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetAuthRequired(required) }
}

func WebSocketReceivingRelayWithMaxMessageBytes[T any](n int) types.Option[*websocketrelay.ReceivingRelay[T]] {
	return func(rr *websocketrelay.ReceivingRelay[T]) { rr.SetMaxMessageBytes(n) }
}

// NewWebSocketReceivingRelay creates a WebSocket receiving relay.
func NewWebSocketReceivingRelay[T any](ctx context.Context, options ...types.Option[*websocketrelay.ReceivingRelay[T]]) *websocketrelay.ReceivingRelay[T] {
	return websocketrelay.NewReceivingRelay[T](ctx, options...)
}

// Auth helpers (same shape as gRPC/QUIC builders)
func NewWebSocketReceivingRelayOAuth2JWTOptions(issuer, jwksURL string, audience, scopes []string, cacheSeconds int32) *relay.OAuth2Options {
	o := &relay.OAuth2Options{
		AcceptJwt:        true,
		Issuer:           issuer,
		JwksUri:          jwksURL,
		RequiredAudience: audience,
		RequiredScopes:   scopes,
		JwksCacheSeconds: cacheSeconds,
	}
	return o
}

func NewWebSocketReceivingRelayOAuth2IntrospectionOptions(url, authType, clientID, clientSecret, bearer string, scopes []string, cacheSeconds int32) *relay.OAuth2Options {
	return &relay.OAuth2Options{
		AcceptIntrospection:      true,
		IntrospectionUrl:         url,
		IntrospectionAuthType:    authType,
		IntrospectionClientId:    clientID,
		IntrospectionClientSecret: clientSecret,
		IntrospectionBearerToken: bearer,
		RequiredScopes:           scopes,
		IntrospectionCacheSeconds: cacheSeconds,
	}
}

func NewWebSocketReceivingRelayMergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	return NewReceivingRelayMergeOAuth2Options(dst, src)
}

func NewWebSocketReceivingRelayAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return NewReceivingRelayAuthenticationOptionsOAuth2(oauth)
}

func NewWebSocketReceivingRelayAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return NewReceivingRelayAuthenticationOptionsNone()
}

// ---- Forward relay token helpers (reuse QUIC helpers) ----
func NewWebSocketForwardRelayStaticBearerTokenSource(token string) types.OAuth2TokenSource {
	return NewQuicForwardRelayStaticBearerTokenSource(token)
}

func NewWebSocketForwardRelayEnvBearerTokenSource(envVar string) types.OAuth2TokenSource {
	return NewQuicForwardRelayEnvBearerTokenSource(envVar)
}

func NewWebSocketForwardRelayRefreshingClientCredentialsSource(
	authBaseURL string,
	clientID string,
	clientSecret string,
	scopes []string,
	leeway time.Duration,
	hc *http.Client,
) types.OAuth2TokenSource {
	return NewQuicForwardRelayRefreshingClientCredentialsSource(authBaseURL, clientID, clientSecret, scopes, leeway, hc)
}
