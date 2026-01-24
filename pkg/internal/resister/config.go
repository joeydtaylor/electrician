package resister

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetComponentMetadata updates the component metadata.
func (r *Resister[T]) SetComponentMetadata(name string, id string) {
	r.metadataLock.Lock()
	r.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: r.componentMetadata.Type}
	r.metadataLock.Unlock()
}
