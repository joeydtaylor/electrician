//go:build webtransport

package webtransportrelay

import (
	"context"
	"net/http"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

type wtServer[T any] struct {
	rr *ReceivingRelay[T]
	wt *webtransport.Server
}

func newWTServer[T any](rr *ReceivingRelay[T]) (*wtServer[T], error) {
	tlsCfg, err := rr.buildServerTLSConfig()
	if err != nil {
		return nil, err
	}

	wt := &webtransport.Server{
		H3: http3.Server{
			Addr:      rr.Address,
			TLSConfig: tlsCfg,
		},
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	mux := http.NewServeMux()
	mux.HandleFunc(rr.Path, func(w http.ResponseWriter, r *http.Request) {
		sess, err := wt.Upgrade(w, r)
		if err != nil {
			return
		}
		go rr.handleSession(r.Context(), sess, r.Header)
	})

	wt.H3.Handler = mux

	return &wtServer[T]{rr: rr, wt: wt}, nil
}

func (s *wtServer[T]) serve(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.wt.Close()
	}()

	s.rr.logKV(types.InfoLevel, "WebTransport listening",
		"event", "Listen",
		"result", "START",
		"address", s.rr.Address,
		"path", s.rr.Path,
	)

	return s.wt.ListenAndServe()
}

func (s *wtServer[T]) close() error {
	if s.wt != nil {
		return s.wt.Close()
	}
	return nil
}
