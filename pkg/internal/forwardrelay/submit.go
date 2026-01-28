package forwardrelay

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Submit wraps and forwards an item to each configured target.
func (fr *ForwardRelay[T]) Submit(ctx context.Context, item T) error {
	var (
		wp  *relay.WrappedPayload
		err error
	)
	if fr.passthrough {
		wp, err = fr.asPassthroughPayload(item)
		if err != nil {
			return fmt.Errorf("failed to passthrough payload: %w", err)
		}
	} else {
		wp, err = WrapPayload(item, fr.PerformanceOptions, fr.SecurityOptions, fr.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to wrap payload: %w", err)
		}

		wp.Seq = atomic.AddUint64(&fr.seq, 1)
		wp.Metadata = nil
	}

	env := &relay.RelayEnvelope{
		Msg: &relay.RelayEnvelope_Payload{Payload: wp},
	}

	for _, address := range fr.Targets {
		sess, err := fr.getOrCreateStreamSession(ctx, address)
		if err != nil {
			fr.logKV(types.ErrorLevel, "Submit: stream setup failed",
				"event", "Submit",
				"result", "FAILURE",
				"target", address,
				"error", err,
			)
			continue
		}

		select {
		case sess.sendCh <- env:
			fr.logKV(types.DebugLevel, "Submit: enqueued",
				"event", "Submit",
				"result", "ENQUEUED",
				"target", address,
				"id", wp.GetId(),
				"seq", wp.GetSeq(),
				"payload_type", wp.GetPayloadType(),
				"payload_encoding", wp.GetPayloadEncoding(),
				"payload_bytes", len(wp.GetPayload()),
			)
		case <-ctx.Done():
			return ctx.Err()
		case <-fr.ctx.Done():
			return fr.ctx.Err()
		}
	}

	return nil
}
