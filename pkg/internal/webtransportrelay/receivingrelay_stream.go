//go:build webtransport

package webtransportrelay

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/quic-go/webtransport-go"
	"google.golang.org/protobuf/proto"

	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

func (rr *ReceivingRelay[T]) handleSession(ctx context.Context, sess *webtransport.Session, hdr http.Header) {
	rr.ensureDefaultAuthValidator()

	connHeaders := make(map[string]string)
	for k, v := range hdr {
		if len(v) == 0 {
			continue
		}
		connHeaders[strings.ToLower(k)] = v[0]
	}

	if rr.enableDatagrams {
		go rr.handleDatagrams(ctx, sess, connHeaders)
	}

	for {
		stream, err := sess.AcceptStream(ctx)
		if err != nil {
			return
		}
		go rr.handleStream(ctx, stream, connHeaders)
	}
}

func (rr *ReceivingRelay[T]) handleDatagrams(ctx context.Context, sess *webtransport.Session, connHeaders map[string]string) {
	for {
		b, err := sess.ReceiveDatagram(ctx)
		if err != nil {
			return
		}
		wp := &relay.WrappedPayload{}
		if err := proto.Unmarshal(b, wp); err != nil {
			continue
		}
		rr.handlePayload(ctx, "datagram", nil, wp, connHeaders, nil, nil, nil, nil, nil, relay.AckMode_ACK_NONE, 0, func(*relay.StreamAcknowledgment) error { return nil })
	}
}

func (rr *ReceivingRelay[T]) handleStream(ctx context.Context, stream webtransport.Stream, connHeaders map[string]string) {
	defer func() { _ = stream.Close() }()

	var (
		streamID  = utils.GenerateUniqueHash()
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
		if err != nil {
			flushBatch("stream closed")
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

			headers := mergeHeaders(connHeaders, headersFromMetadata(defaults))
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
					_ = sendAck(&relay.StreamAcknowledgment{Success: false, Message: "auth failed: " + err.Error(), StreamId: streamID, Code: 401})
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

		rr.handlePayload(ctx, streamID, defaults, payload, connHeaders, &validated, &batchOK, &batchErr, &batchSeen, &lastSeq, ackMode, ackEveryN, sendAck)
	}
}

func (rr *ReceivingRelay[T]) handlePayload(
	ctx context.Context,
	streamID string,
	defaults *relay.MessageMetadata,
	payload *relay.WrappedPayload,
	connHeaders map[string]string,
	validated *bool,
	batchOK, batchErr *uint32,
	batchSeen, lastSeq *uint64,
	ackMode relay.AckMode,
	ackEveryN uint64,
	sendAck func(*relay.StreamAcknowledgment) error,
) {
	effectiveMeta := payload.GetMetadata()
	if effectiveMeta == nil {
		effectiveMeta = defaults
	}

	if validated != nil && !*validated {
		headers := mergeHeaders(connHeaders, headersFromMetadata(effectiveMeta))
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
		} else if validated != nil {
			*validated = true
		}
	}

	seq := payload.GetSeq()
	if lastSeq != nil {
		*lastSeq = seq
	}

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

	if rr.passthrough {
		if cast, ok := any(wp).(T); ok {
			rr.DataCh <- cast
			if batchOK != nil {
				*batchOK = *batchOK + 1
			}
		} else {
			if batchErr != nil {
				*batchErr = *batchErr + 1
			}
			_ = sendAck(&relay.StreamAcknowledgment{Success: false, Message: "passthrough type mismatch", StreamId: streamID, Id: wp.GetId(), Seq: wp.GetSeq(), Code: 500})
			return
		}
	} else {
		var out T
		if err := receivingrelay.UnwrapPayload(wp, rr.DecryptionKey, &out); err != nil {
			rr.logKV(types.ErrorLevel, "Unwrap failed",
				"event", "Receive",
				"result", "FAILURE",
				"id", wp.GetId(),
				"seq", wp.GetSeq(),
				"error", err,
			)
			if batchErr != nil {
				*batchErr = *batchErr + 1
			}
			_ = sendAck(&relay.StreamAcknowledgment{Success: false, Message: fmt.Sprintf("unwrap: %v", err), StreamId: streamID, Id: wp.GetId(), Seq: wp.GetSeq(), Code: 400})
			return
		}
		rr.DataCh <- out
		if batchOK != nil {
			*batchOK = *batchOK + 1
		}
	}

	if batchSeen != nil {
		*batchSeen = *batchSeen + 1
	}

	if ackMode == relay.AckMode_ACK_PER_MESSAGE {
		_ = sendAck(&relay.StreamAcknowledgment{Success: true, Message: "ok", StreamId: streamID, Id: wp.GetId(), Seq: wp.GetSeq(), Code: 0})
		return
	}
	if ackMode == relay.AckMode_ACK_BATCH && ackEveryN > 0 && batchSeen != nil && (*batchSeen%ackEveryN == 0) {
		_ = sendAck(&relay.StreamAcknowledgment{
			Success:  batchErr == nil || *batchErr == 0,
			Message:  "batch",
			StreamId: streamID,
			LastSeq: func() uint64 {
				if lastSeq != nil {
					return *lastSeq
				}
				return 0
			}(),
			OkCount: func() uint32 {
				if batchOK != nil {
					return *batchOK
				}
				return 0
			}(),
			ErrCount: func() uint32 {
				if batchErr != nil {
					return *batchErr
				}
				return 0
			}(),
		})
		if batchOK != nil {
			*batchOK = 0
		}
		if batchErr != nil {
			*batchErr = 0
		}
		if batchSeen != nil {
			*batchSeen = 0
		}
	}
}

func mergeHeaders(a, b map[string]string) map[string]string {
	out := make(map[string]string, len(a)+len(b))
	for k, v := range a {
		out[strings.ToLower(k)] = v
	}
	for k, v := range b {
		out[strings.ToLower(k)] = v
	}
	return out
}
