package generator

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Optional fast lane. Submitters may implement FastSubmit without changing interfaces.
type fastSubmitter[T any] interface {
	FastSubmit(context.Context, T) error
}

// buildSubmitFunc snapshots connected components once and returns a submit func that does not lock per item.
func (g *Generator[T]) buildSubmitFunc() func(context.Context, T) error {
	g.configLock.Lock()
	components := append([]types.Submitter[T](nil), g.connectedComponents...)
	g.configLock.Unlock()

	cb := g.getCircuitBreaker()

	// Cache method values (one-time). No per-item type assertions.
	submitFns := make([]func(context.Context, T) error, 0, len(components))
	for _, c := range components {
		if c == nil {
			continue
		}
		if fs, ok := c.(fastSubmitter[T]); ok {
			submitFns = append(submitFns, fs.FastSubmit)
		} else {
			submitFns = append(submitFns, c.Submit)
		}
	}

	return func(ctx context.Context, item T) error {
		if cb != nil && !cb.Allow() {
			return fmt.Errorf("operation halted, circuit breaker is open")
		}
		for _, submit := range submitFns {
			if err := submit(ctx, item); err != nil {
				if cb != nil {
					cb.RecordError()
				}
				return err
			}
		}
		return nil
	}
}

func (g *Generator[T]) startCircuitBreakerTicker(ctx context.Context) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			allowed := true
			if cb := g.getCircuitBreaker(); cb != nil {
				allowed = cb.Allow()
			}

			select {
			case g.controlChan <- allowed:
			default:
			}
		}
	}
}

func (g *Generator[T]) runControlLoop(ctx context.Context) {
	defer g.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case allowed := <-g.controlChan:
			if allowed {
				g.restartConnectors(ctx)
			}
		}
	}
}

func (g *Generator[T]) restartConnectors(ctx context.Context) {
	g.configLock.Lock()
	plugs := append([]types.Plug[T](nil), g.plugs...)
	g.configLock.Unlock()

	cb := g.getCircuitBreaker()
	submitFunc := g.buildSubmitFunc()

	for _, plug := range plugs {
		if plug == nil {
			continue
		}
		for _, pc := range plug.GetConnectors() {
			g.wg.Add(1)
			go func(pc types.Adapter[T]) {
				defer g.wg.Done()

				if cb != nil && !cb.Allow() {
					g.NotifyLoggers(types.WarnLevel, "Serve: skipped restart due to open circuit breaker")
					return
				}

				if err := pc.Serve(ctx, submitFunc); err != nil && cb != nil {
					cb.RecordError()
				}
			}(pc)
		}
	}
}

func (g *Generator[T]) getCircuitBreaker() types.CircuitBreaker[T] {
	g.cbLock.Lock()
	cb := g.CircuitBreaker
	g.cbLock.Unlock()
	return cb
}
