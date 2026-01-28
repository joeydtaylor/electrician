package quicrelay

import (
	"fmt"
	"io"

	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/quic-go/quic-go"
)

func (rr *ReceivingRelay[T]) handleStream(stream quic.Stream) {
	defer func() {
		_ = stream.Close()
	}()

	rr.ensureDefaultAuthValidator()

	var (
		streamID string
		defaults *relay.MessageMetadata
		ackMode  relay.AckMode = relay.AckMode_ACK_PER_MESSAGE
		ackEveryN uint64
	)

	var (
		batchOK   uint32
		batchErr  uint32
		lastSeq   uint64
		batchSeen uint64
	)

	sendAck := func(a *relay.StreamAcknowledgment) error {
		return writeProtoFrame(stream, a)
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

	validated := false

	for {
		env := &relay.RelayEnvelope{}
		err := readProtoFrame(stream, env, rr.maxFrameBytes)
		if err == io.EOF {
			flushBatch("EOF")
			return
		}
		if err != nil {
			rr.NotifyLoggers(types.DebugLevel, "StreamReceive: recv failed: %v", err)
			flushBatch("recv error")
			return
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
			}

			headers := headersFromMetadata(defaults)
			if err := rr.validateHeaders(rr.ctx, headers); err != nil {
				if err == errMissingHeaders {
					validated = false
				} else if rr.authRequired {
					_ = sendAck(&relay.StreamAcknowledgment{
						Success:  false,
						Message:  "auth failed: " + err.Error(),
						StreamId: streamID,
						Code:     401,
					})
					return
				} else {
					rr.NotifyLoggers(types.WarnLevel, "Auth: soft-failing policy error: %v", err)
				}
			} else {
				validated = true
			}

			rr.NotifyLoggers(types.InfoLevel, "StreamReceive: open stream_id=%s ack_mode=%v ack_every_n=%d", streamID, ackMode, ackEveryN)

			if ackMode != relay.AckMode_ACK_NONE {
				if err := sendAck(&relay.StreamAcknowledgment{Success: true, Message: "Stream open accepted", StreamId: streamID, Code: 0}); err != nil {
					return
				}
			}
			continue
		}

		if closeMsg := env.GetClose(); closeMsg != nil {
			rr.NotifyLoggers(types.InfoLevel, "StreamReceive: close stream_id=%s reason=%s", streamID, closeMsg.GetReason())
			flushBatch("Stream closed")
			return
		}

		payload := env.GetPayload()
		if payload == nil {
			continue
		}

		effectiveMeta := payload.GetMetadata()
		if effectiveMeta == nil {
			effectiveMeta = defaults
		}

		if !validated {
			headers := headersFromMetadata(effectiveMeta)
			if err := rr.validateHeaders(rr.ctx, headers); err != nil {
				if rr.authRequired {
					_ = sendAck(&relay.StreamAcknowledgment{Success: false, Message: "auth failed: " + err.Error(), StreamId: streamID, Id: payload.GetId(), Seq: payload.GetSeq(), Code: 401})
					return
				}
				rr.NotifyLoggers(types.WarnLevel, "Auth: soft-failing policy error: %v", err)
			} else {
				validated = true
			}
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
		if rr.passthrough {
			var err error
			data, err = rr.asPassthroughItem(wp)
			if err != nil {
				rr.NotifyLoggers(types.ErrorLevel, "StreamReceive: passthrough failed id=%s seq=%d err=%v", wp.GetId(), seq, err)
				switch ackMode {
				case relay.AckMode_ACK_PER_MESSAGE:
					_ = sendAck(&relay.StreamAcknowledgment{Success: false, Message: "passthrough failed: " + err.Error(), StreamId: streamID, Id: wp.GetId(), Seq: seq, Code: 1, Retryable: false})
				case relay.AckMode_ACK_BATCH:
					batchErr++
					batchSeen++
					if ackEveryN > 0 && batchSeen%ackEveryN == 0 {
						flushBatch("Batch ack")
					}
				}
				continue
			}
		} else if err := receivingrelay.UnwrapPayload(wp, rr.DecryptionKey, &data); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "StreamReceive: unwrap failed id=%s seq=%d err=%v", wp.GetId(), seq, err)
			switch ackMode {
			case relay.AckMode_ACK_PER_MESSAGE:
				_ = sendAck(&relay.StreamAcknowledgment{Success: false, Message: "unwrap failed: " + err.Error(), StreamId: streamID, Id: wp.GetId(), Seq: seq, Code: 1, Retryable: false})
			case relay.AckMode_ACK_BATCH:
				batchErr++
				batchSeen++
				if ackEveryN > 0 && batchSeen%ackEveryN == 0 {
					flushBatch("Batch ack")
				}
			}
			continue
		}

		rr.DataCh <- data

		switch ackMode {
		case relay.AckMode_ACK_PER_MESSAGE:
			if err := sendAck(&relay.StreamAcknowledgment{Success: true, Message: "OK", StreamId: streamID, Id: wp.GetId(), Seq: seq, Code: 0}); err != nil {
				return
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

func (rr *ReceivingRelay[T]) asPassthroughItem(payload *relay.WrappedPayload) (T, error) {
	var zero T
	if payload == nil {
		return zero, fmt.Errorf("nil payload")
	}

	switch any(zero).(type) {
	case relay.WrappedPayload:
		return any(*payload).(T), nil
	case *relay.WrappedPayload:
		return any(payload).(T), nil
	default:
		return zero, fmt.Errorf("passthrough requires relay.WrappedPayload or *relay.WrappedPayload, got %T", zero)
	}
}
