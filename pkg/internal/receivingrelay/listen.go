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
		rr.NotifyLoggers(types.DebugLevel, "Listen: TLS enabled, loading credentials")
		creds, err := rr.loadTLSCredentials(rr.TlsConfig)
		if err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "Listen: load TLS credentials failed: %v", err)
			return fmt.Errorf("failed to load TLS credentials: %v", err)
		}
		opts = append(opts, grpc.Creds(creds))
	}

	rr.NotifyLoggers(types.InfoLevel, "Listen: binding %s", rr.Address)
	lis, err := net.Listen("tcp", rr.Address)
	if err != nil {
		rr.NotifyLoggers(types.ErrorLevel, "Listen: bind failed %s: %v", rr.Address, err)
		if !listenForever {
			return err
		}
		rr.NotifyLoggers(types.WarnLevel, "Listen: retrying in %d seconds", retryInSeconds)
		time.Sleep(time.Duration(retryInSeconds) * time.Second)
		return rr.Listen(listenForever, retryInSeconds)
	}

	opts = rr.appendAuthServerOptions(opts)

	s := grpc.NewServer(opts...)
	relay.RegisterRelayServiceServer(s, rr)

	rr.NotifyLoggers(types.InfoLevel, "Listen: gRPC server started")
	if !listenForever {
		if err := s.Serve(lis); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "Listen: server stopped: %v", err)
		}
		return nil
	}

	go func() {
		if err := s.Serve(lis); err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "Listen: server failed: %v", err)
		}
	}()

	return nil
}
