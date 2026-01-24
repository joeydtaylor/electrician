package resister

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns the resister metadata.
func (r *Resister[T]) GetComponentMetadata() types.ComponentMetadata {
	r.metadataLock.Lock()
	metadata := r.componentMetadata
	r.metadataLock.Unlock()
	return metadata
}

// GetResisterQueueLen returns the current queue size.
func (r *Resister[T]) GetResisterQueueLen() int {
	return r.Len()
}
