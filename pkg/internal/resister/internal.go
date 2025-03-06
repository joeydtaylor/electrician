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

func (r *Resister[T]) notifyResisterQueued(elem T) {

	for _, sensor := range r.sensors {
		r.sensorLock.Lock()
		defer r.sensorLock.Unlock()
		sensor.InvokeOnResisterQueued(r.componentMetadata, elem)

		r.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyResisterProcessed, target_component: %s => Invoked InvokeOnResisterProcessed for sensor", r.componentMetadata, sensor.GetComponentMetadata())
	}
}

func (r *Resister[T]) notifyResisterElementRequeued(elem T) {

	for _, sensor := range r.sensors {
		r.sensorLock.Lock()
		defer r.sensorLock.Unlock()
		sensor.InvokeOnResisterRequeued(r.componentMetadata, elem)

		r.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyResisterElementRequeued, target_component: %s => Invoked InvokeOnResisterRequeue for sensor", r.componentMetadata, sensor.GetComponentMetadata())
	}
}

func (r *Resister[T]) notifyResisterProcessed(elem T) {
	for _, sensor := range r.sensors {
		r.sensorLock.Lock()
		defer r.sensorLock.Unlock()
		sensor.InvokeOnResisterDequeued(r.componentMetadata, elem)

		r.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyResisterProcessed, target_component: %s => Invoked InvokeOnResisterProcessed for sensor", r.componentMetadata, sensor.GetComponentMetadata())
	}
}

func (r *Resister[T]) notifyResisterEmpty() {
	for _, sensor := range r.sensors {
		r.sensorLock.Lock()
		defer r.sensorLock.Unlock()
		sensor.InvokeOnResisterEmpty(r.componentMetadata)

		r.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyResisterEmpty, target_component: %s => Invoked InvokeOnResisterEmpty for sensor", r.componentMetadata, sensor.GetComponentMetadata())
	}
}
