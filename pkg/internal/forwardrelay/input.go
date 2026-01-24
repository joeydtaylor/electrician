package forwardrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (fr *ForwardRelay[T]) readFromInput(input types.Receiver[T]) {
	for {
		select {
		case <-fr.ctx.Done():
			fr.NotifyLoggers(types.InfoLevel, "readFromInput: context canceled")
			return
		case data, ok := <-input.GetOutputChannel():
			if !ok {
				return
			}
			if err := fr.Submit(fr.ctx, data); err != nil {
				fr.NotifyLoggers(types.ErrorLevel, "readFromInput: submit failed: %v", err)
			}
		}
	}
}
