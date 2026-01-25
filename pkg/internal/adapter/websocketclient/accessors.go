package websocketclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns metadata (ID, Name, Type).
func (c *WebSocketClientAdapter[T]) GetComponentMetadata() types.ComponentMetadata {
	return c.componentMetadata
}

// SetComponentMetadata sets Name and ID.
func (c *WebSocketClientAdapter[T]) SetComponentMetadata(name string, id string) {
	c.componentMetadata.Name = name
	c.componentMetadata.ID = id
}
