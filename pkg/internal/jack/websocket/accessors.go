package websocket

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns component metadata (ID, Name, Type).
func (s *serverAdapter[T]) GetComponentMetadata() types.ComponentMetadata {
	return s.componentMetadata
}

// SetComponentMetadata sets Name and ID for the component.
func (s *serverAdapter[T]) SetComponentMetadata(name string, id string) {
	s.requireNotFrozen("SetComponentMetadata")
	s.componentMetadata.Name = name
	s.componentMetadata.ID = id
}
