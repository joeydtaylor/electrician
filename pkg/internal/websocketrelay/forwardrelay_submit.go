package websocketrelay

import (
	"context"
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Submit sends an item to all configured targets.
func (fr *ForwardRelay[T]) Submit(ctx context.Context, item T) error {
	if len(fr.Targets) == 0 {
		return fmt.Errorf("no targets configured")
	}
	if ctx == nil {
		ctx = fr.ctx
	}

	wp, err := fr.wrapItem(ctx, item)
	if err != nil {
		return err
	}

	for _, target := range fr.Targets {
		sess, err := fr.getOrCreateSession(ctx, target)
		if err != nil {
			fr.logKV(types.ErrorLevel, "WebSocket session failed",
				"event", "Session",
				"result", "FAILURE",
				"target", target,
				"error", err,
			)
			return err
		}
		env := &relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Payload{Payload: wp}}
		select {
		case sess.sendCh <- env:
		default:
			if fr.dropOnFull {
				fr.logKV(types.WarnLevel, "WebSocket send buffer full; dropping",
					"event", "StreamSend",
					"result", "DROP",
					"target", target,
					"seq", wp.Seq,
				)
				continue
			}
			return fmt.Errorf("send buffer full")
		}
	}
	return nil
}
