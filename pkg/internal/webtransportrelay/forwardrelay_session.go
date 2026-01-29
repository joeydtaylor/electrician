//go:build webtransport

package webtransportrelay

import (
	"context"
	"fmt"
	"net/http"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

type wtSession struct {
	target   string
	sess     *webtransport.Session
	stream   webtransport.Stream
	sendCh   chan *relay.RelayEnvelope
	doneCh   chan struct{}
	streamID string
}

func (fr *ForwardRelay[T]) ensureSessionsInit() {
	fr.sessionsMu.Lock()
	defer fr.sessionsMu.Unlock()
	if fr.sessions == nil {
		fr.sessions = make(map[string]*wtSession, len(fr.Targets))
	}
	if fr.streamSendBuf <= 0 {
		fr.streamSendBuf = 8192
	}
}

func (fr *ForwardRelay[T]) getOrCreateSession(ctx context.Context, target string) (*wtSession, error) {
	fr.ensureSessionsInit()

	fr.sessionsMu.Lock()
	s := fr.sessions[target]
	fr.sessionsMu.Unlock()

	if s != nil {
		select {
		case <-s.doneCh:
		default:
			return s, nil
		}

		fr.sessionsMu.Lock()
		if fr.sessions[target] == s {
			delete(fr.sessions, target)
		}
		fr.sessionsMu.Unlock()
	}

	ns, err := fr.openSession(ctx, target)
	if err != nil {
		return nil, err
	}

	fr.sessionsMu.Lock()
	if existing := fr.sessions[target]; existing != nil {
		fr.sessionsMu.Unlock()
		_ = ns.close("race lost")
		return existing, nil
	}
	fr.sessions[target] = ns
	fr.sessionsMu.Unlock()

	return ns, nil
}

func (fr *ForwardRelay[T]) openSession(ctx context.Context, target string) (*wtSession, error) {
	defaults, err := fr.buildStreamDefaults(ctx)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	for k, v := range defaults.GetHeaders() {
		headers.Set(k, v)
	}

	tlsCfg, err := fr.buildClientTLSConfig()
	if err != nil {
		return nil, err
	}
	rt := &http3.RoundTripper{TLSClientConfig: tlsCfg}
	dialer := webtransport.Dialer{RoundTripper: rt}

	sess, _, err := dialer.Dial(ctx, target, headers)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	stream, err := sess.OpenStreamSync(ctx)
	if err != nil {
		_ = sess.CloseWithError(0, "open stream failed")
		return nil, fmt.Errorf("open stream failed: %w", err)
	}

	s := &wtSession{
		target:   target,
		sess:     sess,
		stream:   stream,
		sendCh:   make(chan *relay.RelayEnvelope, fr.streamSendBuf),
		doneCh:   make(chan struct{}),
		streamID: utils.GenerateUniqueHash(),
	}

	open := &relay.StreamOpen{
		StreamId: s.streamID,
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
		_ = sess.CloseWithError(0, "send open failed")
		return nil, fmt.Errorf("send open failed: %w", err)
	}

	go fr.sessionSendLoop(s)
	go fr.sessionAckLoop(s)
	if fr.useDatagrams {
		go fr.sessionDatagramLoop(s)
	}

	fr.logKV(types.InfoLevel, "WebTransport stream opened",
		"event", "StreamOpen",
		"result", "SUCCESS",
		"target", target,
		"stream_id", s.streamID,
		"ack_mode", open.AckMode,
		"ack_every_n", open.AckEveryN,
		"max_in_flight", open.MaxInFlight,
		"omit_payload_metadata", open.OmitPayloadMetadata,
	)

	return s, nil
}

func (fr *ForwardRelay[T]) sessionSendLoop(s *wtSession) {
	defer close(s.doneCh)

	for env := range s.sendCh {
		if err := writeProtoFrame(s.stream, env); err != nil {
			fr.logKV(types.ErrorLevel, "WebTransport stream send failed",
				"event", "StreamSend",
				"result", "FAILURE",
				"target", s.target,
				"error", err,
			)
			_ = s.stream.Close()
			_ = s.sess.CloseWithError(0, "send failed")
			return
		}
	}

	_ = s.stream.Close()
	_ = s.sess.CloseWithError(0, "stream closed")
}

func (fr *ForwardRelay[T]) sessionAckLoop(s *wtSession) {
	for {
		ack := &relay.StreamAcknowledgment{}
		err := readProtoFrame(s.stream, ack, fr.maxFrameBytes)
		if err != nil {
			return
		}
		fr.logKV(types.DebugLevel, "WebTransport ack received",
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

func (fr *ForwardRelay[T]) sessionDatagramLoop(s *wtSession) {
	// No-op: acknowledgments are stream-based; datagrams are fire-and-forget.
}

func (s *wtSession) close(reason string) error {
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

func (fr *ForwardRelay[T]) closeAllSessions(reason string) {
	fr.sessionsMu.Lock()
	sessions := fr.sessions
	fr.sessions = nil
	fr.sessionsMu.Unlock()

	for _, s := range sessions {
		_ = s.close(reason)
	}
}
