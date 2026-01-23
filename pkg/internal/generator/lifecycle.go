package generator

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Start begins plug execution and enables submissions.
func (g *Generator[T]) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&g.started, 0, 1) {
		return fmt.Errorf("generator already started")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		atomic.StoreInt32(&g.started, 0)
		return err
	}

	g.stopLock.Lock()
	g.stopOnce = sync.Once{}
	g.stopLock.Unlock()
	g.notifyStart()

	g.configLock.Lock()
	g.ctx, g.cancel = context.WithCancel(ctx)
	runCtx := g.ctx
	plugs := append([]types.Plug[T](nil), g.plugs...)
	g.configLock.Unlock()

	cb := g.getCircuitBreaker()

	for _, plug := range plugs {
		if plug == nil {
			continue
		}
		if freezer, ok := plug.(interface{ Freeze() }); ok {
			freezer.Freeze()
		}
	}

	g.wg.Add(1)
	go g.runControlLoop(runCtx)

	go func() {
		g.wg.Wait()
		_ = g.Stop()
	}()

	if cb != nil {
		go g.startCircuitBreakerTicker(runCtx)
	}

	submitFunc := g.buildSubmitFunc()

	for _, plug := range plugs {
		if plug == nil {
			continue
		}

		for _, pf := range plug.GetAdapterFuncs() {
			g.wg.Add(1)
			go func(pf types.AdapterFunc[T]) {
				defer g.wg.Done()
				pf(runCtx, submitFunc)
			}(pf)
		}

		for _, pc := range plug.GetConnectors() {
			g.wg.Add(1)
			go func(pc types.Adapter[T]) {
				defer g.wg.Done()

				if cb != nil && !cb.Allow() {
					g.NotifyLoggers(types.WarnLevel, "Serve: skipped start due to open circuit breaker")
					return
				}

				if err := pc.Serve(runCtx, submitFunc); err != nil && cb != nil {
					cb.RecordError()
				}
			}(pc)
		}
	}

	return nil
}

// Stop cancels the generator context and waits for plugs to exit.
func (g *Generator[T]) Stop() error {
	g.stopLock.Lock()
	defer g.stopLock.Unlock()

	g.stopOnce.Do(func() {
		if atomic.CompareAndSwapInt32(&g.started, 1, 0) {
			g.notifyStop()
			g.configLock.Lock()
			cancel := g.cancel
			g.configLock.Unlock()
			if cancel != nil {
				cancel()
			}
			g.wg.Wait()
		}
	})
	return nil
}

// Restart stops the generator and starts it with a new context.
func (g *Generator[T]) Restart(ctx context.Context) error {
	if err := g.Stop(); err != nil {
		return fmt.Errorf("failed to stop generator during restart: %w", err)
	}

	g.notifyRestart()
	return g.Start(ctx)
}
