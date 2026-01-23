package generator

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetComponentMetadata updates the generator metadata.
// Panics if called after Start.
func (g *Generator[T]) SetComponentMetadata(name string, id string) {
	g.requireNotStarted("SetComponentMetadata")

	g.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: g.componentMetadata.Type}
}
