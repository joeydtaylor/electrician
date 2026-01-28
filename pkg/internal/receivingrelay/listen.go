package receivingrelay

import (
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
)

// Listen starts the gRPC server on the configured address.
func (rr *ReceivingRelay[T]) Listen(listenForever bool, retryInSeconds int) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	var opts []grpc.ServerOption
	if rr.TlsConfig != nil && rr.TlsConfig.UseTLS {
		rr.logKV(
			types.DebugLevel,
			"TLS enabled",
			"event", "ListenTLS",
			"result", "SUCCESS",
			"tls", rr.TlsConfig,
		)
		creds, err := rr.loadTLSCredentials(rr.TlsConfig)
		if err != nil {
			rr.logKV(
				types.ErrorLevel,
				"TLS credentials failed",
				"event", "ListenTLS",
				"result", "FAILURE",
				"error", err,
			)
			return fmt.Errorf("failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	rr.logKV(
		types.InfoLevel,
		"Binding listener",
		"event", "Listen",
		"result", "START",
		"address", rr.Address,
	)
	lis, err := net.Listen("tcp", rr.Address)
	if err != nil {
		rr.logKV(
			types.ErrorLevel,
			"Bind failed",
			"event", "Listen",
			"result", "FAILURE",
			"address", rr.Address,
			"error", err,
		)
		if !listenForever {
			return err
		}
		rr.logKV(
			types.WarnLevel,
			"Bind retrying",
			"event", "ListenRetry",
			"result", "RETRY",
			"retry_in_seconds", retryInSeconds,
		)
		time.Sleep(time.Duration(retryInSeconds) * time.Second)
		return rr.Listen(listenForever, retryInSeconds)
	}

	opts = rr.appendAuthServerOptions(opts)

	s := grpc.NewServer(opts...)
	relay.RegisterRelayServiceServer(s, rr)

	rr.logKV(
		types.InfoLevel,
		"gRPC server started",
		"event", "ListenStarted",
		"result", "SUCCESS",
		"address", rr.Address,
	)
	if !listenForever {
		if err := s.Serve(lis); err != nil {
			rr.logKV(
				types.ErrorLevel,
				"gRPC server stopped",
				"event", "ListenStopped",
				"result", "FAILURE",
				"error", err,
			)
		}
		return nil
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			rr.logKV(
				types.ErrorLevel,
				"gRPC server failed",
				"event", "ListenFailed",
				"result", "FAILURE",
				"error", err,
			)
		}
	}()

	return nil
}
