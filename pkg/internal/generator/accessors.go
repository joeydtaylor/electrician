package generator

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// GetComponentMetadata returns the generator metadata.
func (g *Generator[T]) GetComponentMetadata() types.ComponentMetadata {
	return g.componentMetadata
}

// IsStarted reports whether the generator is running.
func (g *Generator[T]) IsStarted() bool {
	return atomic.LoadInt32(&g.started) == 1
}
