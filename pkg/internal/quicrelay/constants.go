package quicrelay

import "github.com/joeydtaylor/electrician/pkg/internal/relay"

// Compression algorithms (mirror relay.CompressionAlgorithm).
const (
	COMPRESS_NONE    relay.CompressionAlgorithm = 0
	COMPRESS_DEFLATE relay.CompressionAlgorithm = 1
	COMPRESS_SNAPPY  relay.CompressionAlgorithm = 2
	COMPRESS_ZSTD    relay.CompressionAlgorithm = 3
	COMPRESS_BROTLI  relay.CompressionAlgorithm = 4
	COMPRESS_LZ4     relay.CompressionAlgorithm = 5

	ENCRYPTION_NONE    relay.EncryptionSuite = 0
	ENCRYPTION_AES_GCM relay.EncryptionSuite = 1
)
