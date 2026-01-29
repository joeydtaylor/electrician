package websocketrelay

import (
	"context"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/relaycodec"
)

func (fr *ForwardRelay[T]) buildStreamDefaults(ctx context.Context) (*relay.MessageMetadata, error) {
	headers := cloneStringMap(fr.staticHeaders)
	if fr.dynamicHeaders != nil {
		for k, v := range fr.dynamicHeaders(ctx) {
			headers[k] = v
		}
	}
	if fr.tokenSource != nil {
		tok, err := fr.tokenSource.AccessToken(ctx)
		if err != nil {
			if fr.authRequired {
				return nil, err
			}
		} else if tok != "" {
			headers["authorization"] = "Bearer " + tok
		}
	}

	meta := &relay.MessageMetadata{
		Headers:     headers,
		Version:     &relay.VersionInfo{Major: 1, Minor: 0},
		Performance: fr.PerformanceOptions,
		Security:    fr.SecurityOptions,
	}

	// Default content type for stream (used when omitPayloadMetadata=true)
	switch strings.ToLower(fr.payloadFormat) {
	case relaycodec.FormatJSON:
		meta.ContentType = "application/json"
	case relaycodec.FormatProto:
		meta.ContentType = "application/protobuf"
	default:
		meta.ContentType = "application/octet-stream"
	}

	return meta, nil
}

func (fr *ForwardRelay[T]) wrapItem(ctx context.Context, item T) (*relay.WrappedPayload, error) {
	if fr.passthrough {
		if wp, ok := any(item).(*relay.WrappedPayload); ok && wp != nil {
			return wp, nil
		}
		return nil, fmt.Errorf("passthrough enabled but item is not *relay.WrappedPayload")
	}

	headers := cloneStringMap(fr.staticHeaders)
	if fr.dynamicHeaders != nil {
		for k, v := range fr.dynamicHeaders(ctx) {
			headers[k] = v
		}
	}
	if fr.tokenSource != nil {
		tok, err := fr.tokenSource.AccessToken(ctx)
		if err != nil {
			if fr.authRequired {
				return nil, err
			}
		} else if tok != "" {
			headers["authorization"] = "Bearer " + tok
		}
	}

	wp, err := relaycodec.WrapPayload(item, fr.payloadFormat, fr.payloadType, fr.PerformanceOptions, fr.SecurityOptions, fr.EncryptionKey, headers)
	if err != nil {
		return nil, err
	}
	wp.Seq = atomic.AddUint64(&fr.seq, 1)
	if fr.omitPayloadMetadata {
		wp.Metadata = nil
	}
	return wp, nil
}
