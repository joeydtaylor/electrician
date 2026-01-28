package receivingrelay

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

// UnwrapPayload decrypts, decompresses, and decodes a wrapped payload into data.
func UnwrapPayload[T any](wrappedPayload *relay.WrappedPayload, decryptionKey string, data *T) error {
	if wrappedPayload == nil {
		return errors.New("unwrap: nil wrappedPayload")
	}

	var (
		secOpts  *relay.SecurityOptions
		perfOpts *relay.PerformanceOptions
		ct       string
	)
	if wrappedPayload.Metadata != nil {
		secOpts = wrappedPayload.Metadata.Security
		perfOpts = wrappedPayload.Metadata.Performance
		ct = wrappedPayload.Metadata.GetContentType()
	}
	if decryptionKey != "" {
		if secOpts == nil || !secOpts.Enabled || secOpts.Suite != ENCRYPTION_AES_GCM {
			return fmt.Errorf("unwrap: encryption required")
		}
	}

	plaintext, err := decryptData(wrappedPayload.Payload, secOpts, decryptionKey)
	if err != nil {
		return fmt.Errorf("unwrap: decryption failed: %w", err)
	}

	if perfOpts != nil && perfOpts.UseCompression {
		buf, err := decompressData(plaintext, perfOpts.CompressionAlgorithm)
		if err != nil {
			return fmt.Errorf("unwrap: decompression failed: %w", err)
		}
		plaintext = buf.Bytes()
	}

	isJSON := func(s string) bool {
		s = strings.ToLower(strings.TrimSpace(s))
		return s == "application/json" || strings.HasSuffix(s, "+json")
	}
	if isJSON(ct) && !isProtoTarget(data) {
		if err := json.Unmarshal(plaintext, data); err != nil {
			return fmt.Errorf("unwrap: json decode failed: %w", err)
		}
		return nil
	}

	enc := wrappedPayload.GetPayloadEncoding()
	if enc == relay.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED {
		if isProtoTarget(data) {
			enc = relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO
		} else {
			enc = relay.PayloadEncoding_PAYLOAD_ENCODING_GOB
		}
	}

	looksLikeJSON := func(b []byte) bool {
		b = bytes.TrimSpace(b)
		if len(b) == 0 {
			return false
		}
		return b[0] == '{' || b[0] == '['
	}

	switch enc {
	case relay.PayloadEncoding_PAYLOAD_ENCODING_GOB:
		dec := gob.NewDecoder(bytes.NewReader(plaintext))
		if err := dec.Decode(data); err != nil {
			if !isProtoTarget(data) && looksLikeJSON(plaintext) {
				if json.Unmarshal(plaintext, data) == nil {
					return nil
				}
			}
			return fmt.Errorf("unwrap: gob decode failed: %w", err)
		}
		return nil
	case relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO:
		if err := decodeProtoInto(wrappedPayload.GetPayloadType(), plaintext, data); err != nil {
			return fmt.Errorf("unwrap: proto decode failed: %w", err)
		}
		return nil
	default:
		if isJSON(ct) && !isProtoTarget(data) {
			if err := json.Unmarshal(plaintext, data); err == nil {
				return nil
			}
		}
		return fmt.Errorf("unwrap: unsupported payload_encoding: %v", enc)
	}
}
