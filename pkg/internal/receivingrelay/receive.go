package receivingrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"google.golang.org/grpc/metadata"
)

// Receive handles a single wrapped payload over unary RPC.
func (rr *ReceivingRelay[T]) Receive(ctx context.Context, payload *relay.WrappedPayload) (*relay.StreamAcknowledgment, error) {
	var traceID string
	if payload.GetMetadata() != nil && payload.GetMetadata().GetTraceId() != "" {
		traceID = payload.GetMetadata().GetTraceId()
	} else if md, ok := metadata.FromIncomingContext(ctx); ok && len(md["trace-id"]) > 0 {
		traceID = md["trace-id"][0]
	} else {
		traceID = utils.GenerateUniqueHash()
	}

	rr.logKV(
		types.InfoLevel,
		"Unary payload received",
		"event", "Receive",
		"result", "SUCCESS",
		"trace_id", traceID,
		"id", payload.GetId(),
		"seq", payload.GetSeq(),
		"payload_type", payload.GetPayloadType(),
		"payload_encoding", payload.GetPayloadEncoding(),
		"payload_bytes", len(payload.GetPayload()),
	)

	ack := &relay.StreamAcknowledgment{
		Success:   true,
		Message:   "Received",
		Id:        payload.GetId(),
		Seq:       payload.GetSeq(),
		StreamId:  "",
		Code:      0,
		Retryable: false,
	}

	go func(p *relay.WrappedPayload, tid string) {
		rr.applyGRPCWebContentTypeDefaults(p, p.GetMetadata())
		var data T
		if rr.passthrough {
			var err error
			data, err = rr.asPassthroughItem(p)
			if err != nil {
				rr.logKV(
					types.ErrorLevel,
					"Passthrough failed",
					"event", "Receive",
					"result", "FAILURE",
					"trace_id", tid,
					"id", p.GetId(),
					"seq", p.GetSeq(),
					"error", err,
				)
				return
			}
		} else {
			if err := UnwrapPayload(p, rr.DecryptionKey, &data); err != nil {
				rr.logKV(
					types.ErrorLevel,
					"Unwrap failed",
					"event", "Receive",
					"result", "FAILURE",
					"trace_id", tid,
					"id", p.GetId(),
					"seq", p.GetSeq(),
					"error", err,
				)
				return
			}
		}

		rr.logKV(
			types.InfoLevel,
			"Forwarding to outputs",
			"event", "ReceiveForward",
			"result", "SUCCESS",
			"trace_id", tid,
			"id", p.GetId(),
			"seq", p.GetSeq(),
		)
		select {
		case rr.DataCh <- data:
		case <-rr.ctx.Done():
		}
	}(payload, traceID)

	return ack, nil
}
