package receivingrelay

import (
	"context"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// IsRunning reports whether the relay is running.
func (rr *ReceivingRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&rr.isRunning) == 1
}

// Start begins relay operations and starts the gRPC server.
func (rr *ReceivingRelay[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	rr.NotifyLoggers(types.InfoLevel, "Start: starting receiving relay")
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
	rr.NotifyLoggers(types.InfoLevel, "Stop: stopping receiving relay")

	rr.cancel()
	rr.shutdownGRPCWebServer()
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
