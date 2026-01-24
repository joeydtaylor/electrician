package forwardrelay

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

// WrapPayload serializes data and applies compression/encryption as configured.
func WrapPayload[T any](
	data T,
	perfOpts *relay.PerformanceOptions,
	secOpts *relay.SecurityOptions,
	encryptionKey string,
) (*relay.WrappedPayload, error) {
	if perfOpts == nil {
		perfOpts = &relay.PerformanceOptions{UseCompression: false, CompressionAlgorithm: COMPRESS_NONE}
	}

	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return nil, fmt.Errorf("gob encode failed: %w", err)
	}

	if perfOpts.UseCompression {
		compressed, err := compressData(buf.Bytes(), perfOpts.CompressionAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}
		buf.Reset()
		buf.Write(compressed)
	}

	if secOpts != nil && secOpts.Enabled && secOpts.Suite == ENCRYPT_AES_GCM {
		encrypted, err := encryptData(buf.Bytes(), secOpts, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		buf.Reset()
		buf.Write(encrypted)
	}

	timestamp := timestamppb.Now()
	id := timestamp.AsTime().Format(time.RFC3339Nano)

	metadata := &relay.MessageMetadata{
		Headers: map[string]string{
			"source": "go",
		},
		ContentType: "application/octet-stream",
		Version: &relay.VersionInfo{
			Major: 1,
			Minor: 0,
		},
		Performance: perfOpts,
		Security:    secOpts,
	}

	return &relay.WrappedPayload{
		Id:        id,
		Timestamp: timestamp,
		Payload:   buf.Bytes(),
		Metadata:  metadata,
	}, nil
}

func compressData(data []byte, algorithm relay.CompressionAlgorithm) ([]byte, error) {
	var b bytes.Buffer
	var w io.WriteCloser

	switch algorithm {
	case COMPRESS_DEFLATE:
		w = gzip.NewWriter(&b)
	case COMPRESS_SNAPPY:
		w = snappy.NewBufferedWriter(&b)
	case COMPRESS_ZSTD:
		var err error
		w, err = zstd.NewWriter(&b)
		if err != nil {
			return nil, err
		}
	case COMPRESS_BROTLI:
		w = brotli.NewWriterLevel(&b, brotli.BestCompression)
	case COMPRESS_LZ4:
		w = lz4.NewWriter(&b)
	default:
		return data, nil
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func normalizeAESKey(s string) ([]byte, error) {
	k := []byte(s)
	switch len(k) {
	case 16, 24, 32:
		return k, nil
	default:
		if b, err := hex.DecodeString(s); err == nil && (len(b) == 16 || len(b) == 24 || len(b) == 32) {
			return b, nil
		}
		if b, err := base64.StdEncoding.DecodeString(s); err == nil && (len(b) == 16 || len(b) == 24 || len(b) == 32) {
			return b, nil
		}
		return nil, fmt.Errorf("invalid AES key length: got %d; need 16/24/32 bytes or hex/base64 of those", len(k))
	}
}

func encryptData(data []byte, secOpts *relay.SecurityOptions, key string) ([]byte, error) {
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != ENCRYPT_AES_GCM {
		return data, nil
	}
	aesKey, err := normalizeAESKey(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}
