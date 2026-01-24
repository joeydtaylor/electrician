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

	rr.NotifyLoggers(types.InfoLevel, "Receive: unary payload trace_id=%s", traceID)

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
		var data T
		if err := UnwrapPayload(p, rr.DecryptionKey, &data); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "Receive: unwrap failed trace_id=%s err=%v", tid, err)
			return
		}

		rr.NotifyLoggers(types.InfoLevel, "Receive: forwarding trace_id=%s", tid)
		select {
		case rr.DataCh <- data:
		case <-rr.ctx.Done():
		}
	}(payload, traceID)

	return ack, nil
}
