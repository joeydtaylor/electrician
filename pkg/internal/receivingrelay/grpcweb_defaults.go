package receivingrelay

import "github.com/joeydtaylor/electrician/pkg/internal/relay"

func (rr *ReceivingRelay[T]) applyGRPCWebContentTypeDefaults(payload *relay.WrappedPayload, meta *relay.MessageMetadata) {
	if payload == nil || rr.grpcWebConfig == nil {
		return
	}
	if meta == nil {
		meta = &relay.MessageMetadata{}
		payload.Metadata = meta
	}
	if meta.GetContentType() != "" {
		return
	}
	enc := payload.GetPayloadEncoding()
	if enc == relay.PayloadEncoding_PAYLOAD_ENCODING_GOB || enc == relay.PayloadEncoding_PAYLOAD_ENCODING_UNSPECIFIED {
		meta.ContentType = "application/json"
	}
}
