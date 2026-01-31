package websocketrelay

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"

	"nhooyr.io/websocket"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type wsServer[T any] struct {
	rr  *ReceivingRelay[T]
	srv *http.Server
	ln  net.Listener
}

func newWSServer[T any](rr *ReceivingRelay[T]) *wsServer[T] {
	return &wsServer[T]{rr: rr}
}

func (s *wsServer[T]) serve(ctx context.Context, ln net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc(s.rr.Path, s.handleWS)

	server := &http.Server{Handler: mux}
	s.srv = server

	if tlsCfg, err := s.rr.buildServerTLSConfig(); err != nil {
		return err
	} else if tlsCfg != nil {
		ln = tls.NewListener(ln, tlsCfg)
	}
	// keep listener
	s.ln = ln

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	s.rr.logKV(types.InfoLevel, "WebSocket listening",
		"event", "Listen",
		"result", "START",
		"address", s.rr.Address,
		"path", s.rr.Path,
		"tls", s.rr.TlsConfig != nil && s.rr.TlsConfig.UseTLS,
	)

	err := server.Serve(ln)
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

func (s *wsServer[T]) close() error {
	if s.srv != nil {
		_ = s.srv.Shutdown(context.Background())
	}
	if s.ln != nil {
		return s.ln.Close()
	}
	return nil
}

func (s *wsServer[T]) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	go s.rr.handleConn(s.rr.ctx, conn, r.Header)
}
