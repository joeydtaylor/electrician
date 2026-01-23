package wire

import "time"

func (w *Wire[T]) startCircuitBreakerTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.cbLock.Lock()
			cb := w.CircuitBreaker
			w.cbLock.Unlock()

			allowed := true
			if cb != nil {
				allowed = cb.Allow()
			}

			select {
			case w.controlChan <- allowed:
			default:
			}
		}
	}
}
