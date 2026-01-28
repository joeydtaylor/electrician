package forwardrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (fr *ForwardRelay[T]) readFromInput(input types.Receiver[T]) {
	for {
		select {
		case <-fr.ctx.Done():
			fr.logKV(types.InfoLevel, "Input read stopped",
				"event", "ReadInput",
				"result", "CANCELLED",
			)
			return
		case data, ok := <-input.GetOutputChannel():
			if !ok {
				return
			}
			if err := fr.Submit(fr.ctx, data); err != nil {
				fr.logKV(types.ErrorLevel, "Input submit failed",
					"event", "ReadInput",
					"result", "FAILURE",
					"error", err,
				)
			}
		}
	}
}
