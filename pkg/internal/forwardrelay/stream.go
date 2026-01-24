package forwardrelay

import (
	"context"
	"fmt"
	"io"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type streamSession struct {
	target string

	conn   *grpc.ClientConn
	client relay.RelayServiceClient
	stream relay.RelayService_StreamReceiveClient

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
	outCtx, err := fr.buildPerRPCContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("metadata build failed: %w", err)
	}

	traceID := ""
	if md, ok := metadata.FromOutgoingContext(outCtx); ok && len(md["trace-id"]) > 0 {
		traceID = md["trace-id"][0]
	}
	if traceID == "" {
		traceID = utils.GenerateUniqueHash()
	}

	creds, useInsecure, err := fr.makeDialOptions()
	if err != nil {
		return nil, fmt.Errorf("dial options: %w", err)
	}
	var dialOpts []grpc.DialOption
	if !useInsecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds[0]))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(outCtx, address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	client := relay.NewRelayServiceClient(conn)
	st, err := client.StreamReceive(outCtx)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("stream open failed: %w", err)
	}

	s := &streamSession{
		target: address,
		conn:   conn,
		client: client,
		stream: st,
		sendCh: make(chan *relay.RelayEnvelope, fr.streamSendBuf),
		doneCh: make(chan struct{}),
	}

	open := &relay.StreamOpen{
		StreamId: utils.GenerateUniqueHash(),
		Defaults: &relay.MessageMetadata{
			Headers: map[string]string{
				"source": "go",
			},
			ContentType: "application/octet-stream",
			Version: &relay.VersionInfo{
				Major: 1,
				Minor: 0,
			},
			Performance: fr.PerformanceOptions,
			Security:    fr.SecurityOptions,
			TraceId:     traceID,
		},
		AckMode:             relay.AckMode_ACK_BATCH,
		AckEveryN:           1024,
		MaxInFlight:         uint32(fr.streamSendBuf),
		OmitPayloadMetadata: true,
	}

	openEnv := &relay.RelayEnvelope{
		Msg: &relay.RelayEnvelope_Open{Open: open},
	}

	if err := st.Send(openEnv); err != nil {
		_ = st.CloseSend()
		_ = conn.Close()
		return nil, fmt.Errorf("send open failed: %w", err)
	}

	go fr.streamSendLoop(s)
	go fr.streamAckLoop(s)

	fr.NotifyLoggers(types.InfoLevel, "stream: opened StreamReceive to %s stream_id=%s trace_id=%s", address, open.StreamId, traceID)
	return s, nil
}

func (fr *ForwardRelay[T]) streamSendLoop(s *streamSession) {
	defer close(s.doneCh)

	for env := range s.sendCh {
		if err := s.stream.Send(env); err != nil {
			code := status.Code(err)
			if err == io.EOF || code == codes.Canceled || code == codes.Unavailable {
				fr.NotifyLoggers(types.DebugLevel, "stream: send ended target=%s code=%v err=%v", s.target, code, err)
			} else {
				fr.NotifyLoggers(types.ErrorLevel, "stream: send failed target=%s code=%v err=%v", s.target, code, err)
			}

			_ = s.stream.CloseSend()
			_ = s.conn.Close()
			return
		}
	}

	_ = s.stream.CloseSend()
	_ = s.conn.Close()
}

func (fr *ForwardRelay[T]) streamAckLoop(s *streamSession) {
	for {
		ack, err := s.stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			fr.NotifyLoggers(types.DebugLevel, "stream: ack recv ended target=%s err=%v", s.target, err)
			return
		}
		fr.NotifyLoggers(types.DebugLevel, "stream: ack target=%s stream_id=%s last_seq=%d ok=%d err=%d msg=%s",
			s.target, ack.GetStreamId(), ack.GetLastSeq(), ack.GetOkCount(), ack.GetErrCount(), ack.GetMessage())
	}
}

func (s *streamSession) close(reason string) error {
	select {
	case <-s.doneCh:
		return nil
	default:
	}

	closeEnv := &relay.RelayEnvelope{
		Msg: &relay.RelayEnvelope_Close{Close: &relay.StreamClose{Reason: reason}},
	}
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
		if s != nil {
			_ = s.close(reason)
		}
	}
}
