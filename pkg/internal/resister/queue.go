package resister

import (
	"container/heap"
	"errors"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type priorityQueue[T any] []*types.Element[T]

func (pq priorityQueue[T]) Len() int { return len(pq) }

func (pq priorityQueue[T]) Less(i, j int) bool {
	return pq[i].QueuePriority > pq[j].QueuePriority
}

func (pq priorityQueue[T]) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue[T]) Push(x interface{}) {
	item := x.(*types.Element[T])
	*pq = append(*pq, item)
}

func (pq *priorityQueue[T]) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

func (pq priorityQueue[T]) indexOf(id string) int {
	for i, el := range pq {
		if el != nil && el.ID == id {
			return i
		}
	}
	return -1
}

// Len returns the number of elements currently queued.
func (r *Resister[T]) Len() int {
	r.queueLock.Lock()
	n := len(r.queue)
	r.queueLock.Unlock()
	return n
}

// Push enqueues an element, requeueing if the id already exists.
func (r *Resister[T]) Push(element *types.Element[T]) error {
	if element == nil {
		return errors.New("resister: nil element")
	}

	var (
		requeued bool
		data     T
	)

	r.queueLock.Lock()
	if existing, exists := r.indexByID[element.ID]; exists {
		requeued = true
		if existing != element {
			existing.Data = element.Data
			existing.QueuePriority = element.QueuePriority
			existing.CreationTime = element.CreationTime
			existing.RetryCount = element.RetryCount
		}
		if idx := r.queue.indexOf(element.ID); idx >= 0 {
			heap.Fix(&r.queue, idx)
		} else {
			heap.Push(&r.queue, existing)
		}
		data = existing.Data
	} else {
		r.indexByID[element.ID] = element
		heap.Push(&r.queue, element)
		data = element.Data
	}
	r.queueLock.Unlock()

	if requeued {
		r.notifyResisterRequeued(data)
	} else {
		r.notifyResisterQueued(data)
	}

	r.NotifyLoggers(
		types.DebugLevel,
		"Resister queued element",
		"component", r.snapshotMetadata(),
		"elementID", element.ID,
		"priority", atomic.LoadInt32(&element.QueuePriority),
		"retryCount", atomic.LoadInt32(&element.RetryCount),
		"requeued", requeued,
	)

	return nil
}

// Pop removes and returns the highest priority element.
func (r *Resister[T]) Pop() *types.Element[T] {
	r.queueLock.Lock()
	if len(r.queue) == 0 {
		r.queueLock.Unlock()
		return nil
	}

	element := heap.Pop(&r.queue).(*types.Element[T])
	delete(r.indexByID, element.ID)
	isEmpty := len(r.queue) == 0
	r.queueLock.Unlock()

	r.notifyResisterDequeued(element.Data)
	r.NotifyLoggers(
		types.DebugLevel,
		"Resister dequeued element",
		"component", r.snapshotMetadata(),
		"elementID", element.ID,
		"priority", atomic.LoadInt32(&element.QueuePriority),
		"retryCount", atomic.LoadInt32(&element.RetryCount),
	)

	if isEmpty {
		r.notifyResisterEmpty()
	}

	return element
}

// DecayPriorities reduces element priorities based on retry policy.
func (r *Resister[T]) DecayPriorities() {
	r.queueLock.Lock()
	changed := 0
	for _, element := range r.queue {
		if element == nil {
			continue
		}
		before := atomic.LoadInt32(&element.QueuePriority)
		element.DecayPriority()
		after := atomic.LoadInt32(&element.QueuePriority)
		if before != after {
			changed++
		}
	}
	if changed != 0 {
		heap.Init(&r.queue)
	}
	r.queueLock.Unlock()

	if changed != 0 {
		r.NotifyLoggers(
			types.DebugLevel,
			"Resister decayed priorities",
			"component", r.snapshotMetadata(),
			"elements", changed,
		)
	}
}
