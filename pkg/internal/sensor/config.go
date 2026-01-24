package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// SetComponentMetadata updates sensor metadata values.
func (s *Sensor[T]) SetComponentMetadata(name string, id string) {
	s.metadataLock.Lock()
	s.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: s.componentMetadata.Type}
	s.metadataLock.Unlock()
}

func (s *Sensor[T]) snapshotMetadata() types.ComponentMetadata {
	s.metadataLock.Lock()
	metadata := s.componentMetadata
	s.metadataLock.Unlock()
	return metadata
}
