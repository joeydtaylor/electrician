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

// ForwardRelayWithAddress sets the network address for the ForwardRelay.
func ForwardRelayWithTarget[T any](t ...string) types.Option[types.ForwardRelay[T]] {
	return fr.WithTarget[T](t...)
}

// ForwardRelayWithComponentMetadata adds component metadata overrides.
func ForwardRelayWithComponentMetadata[T any](name string, id string) types.Option[types.ForwardRelay[T]] {
	return fr.WithComponentMetadata[T](name, id)
}

// ForwardRelayWithInput sets the input conduit for the ForwardRelay.
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

// ForwardRelayWithSecurityOptions sets security (encryption) options for the ForwardRelay.
// It mirrors other builder functions, calling an internal WithSecurityOptions under the hood.
//
// Parameters:
//   - secOpts: A pointer to relay.SecurityOptions specifying whether encryption is enabled, which suite, etc.
//   - encryptionKey: The encryption key to use for AES-GCM, if enabled.
//
// Returns:
//
//	A builder option conforming to types.Option[types.ForwardRelay[T]].
func ForwardRelayWithSecurityOptions[T any](secOpts *relay.SecurityOptions, encryptionKey string) types.Option[types.ForwardRelay[T]] {
	return fr.WithSecurityOptions[T](secOpts, encryptionKey)
}

// ForwardRelayWithTLSConfig sets the TLS configuration for the ForwardRelay.
func ForwardRelayWithTLSConfig[T any](config *types.TLSConfig) types.Option[types.ForwardRelay[T]] {
	return fr.WithTLSConfig[T](config)
}

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

// NewSecurityOptions creates a new SecurityOptions object indicating
// whether encryption is enabled and which encryption suite to use.
//
// Parameters:
//   - enabled:      Whether encryption is enabled at all.
//   - suite:        Which suite to use (e.g. ENCRYPTION_AES_GCM).
//
// Returns:
//   - A pointer to a relay.SecurityOptions struct, usable when building
//     a ForwardRelay via ForwardRelayWithSecurityOptions.
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
