package plug

import "sync/atomic"

// Freeze marks the plug configuration as immutable.
func (p *Plug[T]) Freeze() {
	atomic.StoreInt32(&p.frozen, 1)
}

func (p *Plug[T]) requireNotFrozen(action string) {
	if atomic.LoadInt32(&p.frozen) == 1 {
		panic("plug: " + action + " called after Freeze")
	}
}
