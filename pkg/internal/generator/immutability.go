package generator

import "sync/atomic"

// requireNotStarted panics if the generator has already been started.
func (g *Generator[T]) requireNotStarted(action string) {
	if atomic.LoadInt32(&g.started) == 1 {
		panic("generator: " + action + " called after Start")
	}
}
