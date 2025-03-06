package sensor

import (
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (m *Sensor[T]) ConnectLogger(logger ...types.Logger) {
	m.loggers = append(m.loggers, logger...)
}

func (m *Sensor[T]) ConnectMeter(meter ...types.Meter[T]) {
	m.meters = append(m.meters, meter...)
}

func (m *Sensor[T]) GetComponentMetadata() types.ComponentMetadata {
	return m.componentMetadata
}

func (m *Sensor[T]) GetMeters() []types.Meter[T] {
	return m.meters
}

func (m *Sensor[T]) SetComponentMetadata(name string, id string) {
	oldMetadata := m.componentMetadata
	m.componentMetadata.Name = name
	m.componentMetadata.ID = id
	m.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: SetComponentMetadata, old: %v, new: %v => Component metadata updated", oldMetadata, m.componentMetadata)
}
