package receivingrelay

import (
	"fmt"
	"io"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// StreamReceive handles the streaming relay RPC.
func (rr *ReceivingRelay[T]) StreamReceive(stream relay.RelayService_StreamReceiveServer) error {
	ctx := stream.Context()

	var (
		streamID  string
		defaults  *relay.MessageMetadata
		ackMode   relay.AckMode = relay.AckMode_ACK_PER_MESSAGE
		ackEveryN uint64
	)

	var (
		batchOK   uint32
		batchErr  uint32
		lastSeq   uint64
		batchSeen uint64
	)

	sendAck := func(a *relay.StreamAcknowledgment) error {
		return stream.Send(a)
	}

	flushBatch := func(finalMsg string) {
		if ackMode != relay.AckMode_ACK_BATCH {
			return
		}
		if batchOK == 0 && batchErr == 0 {
			return
		}
		_ = sendAck(&relay.StreamAcknowledgment{
			Success:  batchErr == 0,
			Message:  finalMsg,
			StreamId: streamID,
			LastSeq:  lastSeq,
			OkCount:  batchOK,
			ErrCount: batchErr,
		})
		batchOK, batchErr, batchSeen = 0, 0, 0
	}

	traceFallback := ""
	if md, ok := metadata.FromIncomingContext(ctx); ok && len(md["trace-id"]) > 0 {
		traceFallback = md["trace-id"][0]
	}
	if traceFallback == "" {
		traceFallback = utils.GenerateUniqueHash()
	}

	for {
		env, err := stream.Recv()
		if err == io.EOF {
			flushBatch("EOF")
			return nil
		}
		if err != nil {
			code := status.Code(err)
			if code == codes.Canceled || code == codes.DeadlineExceeded || ctx.Err() != nil {
				flushBatch("context done")
				rr.NotifyLoggers(types.DebugLevel, "StreamReceive: ended code=%v err=%v ctx_err=%v", code, err, ctx.Err())
				return nil
			}
			rr.NotifyLoggers(types.ErrorLevel, "StreamReceive: recv failed: %v", err)
			return err
		}

		if open := env.GetOpen(); open != nil {
			streamID = open.GetStreamId()
			defaults = open.GetDefaults()

			if open.GetAckMode() != relay.AckMode_ACK_MODE_UNSPECIFIED {
				ackMode = open.GetAckMode()
			}

			if ackMode == relay.AckMode_ACK_BATCH {
				n := open.GetAckEveryN()
				if n == 0 {
					n = 1024
				}
				ackEveryN = uint64(n)
			} else {
				ackEveryN = 0
			}

			rr.NotifyLoggers(types.InfoLevel, "StreamReceive: open stream_id=%s ack_mode=%v ack_every_n=%d", streamID, ackMode, ackEveryN)

			if ackMode != relay.AckMode_ACK_NONE {
				if err := sendAck(&relay.StreamAcknowledgment{
					Success:  true,
					Message:  "Stream open accepted",
					StreamId: streamID,
					Code:     0,
				}); err != nil {
					return fmt.Errorf("failed to send open ack: %w", err)
				}
			}
			continue
		}

		if closeMsg := env.GetClose(); closeMsg != nil {
			rr.NotifyLoggers(types.InfoLevel, "StreamReceive: close stream_id=%s reason=%s", streamID, closeMsg.GetReason())
			flushBatch("Stream closed")
			return nil
		}

		payload := env.GetPayload()
		if payload == nil {
			continue
		}

		effectiveMeta := payload.GetMetadata()
		if effectiveMeta == nil {
			effectiveMeta = defaults
		}

		traceID := traceFallback
		if effectiveMeta != nil && effectiveMeta.GetTraceId() != "" {
			traceID = effectiveMeta.GetTraceId()
		}

		seq := payload.GetSeq()
		lastSeq = seq

		wp := &relay.WrappedPayload{
			Id:              payload.GetId(),
			Timestamp:       payload.GetTimestamp(),
			Payload:         payload.GetPayload(),
			Metadata:        effectiveMeta,
			ErrorInfo:       payload.GetErrorInfo(),
			Seq:             seq,
			PayloadEncoding: payload.GetPayloadEncoding(),
			PayloadType:     payload.GetPayloadType(),
		}

		var data T
		if err := UnwrapPayload(wp, rr.DecryptionKey, &data); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "StreamReceive: unwrap failed trace_id=%s id=%s seq=%d err=%v", traceID, wp.GetId(), seq, err)

			switch ackMode {
			case relay.AckMode_ACK_PER_MESSAGE:
				if err := sendAck(&relay.StreamAcknowledgment{
					Success:   false,
					Message:   "unwrap failed: " + err.Error(),
					StreamId:  streamID,
					Id:        wp.GetId(),
					Seq:       seq,
					Code:      1,
					Retryable: false,
				}); err != nil {
					return fmt.Errorf("failed to send error ack: %w", err)
				}
			case relay.AckMode_ACK_BATCH:
				batchErr++
				batchSeen++
				if ackEveryN > 0 && batchSeen%ackEveryN == 0 {
					flushBatch("Batch ack")
				}
			case relay.AckMode_ACK_NONE:
			}
			continue
		}

		rr.DataCh <- data

		switch ackMode {
		case relay.AckMode_ACK_PER_MESSAGE:
			if err := sendAck(&relay.StreamAcknowledgment{
				Success:  true,
				Message:  "OK",
				StreamId: streamID,
				Id:       wp.GetId(),
				Seq:      seq,
				Code:     0,
			}); err != nil {
				return fmt.Errorf("failed to send ack: %w", err)
			}
		case relay.AckMode_ACK_BATCH:
			batchOK++
			batchSeen++
			if ackEveryN > 0 && batchSeen%ackEveryN == 0 {
				flushBatch("Batch ack")
			}
		case relay.AckMode_ACK_NONE:
		}
	}
}
