package receivingrelay

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
)

// ListenGRPCWeb starts a gRPC-Web server on the configured address.
func (rr *ReceivingRelay[T]) ListenGRPCWeb(listenForever bool, retryInSeconds int) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	cfg := rr.snapshotGRPCWebConfig()
	grpcServer := rr.newGRPCWebServer()
	handler := rr.buildGRPCWebHandler(grpcServer, cfg)

	rr.NotifyLoggers(types.InfoLevel, "ListenGRPCWeb: binding %s", rr.Address)
	lis, err := net.Listen("tcp", rr.Address)
	if err != nil {
		rr.NotifyLoggers(types.ErrorLevel, "ListenGRPCWeb: bind failed %s: %v", rr.Address, err)
		if !listenForever {
			return err
		}
		rr.NotifyLoggers(types.WarnLevel, "ListenGRPCWeb: retrying in %d seconds", retryInSeconds)
		time.Sleep(time.Duration(retryInSeconds) * time.Second)
		return rr.ListenGRPCWeb(listenForever, retryInSeconds)
	}

	srv := &http.Server{
		Addr:    rr.Address,
		Handler: handler,
	}
	rr.setGRPCWebServer(srv)

	serve := func(l net.Listener) error {
		if err := srv.Serve(l); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
			rr.NotifyLoggers(types.ErrorLevel, "ListenGRPCWeb: server stopped: %v", err)
			return err
		}
		return nil
	}

	if rr.TlsConfig != nil && rr.TlsConfig.UseTLS {
		tlsCfg, err := buildGRPCWebTLSConfig(rr.TlsConfig)
		if err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "ListenGRPCWeb: TLS config error: %v", err)
			return err
		}
		srv.TLSConfig = tlsCfg
		lis = tls.NewListener(lis, tlsCfg)
		rr.NotifyLoggers(types.InfoLevel, "ListenGRPCWeb: TLS enabled")
	}

	if !listenForever {
		return serve(lis)
	}

	go func() {
		_ = serve(lis)
	}()

	return nil
}

// StartGRPCWeb begins relay operations and starts the gRPC-Web server.
func (rr *ReceivingRelay[T]) StartGRPCWeb(ctx context.Context) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	rr.NotifyLoggers(types.InfoLevel, "StartGRPCWeb: starting receiving relay")
	for _, output := range rr.Outputs {
		if !output.IsStarted() {
			output.Start(ctx)
		}
	}

	go rr.ListenGRPCWeb(true, 0)
	atomic.StoreInt32(&rr.isRunning, 1)
	return nil
}

func (rr *ReceivingRelay[T]) newGRPCWebServer() *grpc.Server {
	opts := rr.appendAuthServerOptions(nil)
	srv := grpc.NewServer(opts...)
	relay.RegisterRelayServiceServer(srv, rr)
	return srv
}

func (rr *ReceivingRelay[T]) buildGRPCWebHandler(srv *grpc.Server, cfg types.GRPCWebConfig) http.Handler {
	originAllowed := func(origin string) bool {
		return grpcWebOriginAllowed(origin, cfg)
	}

	opts := []grpcweb.Option{
		grpcweb.WithOriginFunc(originAllowed),
	}
	if len(cfg.AllowedHeaders) > 0 {
		opts = append(opts, grpcweb.WithAllowedRequestHeaders(cfg.AllowedHeaders))
	}
	if !cfg.DisableWebsockets {
		opts = append(opts,
			grpcweb.WithWebsockets(true),
			grpcweb.WithWebsocketOriginFunc(func(r *http.Request) bool {
				return originAllowed(r.Header.Get("Origin"))
			}),
		)
	}

	wrapped := grpcweb.WrapServer(srv, opts...)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case wrapped.IsGrpcWebRequest(r),
			wrapped.IsAcceptableGrpcCorsRequest(r),
			wrapped.IsGrpcWebSocketRequest(r):
			wrapped.ServeHTTP(w, r)
			return
		case isGRPCRequest(r):
			srv.ServeHTTP(w, r)
			return
		default:
			http.NotFound(w, r)
		}
	})
}

func grpcWebOriginAllowed(origin string, cfg types.GRPCWebConfig) bool {
	if origin == "" {
		return true
	}
	if cfg.AllowAllOrigins {
		return true
	}
	for _, allowed := range cfg.AllowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if allowed == "" {
			continue
		}
		if allowed == "*" {
			return true
		}
		if strings.EqualFold(allowed, origin) {
			return true
		}
	}
	return false
}

func isGRPCRequest(r *http.Request) bool {
	if r.ProtoMajor != 2 {
		return false
	}
	ct := r.Header.Get("Content-Type")
	return strings.HasPrefix(ct, "application/grpc")
}

func buildGRPCWebTLSConfig(cfg *types.TLSConfig) (*tls.Config, error) {
	if cfg == nil || !cfg.UseTLS {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate/key: %w", err)
	}

	minTLS := cfg.MinTLSVersion
	if minTLS == 0 {
		minTLS = tls.VersionTLS12
	}
	maxTLS := cfg.MaxTLSVersion
	if maxTLS == 0 {
		maxTLS = tls.VersionTLS13
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   minTLS,
		MaxVersion:   maxTLS,
		NextProtos:   []string{"h2", "http/1.1"},
	}

	if cfg.CAFile != "" {
		caData, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if ok := caPool.AppendCertsFromPEM(caData); !ok {
			return nil, fmt.Errorf("failed to parse CA certificate(s) in %s", cfg.CAFile)
		}
		tlsConf.ClientCAs = caPool
	}

	return tlsConf, nil
}

func (rr *ReceivingRelay[T]) setGRPCWebServer(srv *http.Server) {
	rr.grpcWebServerMu.Lock()
	rr.grpcWebServer = srv
	rr.grpcWebServerMu.Unlock()
}

func (rr *ReceivingRelay[T]) shutdownGRPCWebServer() {
	rr.grpcWebServerMu.Lock()
	srv := rr.grpcWebServer
	rr.grpcWebServer = nil
	rr.grpcWebServerMu.Unlock()
	if srv == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, net.ErrClosed) {
		rr.NotifyLoggers(types.WarnLevel, "Stop: grpc-web shutdown error: %v", err)
	}
}
