package forwardrelay

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// IsRunning reports whether the relay is running.
func (fr *ForwardRelay[T]) IsRunning() bool {
	return atomic.LoadInt32(&fr.isRunning) == 1
}

// Start begins reading from inputs and enables submissions.
func (fr *ForwardRelay[T]) Start(ctx context.Context) error {
	if fr.Input == nil {
		return fmt.Errorf("no inputs configured")
	}

	atomic.StoreInt32(&fr.configFrozen, 1)

	for _, input := range fr.Input {
		if !input.IsStarted() {
			if err := input.Start(ctx); err != nil {
				return fmt.Errorf("failed to start input %v: %w", input.GetComponentMetadata(), err)
			}
		}
		go fr.readFromInput(input)
	}

	atomic.StoreInt32(&fr.isRunning, 1)
	fr.NotifyLoggers(types.InfoLevel, "Start: forward relay running")
	return nil
}

// Stop halts the relay and closes active streams.
func (fr *ForwardRelay[T]) Stop() {
	fr.NotifyLoggers(types.InfoLevel, "Stop: stopping forward relay")

	fr.cancel()
	fr.closeAllStreams("relay stop")

	atomic.StoreInt32(&fr.isRunning, 0)
}
