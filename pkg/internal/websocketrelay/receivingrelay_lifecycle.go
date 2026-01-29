package websocketrelay

import (
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// IsRunning reports whether the relay is running.
func (rr *ReceivingRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&rr.isRunning) == 1
}

// Start begins relay operations and starts the WebSocket listener.
func (rr *ReceivingRelay[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	rr.logKV(types.InfoLevel, "WebSocket receiving relay starting",
		"event", "Start",
		"result", "SUCCESS",
	)
	for _, output := range rr.Outputs {
		if !output.IsStarted() {
			output.Start(ctx)
		}
	}

	go rr.Listen(true, 0)
	atomic.StoreInt32(&rr.isRunning, 1)
	return nil
}

// Stop halts relay operations and closes resources.
func (rr *ReceivingRelay[T]) Stop() {
	rr.logKV(types.InfoLevel, "WebSocket receiving relay stopping",
		"event", "Stop",
		"result", "SUCCESS",
	)

	rr.cancel()
	rr.shutdownServer()
	close(rr.DataCh)

	for _, output := range rr.Outputs {
		output.Stop()
	}

	atomic.StoreInt32(&rr.isRunning, 0)
}

func (rr *ReceivingRelay[T]) startOutputFanout() {
	if len(rr.Outputs) == 0 {
		return
	}
	rr.outputsOnce.Do(func() {
		go func() {
			for data := range rr.DataCh {
				for _, out := range rr.Outputs {
					_ = out.Submit(rr.ctx, data)
				}
			}
		}()
	})
}

// Listen starts the WebSocket listener on the configured address.
func (rr *ReceivingRelay[T]) Listen(listenForever bool, retryInSeconds int) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	if retryInSeconds <= 0 {
		retryInSeconds = 2
	}

	for {
		ln, err := net.Listen("tcp", rr.Address)
		if err != nil {
			rr.logKV(types.ErrorLevel, "WebSocket listen failed",
				"event", "Listen",
				"result", "FAILURE",
				"addr", rr.Address,
				"error", err,
			)
			if !listenForever {
				return err
			}
			rr.logKV(types.WarnLevel, "WebSocket listen retrying",
				"event", "ListenRetry",
				"result", "RETRY",
				"addr", rr.Address,
				"retry_seconds", retryInSeconds,
			)
			time.Sleep(time.Duration(retryInSeconds) * time.Second)
			continue
		}

		if err := rr.serveOnListener(ln); err != nil {
			if !listenForever {
				return err
			}
			time.Sleep(time.Duration(retryInSeconds) * time.Second)
			continue
		}

		if !listenForever {
			return nil
		}
	}
}

func (rr *ReceivingRelay[T]) serveOnListener(ln net.Listener) error {
	server := newWSServer(rr)
	rr.setServer(server)
	return server.serve(rr.ctx, ln)
}

func (rr *ReceivingRelay[T]) setServer(s *wsServer[T]) {
	rr.serverMu.Lock()
	rr.server = s
	rr.serverMu.Unlock()
}

func (rr *ReceivingRelay[T]) shutdownServer() {
	rr.serverMu.Lock()
	s := rr.server
	rr.server = nil
	rr.serverMu.Unlock()

	if s != nil {
		_ = s.close()
	}
}
