package conduit

import (
	"context"
	"sync"
	"sync/atomic"
)

// Start starts the conduit and its managed components.
func (c *Conduit[T]) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		return nil
	}
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			atomic.StoreInt32(&c.started, 0)
			return err
		}
	}

	c.ensureDefaults()

	for _, w := range c.snapshotWires() {
		if w == nil {
			continue
		}
		_ = w.Start(c.ctx)
	}

	for _, g := range c.snapshotGenerators() {
		if g == nil {
			continue
		}
		if !g.IsStarted() {
			_ = g.Start(c.ctx)
		}
	}

	if c.InputChan != nil {
		wires := c.snapshotWires()
		if len(wires) > 0 && wires[0] != nil {
			c.completeSignal.Add(1)
			go func() {
				defer c.completeSignal.Done()
				for {
					select {
					case <-c.ctx.Done():
						return
					case elem, ok := <-c.InputChan:
						if !ok {
							return
						}
						_ = wires[0].Submit(c.ctx, elem)
					}
				}
			}()
		}
	}

	if c.NextConduit != nil && c.OutputChan != nil {
		c.completeSignal.Add(1)
		go func() {
			defer c.completeSignal.Done()
			for {
				select {
				case <-c.ctx.Done():
					return
				case elem, ok := <-c.OutputChan:
					if !ok {
						return
					}
					_ = c.NextConduit.Submit(c.ctx, elem)
				}
			}
		}()
	}

	return nil
}

// Stop cancels the conduit and waits for workers to exit.
func (c *Conduit[T]) Stop() error {
	if !atomic.CompareAndSwapInt32(&c.started, 1, 0) {
		return nil
	}

	c.terminateOnce.Do(func() {
		c.cancel()

		for _, w := range c.snapshotWires() {
			if w == nil {
				continue
			}
			_ = w.Stop()
		}

		c.completeSignal.Wait()
	})

	return nil
}

// Restart stops the conduit, reinitializes channels, and starts again.
func (c *Conduit[T]) Restart(ctx context.Context) error {
	_ = c.Stop()

	if ctx == nil {
		ctx = context.Background()
	}

	c.ctx, c.cancel = context.WithCancel(ctx)
	c.terminateOnce = sync.Once{}

	c.configLock.Lock()
	c.InputChan = make(chan T, c.MaxBufferSize)
	c.OutputChan = make(chan T, c.MaxBufferSize)
	c.configLock.Unlock()

	c.chainWires()

	if c.NextConduit != nil {
		_ = c.NextConduit.Restart(ctx)
	}

	return c.Start(c.ctx)
}
