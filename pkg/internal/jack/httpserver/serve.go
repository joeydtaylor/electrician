package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type serverConfig struct {
	address       string
	method        string
	endpoint      string
	headers       map[string]string
	timeout       time.Duration
	tlsConfig     *tls.Config
	tlsConfigErr  error
	authRequired  bool
	staticHeaders map[string]string
	authValidator func(ctx context.Context, headers map[string]string) error
}

// Serve starts listening for incoming HTTP requests on the configured address/endpoint.
func (h *httpServerAdapter[T]) Serve(ctx context.Context, submitFunc func(ctx context.Context, req T) (types.HTTPServerResponse, error)) error {
	if submitFunc == nil {
		return errors.New("submitFunc cannot be nil")
	}

	h.serverMu.Lock()
	defer h.serverMu.Unlock()

	if h.server != nil {
		return fmt.Errorf("server already started")
	}

	cfg := h.snapshotConfig()
	if cfg.tlsConfigErr != nil {
		return cfg.tlsConfigErr
	}
	if cfg.method == "" {
		return errors.New("method not configured")
	}
	if cfg.endpoint == "" {
		return errors.New("endpoint not configured")
	}

	atomic.StoreInt32(&h.configFrozen, 1)

	handler := h.buildHandler(cfg, submitFunc)

	h.server = &http.Server{
		Addr:         cfg.address,
		Handler:      handler,
		ReadTimeout:  cfg.timeout,
		WriteTimeout: cfg.timeout,
		TLSConfig:    cfg.tlsConfig,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	errCh := make(chan error, 1)
	go func() {
		var err error
		if cfg.tlsConfig != nil {
			h.NotifyLoggers(types.InfoLevel, "Serve: starting HTTPS server on %s %s", cfg.address, cfg.endpoint)
			err = h.server.ListenAndServeTLS("", "")
		} else {
			h.NotifyLoggers(types.InfoLevel, "Serve: starting HTTP server on %s %s", cfg.address, cfg.endpoint)
			err = h.server.ListenAndServe()
		}
		errCh <- err
	}()

	select {
	case <-ctx.Done():
		h.NotifyLoggers(types.WarnLevel, "Serve: context canceled, shutting down")
		shutdownTimeout := cfg.timeout
		if shutdownTimeout <= 0 {
			shutdownTimeout = 5 * time.Second
		}
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = h.server.Shutdown(shutdownCtx)
		return ctx.Err()
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			h.NotifyLoggers(types.ErrorLevel, "Serve: server error: %v", err)
			return err
		}
		return nil
	}
}

func (h *httpServerAdapter[T]) buildHandler(cfg serverConfig, submitFunc func(ctx context.Context, req T) (types.HTTPServerResponse, error)) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc(cfg.endpoint, func(w http.ResponseWriter, r *http.Request) {
		if cfg.method != "" && r.Method != cfg.method {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := h.authorizeRequest(r.Context(), r, cfg); err != nil {
			h.NotifyLoggers(types.WarnLevel, "Auth: request rejected: %v", err)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		parsedData, err := h.parseRequest(r)
		if err != nil {
			h.notifyHTTPServerError(err)
			http.Error(w, fmt.Sprintf("Error parsing request: %v", err), http.StatusBadRequest)
			return
		}

		resp, err := submitFunc(r.Context(), parsedData)
		if err != nil {
			h.notifyHTTPServerError(err)
			status := http.StatusInternalServerError
			var serverErr *types.HTTPServerError
			if errors.As(err, &serverErr) {
				if serverErr.StatusCode != 0 {
					status = serverErr.StatusCode
				}
				msg := serverErr.Message
				if msg == "" {
					msg = err.Error()
				}
				http.Error(w, msg, status)
				return
			}
			http.Error(w, fmt.Sprintf("Pipeline error: %v", err), status)
			return
		}

		for key, val := range cfg.headers {
			if _, exists := resp.Headers[key]; !exists {
				w.Header().Set(key, val)
			}
		}

		for key, val := range resp.Headers {
			w.Header().Set(key, val)
		}

		status := resp.StatusCode
		if status == 0 {
			status = http.StatusOK
		}
		w.WriteHeader(status)

		if len(resp.Body) > 0 {
			_, _ = w.Write(resp.Body)
		}
	})

	return mux
}
