package wire

import "sync/atomic"

// requireNotStarted panics if the wire has already been started.
func (w *Wire[T]) requireNotStarted(action string) {
	if atomic.LoadInt32(&w.started) == 1 {
		panic("wire: " + action + " called after Start")
	}
}
