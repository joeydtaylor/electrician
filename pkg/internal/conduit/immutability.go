package conduit

import "sync/atomic"

func (c *Conduit[T]) requireNotStarted(action string) {
	if atomic.LoadInt32(&c.started) == 1 {
		panic("conduit: " + action + " called after Start")
	}
}
