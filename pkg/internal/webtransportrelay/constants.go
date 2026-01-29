//go:build webtransport

package webtransportrelay

import "github.com/joeydtaylor/electrician/pkg/internal/relay"

const (
	COMPRESS_NONE    relay.CompressionAlgorithm = 0
	COMPRESS_DEFLATE relay.CompressionAlgorithm = 1
	COMPRESS_SNAPPY  relay.CompressionAlgorithm = 2
	COMPRESS_ZSTD    relay.CompressionAlgorithm = 3
	COMPRESS_BROTLI  relay.CompressionAlgorithm = 4
	COMPRESS_LZ4     relay.CompressionAlgorithm = 5

	ENCRYPT_NONE    relay.EncryptionSuite = 0
	ENCRYPT_AES_GCM relay.EncryptionSuite = 1
)

const defaultMaxFrameBytes = 64 << 20 // 64MB
