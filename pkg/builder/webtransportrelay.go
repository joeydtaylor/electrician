//go:build webtransport

package builder

import (
	"context"
	"net/http"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/webtransportrelay"
)

// ---- Forward relay options ----

func WebTransportForwardRelayWithTarget[T any](t ...string) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetTargets(t...) }
}

func WebTransportForwardRelayWithComponentMetadata[T any](name string, id string) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetComponentMetadata(name, id) }
}

func WebTransportForwardRelayWithInput[T any](input ...types.Receiver[T]) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.ConnectInput(input...) }
}

func WebTransportForwardRelayWithLogger[T any](logger ...types.Logger) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.ConnectLogger(logger...) }
}

func WebTransportForwardRelayWithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetPerformanceOptions(perfOptions) }
}

func WebTransportForwardRelayWithSecurityOptions[T any](secOpts *relay.SecurityOptions, encryptionKey string) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetSecurityOptions(secOpts, encryptionKey) }
}

func WebTransportForwardRelayWithPassthrough[T any](enabled bool) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetPassthrough(enabled) }
}

func WebTransportForwardRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetTLSConfig(config) }
}

func WebTransportForwardRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetAuthenticationOptions(opts) }
}

func WebTransportForwardRelayWithOAuthBearer[T any](ts types.OAuth2TokenSource) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetOAuth2(ts) }
}

func WebTransportForwardRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetStaticHeaders(headers) }
}

func WebTransportForwardRelayWithDynamicHeaders[T any](fn func(ctx context.Context) map[string]string) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetDynamicHeaders(fn) }
}

func WebTransportForwardRelayWithAuthRequired[T any](required bool) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetAuthRequired(required) }
}

func WebTransportForwardRelayWithAckMode[T any](mode relay.AckMode) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetAckMode(mode) }
}

func WebTransportForwardRelayWithAckEveryN[T any](n uint32) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetAckEveryN(n) }
}

func WebTransportForwardRelayWithMaxInFlight[T any](n uint32) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetMaxInFlight(n) }
}

func WebTransportForwardRelayWithOmitPayloadMetadata[T any](omit bool) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetOmitPayloadMetadata(omit) }
}

func WebTransportForwardRelayWithStreamSendBuffer[T any](n int) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetStreamSendBuffer(n) }
}

func WebTransportForwardRelayWithMaxFrameBytes[T any](n int) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetMaxFrameBytes(n) }
}

func WebTransportForwardRelayWithDropOnFull[T any](drop bool) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetDropOnFull(drop) }
}

func WebTransportForwardRelayWithPayloadFormat[T any](format string) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetPayloadFormat(format) }
}

func WebTransportForwardRelayWithPayloadType[T any](payloadType string) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetPayloadType(payloadType) }
}

func WebTransportForwardRelayWithDatagrams[T any](enabled bool) types.Option[*webtransportrelay.ForwardRelay[T]] {
	return func(fr *webtransportrelay.ForwardRelay[T]) { fr.SetUseDatagrams(enabled) }
}

// NewWebTransportForwardRelay creates a WebTransport forward relay.
func NewWebTransportForwardRelay[T any](ctx context.Context, options ...types.Option[*webtransportrelay.ForwardRelay[T]]) *webtransportrelay.ForwardRelay[T] {
	return webtransportrelay.NewForwardRelay[T](ctx, options...)
}

// ---- Receiving relay options ----

func WebTransportReceivingRelayWithAddress[T any](address string) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetAddress(address) }
}

func WebTransportReceivingRelayWithPath[T any](path string) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetPath(path) }
}

func WebTransportReceivingRelayWithOutput[T any](output ...types.Submitter[T]) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.ConnectOutput(output...) }
}

func WebTransportReceivingRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetTLSConfig(config) }
}

func WebTransportReceivingRelayWithPassthrough[T any](enabled bool) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetPassthrough(enabled) }
}

func WebTransportReceivingRelayWithBufferSize[T any](bufferSize uint32) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetBufferSize(bufferSize) }
}

func WebTransportReceivingRelayWithLogger[T any](logger ...types.Logger) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.ConnectLogger(logger...) }
}

func WebTransportReceivingRelayWithComponentMetadata[T any](name string, id string) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetComponentMetadata(name, id) }
}

func WebTransportReceivingRelayWithDecryptionKey[T any](key string) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetDecryptionKey(key) }
}

func WebTransportReceivingRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetAuthenticationOptions(opts) }
}

func WebTransportReceivingRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetStaticHeaders(headers) }
}

func WebTransportReceivingRelayWithDynamicAuthValidator[T any](fn func(ctx context.Context, md map[string]string) error) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetDynamicAuthValidator(fn) }
}

func WebTransportReceivingRelayWithAuthRequired[T any](required bool) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetAuthRequired(required) }
}

func WebTransportReceivingRelayWithMaxFrameBytes[T any](n int) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetMaxFrameBytes(n) }
}

func WebTransportReceivingRelayWithDatagrams[T any](enabled bool) types.Option[*webtransportrelay.ReceivingRelay[T]] {
	return func(rr *webtransportrelay.ReceivingRelay[T]) { rr.SetEnableDatagrams(enabled) }
}

// NewWebTransportReceivingRelay creates a WebTransport receiving relay.
func NewWebTransportReceivingRelay[T any](ctx context.Context, options ...types.Option[*webtransportrelay.ReceivingRelay[T]]) *webtransportrelay.ReceivingRelay[T] {
	return webtransportrelay.NewReceivingRelay[T](ctx, options...)
}

// Auth helpers
func NewWebTransportReceivingRelayOAuth2JWTOptions(issuer, jwksURL string, audience, scopes []string, cacheSeconds int32) *relay.OAuth2Options {
	return NewReceivingRelayOAuth2JWTOptions(issuer, jwksURL, audience, scopes, cacheSeconds)
}

func NewWebTransportReceivingRelayMergeOAuth2Options(dst *relay.OAuth2Options, src *relay.OAuth2Options) *relay.OAuth2Options {
	return NewReceivingRelayMergeOAuth2Options(dst, src)
}

func NewWebTransportReceivingRelayAuthenticationOptionsOAuth2(oauth *relay.OAuth2Options) *relay.AuthenticationOptions {
	return NewReceivingRelayAuthenticationOptionsOAuth2(oauth)
}

func NewWebTransportReceivingRelayAuthenticationOptionsNone() *relay.AuthenticationOptions {
	return NewReceivingRelayAuthenticationOptionsNone()
}

// Forward relay token helpers
func NewWebTransportForwardRelayStaticBearerTokenSource(token string) types.OAuth2TokenSource {
	return NewQuicForwardRelayStaticBearerTokenSource(token)
}

func NewWebTransportForwardRelayEnvBearerTokenSource(envVar string) types.OAuth2TokenSource {
	return NewQuicForwardRelayEnvBearerTokenSource(envVar)
}

func NewWebTransportForwardRelayRefreshingClientCredentialsSource(
	authBaseURL string,
	clientID string,
	clientSecret string,
	scopes []string,
	leeway time.Duration,
	hc *http.Client,
) types.OAuth2TokenSource {
	return NewQuicForwardRelayRefreshingClientCredentialsSource(authBaseURL, clientID, clientSecret, scopes, leeway, hc)
}
