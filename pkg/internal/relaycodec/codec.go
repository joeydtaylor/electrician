package relaycodec

import (
	"bytes"
	"compress/gzip"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	FormatJSON  = "json"
	FormatGOB   = "gob"
	FormatProto = "proto"
	FormatBytes = "bytes"
)

func validateEncryptionRequirement(secOpts *relay.SecurityOptions, key string) error {
	if secOpts != nil && secOpts.Enabled && secOpts.Suite == relay.EncryptionSuite_ENCRYPTION_AES_GCM && key == "" {
		return fmt.Errorf("encryption enabled but encryption key is empty")
	}
	if key == "" {
		return nil
	}
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != relay.EncryptionSuite_ENCRYPTION_AES_GCM {
		return fmt.Errorf("encryption key provided but AES-GCM is not enabled")
	}
	return nil
}

// EncodePayload marshals a value into bytes and sets payload metadata.
func EncodePayload[T any](data T, format string) (payload []byte, contentType string, enc relay.PayloadEncoding, err error) {
	switch strings.ToLower(format) {
	case FormatJSON:
		payload, err = json.Marshal(data)
		if err != nil {
			return nil, "", relay.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED, err
		}
		return payload, "application/json", relay.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED, nil
	case FormatProto:
		msg, ok := any(data).(proto.Message)
		if !ok {
			return nil, "", relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO, fmt.Errorf("proto format requires proto.Message")
		}
		payload, err = proto.Marshal(msg)
		if err != nil {
			return nil, "", relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO, err
		}
		return payload, "application/protobuf", relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO, nil
	case FormatBytes:
		b, ok := any(data).([]byte)
		if !ok {
			return nil, "", relay.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED, fmt.Errorf("bytes format requires []byte")
		}
		return b, "application/octet-stream", relay.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED, nil
	case FormatGOB, "":
		var buf bytes.Buffer
		enc := gob.NewEncoder(&buf)
		if err := enc.Encode(data); err != nil {
			return nil, "", relay.PayloadEncoding_PAYLOAD_ENCODING_GOB, err
		}
		return buf.Bytes(), "application/octet-stream", relay.PayloadEncoding_PAYLOAD_ENCODING_GOB, nil
	default:
		return nil, "", relay.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED, fmt.Errorf("unsupported format: %q", format)
	}
}

// WrapPayload builds a WrappedPayload with compression/encryption applied.
func WrapPayload[T any](
	data T,
	format string,
	payloadType string,
	perfOpts *relay.PerformanceOptions,
	secOpts *relay.SecurityOptions,
	encryptionKey string,
	headers map[string]string,
) (*relay.WrappedPayload, error) {
	if perfOpts == nil {
		perfOpts = &relay.PerformanceOptions{UseCompression: false, CompressionAlgorithm: relay.CompressionAlgorithm_COMPRESS_NONE}
	}
	if err := validateEncryptionRequirement(secOpts, encryptionKey); err != nil {
		return nil, err
	}

	payload, contentType, enc, err := EncodePayload(data, format)
	if err != nil {
		return nil, err
	}

	if perfOpts.UseCompression {
		compressed, err := compressData(payload, perfOpts.CompressionAlgorithm)
		if err != nil {
			return nil, fmt.Errorf("compression failed: %w", err)
		}
		payload = compressed
	}

	if secOpts != nil && secOpts.Enabled && secOpts.Suite == relay.EncryptionSuite_ENCRYPTION_AES_GCM {
		encrypted, err := encryptData(payload, secOpts, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("encryption failed: %w", err)
		}
		payload = encrypted
	}

	timestamp := timestamppb.Now()
	id := timestamp.AsTime().Format(time.RFC3339Nano)

	metaHeaders := make(map[string]string, len(headers)+1)
	metaHeaders["source"] = "go"
	for k, v := range headers {
		if k == "" {
			continue
		}
		metaHeaders[k] = v
	}

	metadata := &relay.MessageMetadata{
		Headers:     metaHeaders,
		ContentType: contentType,
		Version:     &relay.VersionInfo{Major: 1, Minor: 0},
		Performance: perfOpts,
		Security:    secOpts,
	}

	return &relay.WrappedPayload{
		Id:              id,
		Timestamp:       timestamp,
		Payload:         payload,
		Metadata:        metadata,
		PayloadEncoding: enc,
		PayloadType:     payloadType,
	}, nil
}

func compressData(data []byte, algorithm relay.CompressionAlgorithm) ([]byte, error) {
	var b bytes.Buffer
	var w io.WriteCloser

	switch algorithm {
	case relay.CompressionAlgorithm_COMPRESS_DEFLATE:
		w = gzip.NewWriter(&b)
	case relay.CompressionAlgorithm_COMPRESS_SNAPPY:
		w = snappy.NewBufferedWriter(&b)
	case relay.CompressionAlgorithm_COMPRESS_ZSTD:
		var err error
		w, err = zstd.NewWriter(&b)
		if err != nil {
			return nil, err
		}
	case relay.CompressionAlgorithm_COMPRESS_BROTLI:
		w = brotli.NewWriterLevel(&b, brotli.BestCompression)
	case relay.CompressionAlgorithm_COMPRESS_LZ4:
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
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != relay.EncryptionSuite_ENCRYPTION_AES_GCM {
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
