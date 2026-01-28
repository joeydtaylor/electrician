package quicrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Forward relay options.

func ForwardRelayWithTarget[T any](targets ...string) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetTargets(targets...) }
}

func ForwardRelayWithLogger[T any](logger ...types.Logger) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.ConnectLogger(logger...) }
}

func ForwardRelayWithInput[T any](input ...types.Receiver[T]) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.ConnectInput(input...) }
}

func ForwardRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetTLSConfig(config) }
}

func ForwardRelayWithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetPerformanceOptions(perfOptions) }
}

func ForwardRelayWithSecurityOptions[T any](secOptions *relay.SecurityOptions, encryptionKey string) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetSecurityOptions(secOptions, encryptionKey) }
}

func ForwardRelayWithPassthrough[T any](enabled bool) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetPassthrough(enabled) }
}

func ForwardRelayWithComponentMetadata[T any](name string, id string) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetComponentMetadata(name, id) }
}

func ForwardRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetAuthenticationOptions(opts) }
}

func ForwardRelayWithOAuth2[T any](ts types.OAuth2TokenSource) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetOAuth2(ts) }
}

func ForwardRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetStaticHeaders(headers) }
}

func ForwardRelayWithDynamicHeaders[T any](fn func(ctx context.Context) map[string]string) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetDynamicHeaders(fn) }
}

func ForwardRelayWithAuthRequired[T any](required bool) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetAuthRequired(required) }
}

func ForwardRelayWithAckMode[T any](mode relay.AckMode) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetAckMode(mode) }
}

func ForwardRelayWithAckEveryN[T any](n uint32) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetAckEveryN(n) }
}

func ForwardRelayWithMaxInFlight[T any](n uint32) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetMaxInFlight(n) }
}

func ForwardRelayWithOmitPayloadMetadata[T any](omit bool) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetOmitPayloadMetadata(omit) }
}

func ForwardRelayWithStreamSendBuffer[T any](n int) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetStreamSendBuffer(n) }
}

func ForwardRelayWithMaxFrameBytes[T any](n int) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetMaxFrameBytes(n) }
}

func ForwardRelayWithDropOnFull[T any](drop bool) types.Option[*ForwardRelay[T]] {
	return func(fr *ForwardRelay[T]) { fr.SetDropOnFull(drop) }
}

// Receiving relay options.

func ReceivingRelayWithAddress[T any](address string) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetAddress(address) }
}

func ReceivingRelayWithOutput[T any](output ...types.Submitter[T]) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.ConnectOutput(output...) }
}

func ReceivingRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetTLSConfig(config) }
}

func ReceivingRelayWithPassthrough[T any](enabled bool) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetPassthrough(enabled) }
}

func ReceivingRelayWithBufferSize[T any](bufferSize uint32) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetDataChannel(bufferSize) }
}

func ReceivingRelayWithLogger[T any](logger ...types.Logger) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.ConnectLogger(logger...) }
}

func ReceivingRelayWithComponentMetadata[T any](name string, id string) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetComponentMetadata(name, id) }
}

func ReceivingRelayWithDecryptionKey[T any](key string) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetDecryptionKey(key) }
}

func ReceivingRelayWithAuthenticationOptions[T any](opts *relay.AuthenticationOptions) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetAuthenticationOptions(opts) }
}

func ReceivingRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetStaticHeaders(headers) }
}

func ReceivingRelayWithDynamicAuthValidator[T any](fn func(ctx context.Context, md map[string]string) error) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetDynamicAuthValidator(fn) }
}

func ReceivingRelayWithAuthRequired[T any](required bool) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetAuthRequired(required) }
}

func ReceivingRelayWithMaxFrameBytes[T any](n int) types.Option[*ReceivingRelay[T]] {
	return func(rr *ReceivingRelay[T]) { rr.SetMaxFrameBytes(n) }
}
