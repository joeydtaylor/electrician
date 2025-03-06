package types

import (
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// ElementError encapsulates an error associated with a specific element of type T, allowing for detailed
// error handling and tracking in processing workflows.
type ElementError[T any] struct {
	Err  error // The error encountered during the processing of the element.
	Elem T     // The element associated with the error.
}

// Element wraps queue items with additional metadata for priority management.
type Element[T any] struct {
	ID            string    // Unique identifier for the element.
	Data          T         // The actual data to be processed.
	QueuePriority int32     // Priority for processing attempts, higher is prioritized.
	CreationTime  time.Time // Timestamp when the element was first created.
	RetryCount    int32     // Number of times this element has been retried.
	Hash          string
}

// NewElement creates a new instance of Element with initial values.
func NewElement[T any](data T) *Element[T] {
	hash := utils.GenerateSha256Hash[T](data) // Handle error appropriately in production
	return &Element[T]{
		ID:            hash,
		Data:          data,
		QueuePriority: 0, // Initial priority could be set based on other factors if needed.
		CreationTime:  time.Now(),
		RetryCount:    0,
	}
}

func (e *Element[T]) IsSame(other *Element[T]) bool {
	return e.Hash == other.Hash
}

// IncrementRetryCount safely increments the retry count of the element.
func (e *Element[T]) IncrementRetryCount() {
	atomic.AddInt32(&e.RetryCount, 1)
}

// AdjustPriority adjusts the priority based on time spent in the queue and retries.
func (e *Element[T]) AdjustPriority() {
	timeInQueue := time.Since(e.CreationTime)
	// Increase priority based on time in queue, e.g., 1 point for every minute.
	atomic.AddInt32(&e.QueuePriority, int32(timeInQueue.Minutes()))
	// Further adjust priority based on retries.
	atomic.AddInt32(&e.QueuePriority, atomic.LoadInt32(&e.RetryCount))
}

// DecayPriority decreases the priority of the element if conditions are met.
func (e *Element[T]) DecayPriority() {
	if atomic.LoadInt32(&e.RetryCount) > 5 { // Example threshold for decay.
		newPriority := atomic.LoadInt32(&e.QueuePriority) - 2
		if newPriority < 0 {
			newPriority = 0 // Ensure priority doesn't go negative.
		}
		atomic.StoreInt32(&e.QueuePriority, newPriority)
	}
}
