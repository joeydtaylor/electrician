package sensor

import "sync"

func snapshotCallbacks[T any](mu *sync.Mutex, callbacks []T) []T {
	mu.Lock()
	out := append([]T(nil), callbacks...)
	mu.Unlock()
	return out
}
