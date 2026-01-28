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
			rr.logKV(types.DebugLevel, "Stream receive failed",
				"event", "StreamReceive",
				"result", "FAILURE",
				"error", err,
			)
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
					rr.logKV(types.ErrorLevel, "Auth failed on stream open",
						"event", "Auth",
						"result", "FAILURE",
						"stream_id", streamID,
						"error", err,
						"headers", headers,
					)
					_ = sendAck(&relay.StreamAcknowledgment{
						Success:  false,
						Message:  "auth failed: " + err.Error(),
						StreamId: streamID,
						Code:     401,
					})
					return
				} else {
					rr.logKV(types.WarnLevel, "Auth policy soft-failed on stream open",
						"event", "Auth",
						"result", "SOFT_FAIL",
						"stream_id", streamID,
						"error", err,
						"headers", headers,
					)
				}
			} else {
				validated = true
			}

			rr.logKV(types.InfoLevel, "Stream opened",
				"event", "StreamOpen",
				"result", "SUCCESS",
				"stream_id", streamID,
				"ack_mode", ackMode,
				"ack_every_n", ackEveryN,
				"max_in_flight", open.GetMaxInFlight(),
				"omit_payload_metadata", open.GetOmitPayloadMetadata(),
			)

			if ackMode != relay.AckMode_ACK_NONE {
				if err := sendAck(&relay.StreamAcknowledgment{Success: true, Message: "Stream open accepted", StreamId: streamID, Code: 0}); err != nil {
					return
				}
			}
			continue
		}

		if closeMsg := env.GetClose(); closeMsg != nil {
			rr.logKV(types.InfoLevel, "Stream closed",
				"event", "StreamClose",
				"result", "SUCCESS",
				"stream_id", streamID,
				"reason", closeMsg.GetReason(),
			)
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
					rr.logKV(types.ErrorLevel, "Auth failed on payload",
						"event", "Auth",
						"result", "FAILURE",
						"stream_id", streamID,
						"id", payload.GetId(),
						"seq", payload.GetSeq(),
						"error", err,
						"headers", headers,
					)
					_ = sendAck(&relay.StreamAcknowledgment{Success: false, Message: "auth failed: " + err.Error(), StreamId: streamID, Id: payload.GetId(), Seq: payload.GetSeq(), Code: 401})
					return
				}
				rr.logKV(types.WarnLevel, "Auth policy soft-failed on payload",
					"event", "Auth",
					"result", "SOFT_FAIL",
					"stream_id", streamID,
					"id", payload.GetId(),
					"seq", payload.GetSeq(),
					"error", err,
					"headers", headers,
				)
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
				rr.logKV(types.ErrorLevel, "Passthrough failed",
					"event", "Passthrough",
					"result", "FAILURE",
					"stream_id", streamID,
					"id", wp.GetId(),
					"seq", seq,
					"error", err,
				)
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
			rr.logKV(types.ErrorLevel, "Unwrap failed",
				"event", "Unwrap",
				"result", "FAILURE",
				"stream_id", streamID,
				"id", wp.GetId(),
				"seq", seq,
				"error", err,
			)
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
		rr.logKV(types.DebugLevel, "Payload received",
			"event", "Receive",
			"result", "SUCCESS",
			"stream_id", streamID,
			"id", wp.GetId(),
			"seq", seq,
			"payload_type", wp.GetPayloadType(),
			"payload_encoding", wp.GetPayloadEncoding(),
			"payload_bytes", len(wp.GetPayload()),
			"trace_id", effectiveMeta.GetTraceId(),
		)

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
