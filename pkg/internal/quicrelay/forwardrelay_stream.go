package quicrelay

import (
	"context"
	"fmt"
	"io"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"github.com/quic-go/quic-go"
)

type streamSession struct {
	target string

	conn   quic.Connection
	stream quic.Stream

	sendCh chan *relay.RelayEnvelope
	doneCh chan struct{}
}

func (fr *ForwardRelay[T]) ensureStreamsInit() {
	fr.streamsMu.Lock()
	defer fr.streamsMu.Unlock()
	if fr.streams == nil {
		fr.streams = make(map[string]*streamSession, len(fr.Targets))
	}
	if fr.streamSendBuf <= 0 {
		fr.streamSendBuf = 8192
	}
}

func (fr *ForwardRelay[T]) getOrCreateStreamSession(ctx context.Context, address string) (*streamSession, error) {
	fr.ensureStreamsInit()

	fr.streamsMu.Lock()
	s := fr.streams[address]
	fr.streamsMu.Unlock()

	if s != nil {
		select {
		case <-s.doneCh:
		default:
			return s, nil
		}

		fr.streamsMu.Lock()
		if fr.streams[address] == s {
			delete(fr.streams, address)
		}
		fr.streamsMu.Unlock()
	}

	ns, err := fr.openStreamSession(ctx, address)
	if err != nil {
		return nil, err
	}

	fr.streamsMu.Lock()
	if existing := fr.streams[address]; existing != nil {
		fr.streamsMu.Unlock()
		_ = ns.close("race lost")
		return existing, nil
	}
	fr.streams[address] = ns
	fr.streamsMu.Unlock()

	return ns, nil
}

func (fr *ForwardRelay[T]) openStreamSession(ctx context.Context, address string) (*streamSession, error) {
	defaults, err := fr.buildStreamDefaults(ctx)
	if err != nil {
		return nil, err
	}

	tlsCfg, err := fr.buildClientTLSConfig()
	if err != nil {
		return nil, err
	}

	qcfg := &quic.Config{EnableDatagrams: false}

	conn, err := quic.DialAddr(ctx, address, tlsCfg, qcfg)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		_ = conn.CloseWithError(0, "open stream failed")
		return nil, fmt.Errorf("open stream failed: %w", err)
	}

	s := &streamSession{
		target: address,
		conn:   conn,
		stream: stream,
		sendCh: make(chan *relay.RelayEnvelope, fr.streamSendBuf),
		doneCh: make(chan struct{}),
	}

	open := &relay.StreamOpen{
		StreamId: utils.GenerateUniqueHash(),
		Defaults: defaults,
		AckMode:  fr.ackMode,
		AckEveryN: func() uint32 {
			if fr.ackEveryN == 0 {
				return 1024
			}
			return fr.ackEveryN
		}(),
		MaxInFlight:         fr.maxInFlight,
		OmitPayloadMetadata: fr.omitPayloadMetadata,
	}

	openEnv := &relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Open{Open: open}}
	if err := writeProtoFrame(stream, openEnv); err != nil {
		_ = stream.Close()
		_ = conn.CloseWithError(0, "send open failed")
		return nil, fmt.Errorf("send open failed: %w", err)
	}

	go fr.streamSendLoop(s)
	go fr.streamAckLoop(s)

	fr.logKV(types.InfoLevel, "QUIC stream opened",
		"event", "StreamOpen",
		"result", "SUCCESS",
		"target", address,
		"stream_id", open.StreamId,
		"ack_mode", open.AckMode,
		"ack_every_n", open.AckEveryN,
		"max_in_flight", open.MaxInFlight,
		"omit_payload_metadata", open.OmitPayloadMetadata,
	)
	return s, nil
}

func (fr *ForwardRelay[T]) streamSendLoop(s *streamSession) {
	defer close(s.doneCh)

	for env := range s.sendCh {
		if err := writeProtoFrame(s.stream, env); err != nil {
			fr.logKV(types.ErrorLevel, "QUIC stream send failed",
				"event", "StreamSend",
				"result", "FAILURE",
				"target", s.target,
				"error", err,
			)
			_ = s.stream.Close()
			_ = s.conn.CloseWithError(0, "send failed")
			return
		}
	}

	_ = s.stream.Close()
	_ = s.conn.CloseWithError(0, "stream closed")
}

func (fr *ForwardRelay[T]) streamAckLoop(s *streamSession) {
	for {
		ack := &relay.StreamAcknowledgment{}
		err := readProtoFrame(s.stream, ack, fr.maxFrameBytes)
		if err == io.EOF {
			return
		}
		if err != nil {
			fr.logKV(types.DebugLevel, "QUIC ack receive ended",
				"event", "StreamAck",
				"result", "END",
				"target", s.target,
				"error", err,
			)
			return
		}
		fr.logKV(types.DebugLevel, "QUIC ack received",
			"event", "StreamAck",
			"result", "SUCCESS",
			"target", s.target,
			"stream_id", ack.GetStreamId(),
			"last_seq", ack.GetLastSeq(),
			"ok_count", ack.GetOkCount(),
			"err_count", ack.GetErrCount(),
			"message", ack.GetMessage(),
			"code", ack.GetCode(),
		)
	}
}

func (s *streamSession) close(reason string) error {
	select {
	case <-s.doneCh:
		return nil
	default:
	}

	closeEnv := &relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Close{Close: &relay.StreamClose{Reason: reason}}}
	select {
	case s.sendCh <- closeEnv:
	default:
	}

	close(s.sendCh)
	<-s.doneCh
	return nil
}

func (fr *ForwardRelay[T]) closeAllStreams(reason string) {
	fr.streamsMu.Lock()
	streams := fr.streams
	fr.streams = nil
	fr.streamsMu.Unlock()

	for _, s := range streams {
		_ = s.close(reason)
	}
}
