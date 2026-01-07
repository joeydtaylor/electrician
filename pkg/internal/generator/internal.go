package generator

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Optional fast lane. Wire can implement FastSubmit without changing any existing interfaces.
type fastSubmitter[T any] interface {
	FastSubmit(context.Context, T) error
}

// buildSubmitFunc snapshots connected components once and returns a submit func that does not lock per item.
func (g *Generator[T]) buildSubmitFunc() func(context.Context, T) error {
	g.configLock.Lock()
	components := append([]types.Submitter[T](nil), g.connectedComponents...)
	cb := g.CircuitBreaker
	g.configLock.Unlock()

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

func (g *Generator[T]) startCircuitBreakerTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-g.ctx.Done():
			return
		case <-ticker.C:
			g.cbLock.Lock() // Lock before checking CircuitBreaker
			allowed := true
			if g.CircuitBreaker != nil {
				allowed = g.CircuitBreaker.Allow()
			}
			g.cbLock.Unlock() // Unlock after checking

			select {
			case g.controlChan <- allowed:
			default:
			}
		}
	}
}

func (g *Generator[T]) listenToControlChan() {
	for {
		select {
		case <-g.ctx.Done():
			g.wg.Done()
			return
		case allowed := <-g.controlChan:
			if allowed {
				g.restartOperations() // Define this to manage restarting operations
			}
		}
	}
}

// notifyElementProcessed notifies all sensors of the wire that an element has been processed.
// It invokes the OnCancel callback for each sensor.
func (g *Generator[T]) notifyRestart() {
	for _, sensor := range g.sensors {
		sensor.InvokeOnRestart(g.componentMetadata)
		g.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: notifyRestart, target_component: %s => Invoked OnRestart for sensor", g.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

// notifyElementProcessed notifies all sensors of the wire that an element has been processed.
// It invokes the OnCancel callback for each sensor.
func (g *Generator[T]) notifyStop() {
	for _, sensor := range g.sensors {
		sensor.InvokeOnStop(g.componentMetadata)
		g.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: notifyRestart, target_component: %s => Invoked OnRestart for sensor", g.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

// notifyElementProcessed notifies all sensors of the wire that an element has been processed.
// It invokes the OnCancel callback for each sensor.
func (g *Generator[T]) notifyStart() {
	for _, sensor := range g.sensors {
		sensor.InvokeOnStart(g.componentMetadata)
		g.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: notifyRestart, target_component: %s => Invoked OnRestart for sensor", g.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

func (g *Generator[T]) restartOperations() {
	// Snapshot plugs and cb
	g.configLock.Lock()
	plugs := append([]types.Plug[T](nil), g.plugs...)
	cb := g.CircuitBreaker
	g.configLock.Unlock()

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
					g.NotifyLoggers(types.WarnLevel, "%s => level: WARN, result: FAILURE, event: Serve => Skipped starting a plug due to open circuit", g.componentMetadata)
					return
				}

				if err := pc.Serve(g.ctx, submitFunc); err != nil && cb != nil {
					cb.RecordError()
				}
			}(pc)
		}
	}
}
