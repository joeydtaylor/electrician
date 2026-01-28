package quicrelay

import (
	"context"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (fr *ForwardRelay[T]) readFromInput(input types.Receiver[T]) {
	for {
		select {
		case <-fr.ctx.Done():
			return
		case item, ok := <-input.GetOutputChannel():
			if !ok {
				return
			}
			_ = fr.Submit(context.Background(), item)
		}
	}
}
