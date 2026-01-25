package httpserver

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns metadata (ID, Name, Type).
func (h *httpServerAdapter[T]) GetComponentMetadata() types.ComponentMetadata {
	return h.componentMetadata
}

// SetComponentMetadata sets Name and ID.
func (h *httpServerAdapter[T]) SetComponentMetadata(name string, id string) {
	h.requireNotFrozen("SetComponentMetadata")
	h.componentMetadata.Name = name
	h.componentMetadata.ID = id
}
