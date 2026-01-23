package plug

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetComponentMetadata updates the plug metadata.
// Panics if called after Freeze.
func (p *Plug[T]) SetComponentMetadata(name string, id string) {
	p.requireNotFrozen("SetComponentMetadata")

	p.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: p.componentMetadata.Type}
}
