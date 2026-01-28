package quicrelay

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/quic-go/quic-go"
)

// IsRunning reports whether the relay is running.
func (rr *ReceivingRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&rr.isRunning) == 1
}

// Start begins relay operations and starts the QUIC listener.
func (rr *ReceivingRelay[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	rr.NotifyLoggers(types.InfoLevel, "Start: starting QUIC receiving relay")
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
	rr.NotifyLoggers(types.InfoLevel, "Stop: stopping QUIC receiving relay")

	rr.cancel()
	rr.shutdownListener()
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

// Listen starts the QUIC listener on the configured address.
func (rr *ReceivingRelay[T]) Listen(listenForever bool, retryInSeconds int) error {
	atomic.StoreInt32(&rr.configFrozen, 1)
	rr.startOutputFanout()

	if retryInSeconds <= 0 {
		retryInSeconds = 2
	}

	tlsCfg, err := rr.buildServerTLSConfig()
	if err != nil {
		return err
	}

	for {
		listener, err := quic.ListenAddr(rr.Address, tlsCfg, &quic.Config{})
		if err != nil {
			rr.NotifyLoggers(types.ErrorLevel, "Listen: bind failed %s: %v", rr.Address, err)
			if !listenForever {
				return err
			}
			rr.NotifyLoggers(types.WarnLevel, "Listen: retrying in %d seconds", retryInSeconds)
			time.Sleep(time.Duration(retryInSeconds) * time.Second)
			continue
		}

		rr.setListener(listener)
		rr.NotifyLoggers(types.InfoLevel, "Listen: QUIC listener started on %s", rr.Address)

		rr.acceptLoop(listener)

		if !listenForever {
			return nil
		}
	}
}

func (rr *ReceivingRelay[T]) acceptLoop(listener *quic.Listener) {
	for {
		conn, err := listener.Accept(rr.ctx)
		if err != nil {
			return
		}
		go rr.handleConn(conn)
	}
}

func (rr *ReceivingRelay[T]) handleConn(conn quic.Connection) {
	for {
		stream, err := conn.AcceptStream(rr.ctx)
		if err != nil {
			return
		}
		go rr.handleStream(stream)
	}
}

func (rr *ReceivingRelay[T]) setListener(listener *quic.Listener) {
	rr.listenerMu.Lock()
	rr.listener = listener
	rr.listenerMu.Unlock()
}

func (rr *ReceivingRelay[T]) shutdownListener() {
	rr.listenerMu.Lock()
	listener := rr.listener
	rr.listener = nil
	rr.listenerMu.Unlock()

	if listener != nil {
		_ = listener.Close()
	}
}
