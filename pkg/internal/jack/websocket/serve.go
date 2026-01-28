package websocket

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"nhooyr.io/websocket"
)

// Serve starts the WebSocket server and dispatches inbound messages to submit.
func (s *serverAdapter[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	if submit == nil {
		return errors.New("submit cannot be nil")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	s.serverMu.Lock()
	defer s.serverMu.Unlock()

	if s.server != nil {
		return fmt.Errorf("server already started")
	}

	cfg := s.snapshotConfig()
	if cfg.tlsConfigErr != nil {
		return cfg.tlsConfigErr
	}
	if cfg.endpoint == "" {
		return errors.New("endpoint not configured")
	}
	if err := validateFormat(cfg.format); err != nil {
		return err
	}

	atomic.StoreInt32(&s.configFrozen, 1)

	handler := s.buildHandler(ctx, cfg, submit)

	s.server = &http.Server{
		Addr:      cfg.address,
		Handler:   handler,
		TLSConfig: cfg.tlsConfig,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	errCh := make(chan error, 1)
	go func() {
		var err error
		if cfg.tlsConfig != nil {
			s.NotifyLoggers(
				types.InfoLevel,
				"Serve: starting WebSocket server",
				"component", s.componentMetadata,
				"event", "ServeStart",
				"tls", true,
				"address", cfg.address,
				"endpoint", cfg.endpoint,
			)
			err = s.server.ListenAndServeTLS("", "")
		} else {
			s.NotifyLoggers(
				types.InfoLevel,
				"Serve: starting WebSocket server",
				"component", s.componentMetadata,
				"event", "ServeStart",
				"tls", false,
				"address", cfg.address,
				"endpoint", cfg.endpoint,
			)
			err = s.server.ListenAndServe()
		}
		errCh <- err
	}()

	s.notifyLifecycleStart()
	defer s.notifyLifecycleStop()
	defer s.closeAllConnections("server shutting down")

	s.startOutputFanout(ctx)

	select {
	case <-ctx.Done():
		s.NotifyLoggers(
			types.WarnLevel,
			"Serve: context canceled, shutting down",
			"component", s.componentMetadata,
			"event", "ServeStop",
			"result", "CANCELLED",
		)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.server.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			s.NotifyLoggers(
				types.ErrorLevel,
				"Serve: server error",
				"component", s.componentMetadata,
				"event", "ServeError",
				"error", err,
			)
			return err
		}
		return nil
	}
}

// Broadcast sends a message to all active connections.
func (s *serverAdapter[T]) Broadcast(ctx context.Context, msg T) error {
	_ = ctx
	conns := s.snapshotConns()
	if len(conns) == 0 {
		return nil
	}

	messageType, payload, err := encodeMessage(s.snapshotConfig().format, msg)
	if err != nil {
		return err
	}

	out := outboundMessage{messageType: messageType, payload: payload}
	dropped := 0
	for _, conn := range conns {
		if conn == nil {
			continue
		}
		if !conn.enqueue(out) {
			dropped++
		}
	}

	if dropped > 0 {
		return fmt.Errorf("broadcast dropped to %d connection(s)", dropped)
	}
	return nil
}

func (s *serverAdapter[T]) buildHandler(baseCtx context.Context, cfg serverConfig, submit func(context.Context, T) error) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(cfg.endpoint, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if cfg.maxConnections > 0 && s.connectionCount() >= cfg.maxConnections {
			http.Error(w, "Too Many Connections", http.StatusServiceUnavailable)
			return
		}

		if err := s.authorizeRequest(r.Context(), r, cfg); err != nil {
			s.NotifyLoggers(
				types.WarnLevel,
				"Auth: request rejected",
				"component", s.componentMetadata,
				"event", "AuthReject",
				"error", err,
			)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		for key, val := range cfg.headers {
			w.Header().Set(key, val)
		}

		acceptOptions := &websocket.AcceptOptions{}
		if len(cfg.allowedOrigins) > 0 {
			acceptOptions.OriginPatterns = cfg.allowedOrigins
		}

		conn, err := websocket.Accept(w, r, acceptOptions)
		if err != nil {
			s.NotifyLoggers(
				types.ErrorLevel,
				"Accept: error",
				"component", s.componentMetadata,
				"event", "AcceptError",
				"error", err,
			)
			return
		}

		if cfg.readLimit > 0 {
			conn.SetReadLimit(cfg.readLimit)
		}

		wsConn, err := s.addConn(conn, cfg)
		if err != nil {
			_ = conn.Close(websocket.StatusPolicyViolation, err.Error())
			s.NotifyLoggers(
				types.WarnLevel,
				"Accept: rejected connection",
				"component", s.componentMetadata,
				"event", "AcceptReject",
				"error", err,
			)
			return
		}

		s.NotifyLoggers(
			types.InfoLevel,
			"Connection accepted",
			"component", s.componentMetadata,
			"event", "ConnectionAccepted",
			"remote", r.RemoteAddr,
		)

		go s.runConn(baseCtx, cfg, wsConn, submit, r.RemoteAddr)
	})

	return mux
}

func (s *serverAdapter[T]) runConn(ctx context.Context, cfg serverConfig, conn *wsConn, submit func(context.Context, T) error, remote string) {
	defer s.dropConn(conn)

	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		if err := s.writeLoop(connCtx, cfg, conn); err != nil {
			s.NotifyLoggers(
				types.WarnLevel,
				"WriteLoop error",
				"component", s.componentMetadata,
				"event", "WriteLoop",
				"error", err,
			)
		}
	}()

	if err := s.readLoop(connCtx, cfg, conn, submit); err != nil {
		s.NotifyLoggers(
			types.WarnLevel,
			"ReadLoop error",
			"component", s.componentMetadata,
			"event", "ReadLoop",
			"error", err,
		)
	}

	conn.close(websocket.StatusNormalClosure, "connection closed")
	s.NotifyLoggers(
		types.InfoLevel,
		"Connection closed",
		"component", s.componentMetadata,
		"event", "ConnectionClosed",
		"remote", remote,
	)
}

