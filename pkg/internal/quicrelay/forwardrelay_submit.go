package quicrelay

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
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
		wp, err = forwardrelay.WrapPayload(item, fr.PerformanceOptions, fr.SecurityOptions, fr.EncryptionKey)
		if err != nil {
			return fmt.Errorf("failed to wrap payload: %w", err)
		}

		wp.Seq = atomic.AddUint64(&fr.seq, 1)
		if fr.omitPayloadMetadata {
			wp.Metadata = nil
		}
	}

	env := &relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Payload{Payload: wp}}

	for _, address := range fr.Targets {
		sess, err := fr.getOrCreateStreamSession(ctx, address)
		if err != nil {
			fr.NotifyLoggers(types.ErrorLevel, "Submit: stream setup failed addr=%s err=%v", address, err)
			continue
		}

		if fr.dropOnFull {
			select {
			case sess.sendCh <- env:
			default:
				fr.NotifyLoggers(types.WarnLevel, "Submit: drop on full addr=%s", address)
			}
			continue
		}

		select {
		case sess.sendCh <- env:
		case <-ctx.Done():
			return ctx.Err()
		case <-fr.ctx.Done():
			return fr.ctx.Err()
		}
	}

	return nil
}
