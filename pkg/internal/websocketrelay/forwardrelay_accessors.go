package websocketrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetTargets returns configured targets.
func (fr *ForwardRelay[T]) GetTargets() []string {
	return append([]string(nil), fr.Targets...)
}

// GetInput returns attached inputs.
func (fr *ForwardRelay[T]) GetInput() []types.Receiver[T] {
	return append([]types.Receiver[T](nil), fr.Input...)
}
