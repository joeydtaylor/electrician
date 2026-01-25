package receivingrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
)

// WithAddress sets the listen address.
func WithAddress[T any](address string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetAddress(address) }
}

// WithOutput configures output submitters.
func WithOutput[T any](output ...types.Submitter[T]) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.ConnectOutput(output...) }
}

// WithTLSConfig sets TLS configuration.
func WithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetTLSConfig(config) }
}

// WithPassthrough enables forwarding raw WrappedPayload values without unwrap/decrypt.
func WithPassthrough[T any](enabled bool) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetPassthrough(enabled) }
}

// WithGRPCWebConfig configures gRPC-Web CORS and transport behavior.
func WithGRPCWebConfig[T any](config *types.GRPCWebConfig) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetGRPCWebConfig(config) }
}

// WithBufferSize adjusts the internal data channel buffer size.
func WithBufferSize[T any](bufferSize uint32) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetDataChannel(bufferSize) }
}

// WithLogger attaches loggers.
func WithLogger[T any](logger ...types.Logger) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.ConnectLogger(logger...) }
}

// WithComponentMetadata sets name and ID metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) { rr.SetComponentMetadata(name, id) }
}

// WithDecryptionKey sets the decryption key.
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

// WithAuthInterceptors installs unary and stream interceptors on the server.
func WithAuthInterceptors[T any](
	unary grpc.UnaryServerInterceptor,
	stream grpc.StreamServerInterceptor,
) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetAuthInterceptors(unary, stream)
	}
}

// WithAuthRequired toggles strict auth enforcement.
func WithAuthRequired[T any](required bool) types.Option[types.ReceivingRelay[T]] {
	return func(rr types.ReceivingRelay[T]) {
		rr.SetAuthRequired(required)
	}
}