func (s *serverAdapter[T]) readLoop(ctx context.Context, cfg serverConfig, conn *wsConn, submit func(context.Context, T) error) error {
	for {
		readCtx := ctx
		var cancel context.CancelFunc
		if cfg.idleTimeout > 0 {
			readCtx, cancel = context.WithTimeout(ctx, cfg.idleTimeout)
		}
		msgType, payload, err := conn.conn.Read(readCtx)
		if cancel != nil {
			cancel()
		}
		if err != nil {
			return s.handleReadError(err)
		}

		if msgType == websocket.MessageBinary && (cfg.format == types.WebSocketFormatText || cfg.format == types.WebSocketFormatJSON) {
			s.NotifyLoggers(
				types.DebugLevel,
				"ReadLoop: received binary payload for text/json format",
				"component", s.componentMetadata,
				"event", "ReadLoopFormatMismatch",
				"message_type", "binary",
				"expected_format", cfg.format,
			)
		}
		if msgType == websocket.MessageText && (cfg.format == types.WebSocketFormatBinary || cfg.format == types.WebSocketFormatProto) {
			s.NotifyLoggers(
				types.DebugLevel,
				"ReadLoop: received text payload for binary/proto format",
				"component", s.componentMetadata,
				"event", "ReadLoopFormatMismatch",
				"message_type", "text",
				"expected_format", cfg.format,
			)
		}

		msg, err := decodeMessage[T](cfg.format, payload)
		if err != nil {
			s.NotifyLoggers(
				types.WarnLevel,
				"ReadLoop: decode error",
				"component", s.componentMetadata,
				"event", "Decode",
				"error", err,
			)
			continue
		}

		if err := submit(ctx, msg); err != nil {
			s.NotifyLoggers(
				types.WarnLevel,
				"ReadLoop: submit error",
				"component", s.componentMetadata,
				"event", "Submit",
				"error", err,
			)
		}
	}
}

func (s *serverAdapter[T]) handleReadError(err error) error {
	if err == nil {
		return nil
	}
	status := websocket.CloseStatus(err)
	if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}

func (s *serverAdapter[T]) writeLoop(ctx context.Context, cfg serverConfig, conn *wsConn) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-conn.send:
			if !ok {
				return nil
			}
			writeCtx := ctx
			var cancel context.CancelFunc
			if cfg.writeTimeout > 0 {
				writeCtx, cancel = context.WithTimeout(ctx, cfg.writeTimeout)
			}
			if err := conn.conn.Write(writeCtx, msg.messageType, msg.payload); err != nil {
				if cancel != nil {
					cancel()
				}
				return err
			}
			if cancel != nil {
				cancel()
			}
		}
	}
}
