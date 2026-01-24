package kafkaclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// GetComponentMetadata returns the adapter metadata.
func (a *KafkaClient[T]) GetComponentMetadata() types.ComponentMetadata { return a.componentMetadata }

// SetComponentMetadata overrides name/id while preserving the component type.
func (a *KafkaClient[T]) SetComponentMetadata(name, id string) {
	a.componentMetadata = types.ComponentMetadata{
		Name: name,
		ID:   id,
		Type: a.componentMetadata.Type,
	}
}

// Name reports the component identifier used in logs.
func (a *KafkaClient[T]) Name() string { return "KAFKA_CLIENT" }
