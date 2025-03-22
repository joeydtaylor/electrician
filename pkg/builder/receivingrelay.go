package builder

import (
	"context"

	rr "github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
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

// ReceivingRelayWithDecryptionKey sets the decryption key for the ReceivingRelay.
// This key is used to decrypt AES-GCM payloads, relying on inbound metadata to
// determine whether decryption is needed.
//
// Parameters:
//   - key: The decryption key.
//
// Returns:
//
//	A functional option that calls SetDecryptionKey on the ReceivingRelay.
func ReceivingRelayWithDecryptionKey[T any](key string) types.Option[types.ReceivingRelay[T]] {
	return rr.WithDecryptionKey[T](key)
}
