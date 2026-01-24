package httpclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns the adapter metadata.
func (hp *HTTPClientAdapter[T]) GetComponentMetadata() types.ComponentMetadata {
	hp.configLock.Lock()
	metadata := hp.componentMetadata
	hp.configLock.Unlock()
	return metadata
}

// SetComponentMetadata updates the adapter metadata.
func (hp *HTTPClientAdapter[T]) SetComponentMetadata(name string, id string) {
	hp.configLock.Lock()
	hp.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: hp.componentMetadata.Type}
	hp.configLock.Unlock()
}
