package resister

import (
	"container/heap"
	"fmt"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (sq *Resister[T]) DecayPriorities() {
	sq.lock.Lock()
	defer sq.lock.Unlock()

	for _, el := range sq.indexMap {
		// Decrease priority if retries exceed a threshold
		if el.RetryCount > 5 {
			el.QueuePriority -= 2
			if el.QueuePriority < 0 {
				el.QueuePriority = 0
			}
			sq.NotifyLoggers(types.InfoLevel, "%s => level: INFO, result: SUCCESS, event: DecayPriorities, element: %v, queuePriority: %d, queueRetryAttempts: %d => Decayed Element Priority", sq.componentMetadata, el, el.QueuePriority, el.RetryCount)
			heap.Fix(&sq.pq, sq.pq.index(el))
		}
	}
}

func (sq *Resister[T]) Pop() *types.Element[T] {

	sq.lock.Lock()
	defer sq.lock.Unlock()
	if len(sq.pq) == 0 {
		return nil
	}
	element := heap.Pop(&sq.pq).(*types.Element[T])
	sq.notifyResisterProcessed(element.Data)
	sq.NotifyLoggers(types.InfoLevel, "%s => level: INFO, result: SUCCESS, event: Pop, element: %v, queuePriority: %d, queueRetryAttempts: %d => Popping Element", sq.componentMetadata, element, element.QueuePriority, element.RetryCount)
	delete(sq.indexMap, element.ID)

	if len(sq.pq) == 0 {
		sq.notifyResisterEmpty()
	}
	return element
}

func (sq *Resister[T]) Len() int {
	sq.lock.Lock()
	defer sq.lock.Unlock()
	return len(sq.pq)
}

func (sq *Resister[T]) Push(element *types.Element[T]) error {
	sq.lock.Lock()
	defer sq.lock.Unlock()

	if _, exists := sq.indexMap[element.ID]; exists {
		sq.notifyResisterElementRequeued(element.Data)
		heap.Fix(&sq.pq, sq.pq.index(element)) // Rebalance heap after priority update
	} else {
		sq.notifyResisterQueued(element.Data)
		sq.indexMap[element.ID] = element
		heap.Push(&sq.pq, element)
	}
	return nil
}

func (pq *PriorityQueue[T]) Len() int { return len(*pq) }

func (pq *PriorityQueue[T]) Less(i, j int) bool {
	// Higher priority items come first.
	return (*pq)[i].QueuePriority > (*pq)[j].QueuePriority
}

func (pq *PriorityQueue[T]) Swap(i, j int) {
	(*pq)[i], (*pq)[j] = (*pq)[j], (*pq)[i]
}

func (pq *PriorityQueue[T]) Push(x interface{}) {
	item := x.(*types.Element[T])
	*pq = append(*pq, item)
}

func (pq *PriorityQueue[T]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

// GetComponentMetadata returns the metadata.
func (r *Resister[T]) GetComponentMetadata() types.ComponentMetadata {
	return r.componentMetadata
}

// GetComponentMetadata returns the metadata.
func (r *Resister[T]) GetResisterQueueLen() int {
	r.notifyResisterEmpty()
	return r.Len()
}

// SetComponentMetadata sets the component metadata.
func (r *Resister[T]) SetComponentMetadata(name string, id string) {
	r.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

// ConnectLogger attaches loggers to the SurgeProtector.
func (r *Resister[T]) ConnectLogger(loggers ...types.Logger) {
	r.loggersLock.Lock()
	defer r.loggersLock.Unlock()
	r.loggers = append(r.loggers, loggers...)
}

// ConnectLogger attaches loggers to the SurgeProtector.
func (r *Resister[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	r.loggersLock.Lock()
	defer r.loggersLock.Unlock()
	r.sensors = append(r.sensors, sensor...)
}

// NotifyLoggers sends a formatted log message to all attached loggers, facilitating unified logging
// across various components of the ReceivingRelay.
func (r *Resister[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if r.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range r.loggers {
			if logger == nil {
				continue // Skip if the logger is nil.
			}
			r.loggersLock.Lock()
			defer r.loggersLock.Unlock()
			switch level {
			case types.DebugLevel:
				logger.Debug(msg)
			case types.InfoLevel:
				logger.Info(msg)
			case types.WarnLevel:
				logger.Warn(msg)
			case types.ErrorLevel:
				logger.Error(msg)
			case types.DPanicLevel:
				logger.DPanic(msg)
			case types.PanicLevel:
				logger.Panic(msg)
			case types.FatalLevel:
				logger.Fatal(msg)
			}
		}
	}
}
