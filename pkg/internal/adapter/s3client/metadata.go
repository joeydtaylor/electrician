package s3client

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns the adapter metadata.
func (a *S3Client[T]) GetComponentMetadata() types.ComponentMetadata { return a.componentMetadata }

// SetComponentMetadata overrides name/id while preserving the component type.
func (a *S3Client[T]) SetComponentMetadata(name, id string) {
	a.componentMetadata = types.ComponentMetadata{
		Name: name,
		ID:   id,
		Type: a.componentMetadata.Type,
	}
}

// Name reports the component identifier used in logs.
func (a *S3Client[T]) Name() string { return "S3_CLIENT" }
