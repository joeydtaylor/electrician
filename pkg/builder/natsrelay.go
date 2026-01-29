//go:build nats

package builder

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/natsrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Forward relay options

func NATSForwardRelayWithURL[T any](url string) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.URL = url }
}

func NATSForwardRelayWithSubject[T any](subject string) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.Subject = subject }
}

func NATSForwardRelayWithLogger[T any](logger ...types.Logger) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.ConnectLogger(logger...) }
}

func NATSForwardRelayWithPerformanceOptions[T any](perfOptions *relay.PerformanceOptions) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.SetPerformanceOptions(perfOptions) }
}

func NATSForwardRelayWithSecurityOptions[T any](secOpts *relay.SecurityOptions, encryptionKey string) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.SetSecurityOptions(secOpts, encryptionKey) }
}

func NATSForwardRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.SetStaticHeaders(headers) }
}

func NATSForwardRelayWithDynamicHeaders[T any](fn func(ctx context.Context) map[string]string) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.SetDynamicHeaders(fn) }
}

func NATSForwardRelayWithPayloadFormat[T any](format string) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.SetPayloadFormat(format) }
}

func NATSForwardRelayWithPayloadType[T any](payloadType string) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.SetPayloadType(payloadType) }
}

func NATSForwardRelayWithPassthrough[T any](enabled bool) types.Option[*natsrelay.ForwardRelay[T]] {
	return func(fr *natsrelay.ForwardRelay[T]) { fr.SetPassthrough(enabled) }
}

// NewNATSForwardRelay creates a NATS forward relay.
func NewNATSForwardRelay[T any](ctx context.Context, options ...types.Option[*natsrelay.ForwardRelay[T]]) *natsrelay.ForwardRelay[T] {
	return natsrelay.NewForwardRelay[T](ctx, options...)
}

// Receiving relay options

func NATSReceivingRelayWithURL[T any](url string) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.SetURL(url) }
}

func NATSReceivingRelayWithSubject[T any](subject string) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.SetSubject(subject) }
}

func NATSReceivingRelayWithQueue[T any](queue string) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.SetQueue(queue) }
}

func NATSReceivingRelayWithOutput[T any](output ...types.Submitter[T]) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.ConnectOutput(output...) }
}

func NATSReceivingRelayWithLogger[T any](logger ...types.Logger) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.ConnectLogger(logger...) }
}

func NATSReceivingRelayWithDecryptionKey[T any](key string) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.SetDecryptionKey(key) }
}

func NATSReceivingRelayWithStaticHeaders[T any](headers map[string]string) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.SetStaticHeaders(headers) }
}

func NATSReceivingRelayWithPassthrough[T any](enabled bool) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.SetPassthrough(enabled) }
}

func NATSReceivingRelayWithBufferSize[T any](bufferSize uint32) types.Option[*natsrelay.ReceivingRelay[T]] {
	return func(rr *natsrelay.ReceivingRelay[T]) { rr.SetBufferSize(bufferSize) }
}

// NewNATSReceivingRelay creates a NATS receiving relay.
func NewNATSReceivingRelay[T any](ctx context.Context, options ...types.Option[*natsrelay.ReceivingRelay[T]]) *natsrelay.ReceivingRelay[T] {
	return natsrelay.NewReceivingRelay[T](ctx, options...)
}
