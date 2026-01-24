package receivingrelay

import (
	"bytes"
	"compress/gzip"
	"io"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
)

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

func decompressData(data []byte, algorithm relay.CompressionAlgorithm) (*bytes.Buffer, error) {
	var b bytes.Buffer
	var r io.Reader

	switch algorithm {
	case COMPRESS_DEFLATE:
		var err error
		r, err = gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	case COMPRESS_SNAPPY:
		r = snappy.NewReader(bytes.NewReader(data))
	case COMPRESS_ZSTD:
		var err error
		r, err = zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	case COMPRESS_BROTLI:
		r = brotli.NewReader(bytes.NewReader(data))
	case COMPRESS_LZ4:
		r = lz4.NewReader(bytes.NewReader(data))
	default:
		r = bytes.NewReader(data)
	}

	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}
	return &b, nil
}
