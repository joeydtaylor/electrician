//go:build webtransport

package webtransportrelay

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (fr *ForwardRelay[T]) GetTargets() []string {
	return append([]string(nil), fr.Targets...)
}

func (fr *ForwardRelay[T]) GetInput() []types.Receiver[T] {
	return append([]types.Receiver[T](nil), fr.Input...)
}
