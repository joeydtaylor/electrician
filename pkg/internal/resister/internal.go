package resister

import (
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (pq *PriorityQueue[T]) index(element *types.Element[T]) int {
	for i, el := range *pq {
		if el.ID == element.ID {
			return i
		}
	}
	return -1
}

func (r *Resister[T]) snapshotSensors() []types.Sensor[T] {
	r.sensorLock.Lock()
	defer r.sensorLock.Unlock()
	if len(r.sensors) == 0 {
		return nil
	}
	out := make([]types.Sensor[T], len(r.sensors))
	copy(out, r.sensors)
	return out
}

func (r *Resister[T]) notifyResisterQueued(elem T) {
	sensors := r.snapshotSensors()
	for _, sensor := range sensors {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnResisterQueued(r.componentMetadata, elem)
		r.NotifyLoggers(
			types.DebugLevel,
			"%s => level: DEBUG, result: SUCCESS, event: notifyResisterQueued, target_component: %s => Invoked InvokeOnResisterQueued for sensor",
			r.componentMetadata,
			sensor.GetComponentMetadata(),
		)
	}
}

func (r *Resister[T]) notifyResisterElementRequeued(elem T) {
	sensors := r.snapshotSensors()
	for _, sensor := range sensors {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnResisterRequeued(r.componentMetadata, elem)
		r.NotifyLoggers(
			types.DebugLevel,
			"%s => level: DEBUG, result: SUCCESS, event: notifyResisterElementRequeued, target_component: %s => Invoked InvokeOnResisterRequeued for sensor",
			r.componentMetadata,
			sensor.GetComponentMetadata(),
		)
	}
}

func (r *Resister[T]) notifyResisterProcessed(elem T) {
	sensors := r.snapshotSensors()
	for _, sensor := range sensors {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnResisterDequeued(r.componentMetadata, elem)
		r.NotifyLoggers(
			types.DebugLevel,
			"%s => level: DEBUG, result: SUCCESS, event: notifyResisterProcessed, target_component: %s => Invoked InvokeOnResisterDequeued for sensor",
			r.componentMetadata,
			sensor.GetComponentMetadata(),
		)
	}
}

func (r *Resister[T]) notifyResisterEmpty() {
	sensors := r.snapshotSensors()
	for _, sensor := range sensors {
		if sensor == nil {
			continue
		}
		sensor.InvokeOnResisterEmpty(r.componentMetadata)
		r.NotifyLoggers(
			types.DebugLevel,
			"%s => level: DEBUG, result: SUCCESS, event: notifyResisterEmpty, target_component: %s => Invoked InvokeOnResisterEmpty for sensor",
			r.componentMetadata,
			sensor.GetComponentMetadata(),
		)
	}
}
