package websocketrelay

import (
	"context"
	"fmt"
	"net/http"

	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

type wsSession struct {
	target   string
	conn     *websocket.Conn
	sendCh   chan *relay.RelayEnvelope
	doneCh   chan struct{}
	streamID string
}

func (fr *ForwardRelay[T]) ensureSessionsInit() {
	fr.streamsMu.Lock()
	defer fr.streamsMu.Unlock()
	if fr.streams == nil {
		fr.streams = make(map[string]*wsSession, len(fr.Targets))
	}
	if fr.streamSendBuf <= 0 {
		fr.streamSendBuf = 8192
	}
}

func (fr *ForwardRelay[T]) getOrCreateSession(ctx context.Context, target string) (*wsSession, error) {
	fr.ensureSessionsInit()

	fr.streamsMu.Lock()
	s := fr.streams[target]
	fr.streamsMu.Unlock()

	if s != nil {
		select {
		case <-s.doneCh:
		default:
			return s, nil
		}

		fr.streamsMu.Lock()
		if fr.streams[target] == s {
			delete(fr.streams, target)
		}
		fr.streamsMu.Unlock()
	}

	ns, err := fr.openSession(ctx, target)
	if err != nil {
		return nil, err
	}

	fr.streamsMu.Lock()
	if existing := fr.streams[target]; existing != nil {
		fr.streamsMu.Unlock()
		_ = ns.close("race lost")
		return existing, nil
	}
	fr.streams[target] = ns
	fr.streamsMu.Unlock()

	return ns, nil
}

func (fr *ForwardRelay[T]) openSession(ctx context.Context, target string) (*wsSession, error) {
	defaults, err := fr.buildStreamDefaults(ctx)
	if err != nil {
		return nil, err
	}

	headers := http.Header{}
	for k, v := range defaults.GetHeaders() {
		headers.Set(k, v)
	}

	var httpClient *http.Client
	if fr.TlsConfig != nil && fr.TlsConfig.UseTLS {
		tlsCfg, err := fr.buildClientTLSConfig()
		if err != nil {
			return nil, err
		}
		httpClient = &http.Client{Transport: &http.Transport{TLSClientConfig: tlsCfg}}
	}

	opts := &websocket.DialOptions{HTTPHeader: headers, HTTPClient: httpClient}
	conn, _, err := websocket.Dial(ctx, target, opts)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	s := &wsSession{
		target:   target,
		conn:     conn,
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
	if err := fr.sendEnvelope(ctx, conn, openEnv); err != nil {
		_ = conn.Close(websocket.StatusInternalError, "send open failed")
		return nil, fmt.Errorf("send open failed: %w", err)
	}

	go fr.sessionSendLoop(s)
	go fr.sessionAckLoop(s)

	fr.logKV(types.InfoLevel, "WebSocket stream opened",
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

func (fr *ForwardRelay[T]) sendEnvelope(ctx context.Context, conn *websocket.Conn, env *relay.RelayEnvelope) error {
	b, err := proto.Marshal(env)
	if err != nil {
		return err
	}
	if len(b) == 0 {
		return fmt.Errorf("empty envelope")
	}
	return conn.Write(ctx, websocket.MessageBinary, b)
}

func (fr *ForwardRelay[T]) sessionSendLoop(s *wsSession) {
	defer close(s.doneCh)

	for env := range s.sendCh {
		if err := fr.sendEnvelope(fr.ctx, s.conn, env); err != nil {
			fr.logKV(types.ErrorLevel, "WebSocket send failed",
				"event", "StreamSend",
				"result", "FAILURE",
				"target", s.target,
				"error", err,
			)
			_ = s.conn.Close(websocket.StatusInternalError, "send failed")
			return
		}
	}

	_ = s.conn.Close(websocket.StatusNormalClosure, "stream closed")
}

func (fr *ForwardRelay[T]) sessionAckLoop(s *wsSession) {
	for {
		_, b, err := s.conn.Read(fr.ctx)
		if err != nil {
			return
		}
		ack := &relay.StreamAcknowledgment{}
		if err := proto.Unmarshal(b, ack); err != nil {
			continue
		}
		fr.logKV(types.DebugLevel, "WebSocket ack received",
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

func (s *wsSession) close(reason string) error {
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
