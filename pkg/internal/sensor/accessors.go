package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns the sensor metadata.
func (s *Sensor[T]) GetComponentMetadata() types.ComponentMetadata {
	s.metadataLock.Lock()
	metadata := s.componentMetadata
	s.metadataLock.Unlock()
	return metadata
}

// GetMeters returns a copy of configured meters.
func (s *Sensor[T]) GetMeters() []types.Meter[T] {
	return s.snapshotMeters()
}
