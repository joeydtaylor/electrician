package generator

import (
	"context"
	"fmt"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

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
	// Assuming that you need to call Start or an equivalent method on each plug

	submitFunc := func(ctx context.Context, item T) error {
		if g.CircuitBreaker != nil && !g.CircuitBreaker.Allow() {
			return fmt.Errorf("operation halted, circuit breaker is open")
		}
		g.configLock.Lock()
		defer g.configLock.Unlock()
		for _, s := range g.connectedComponents {
			if err := s.Submit(ctx, item); err != nil {
				if g.CircuitBreaker != nil {
					g.CircuitBreaker.RecordError() // Record error to possibly trip the breaker
				}
				return err
			}
		}
		return nil
	}

	for _, plug := range g.plugs {
		pcs := plug.GetConnectors()
		for _, pc := range pcs {
			g.wg.Add(1)
			go func(pc types.Adapter[T]) {
				defer g.wg.Done()
				if (g.CircuitBreaker != nil && g.CircuitBreaker.Allow()) || g.CircuitBreaker == nil {
					err := pc.Serve(g.ctx, submitFunc)
					if err != nil {
						if g.CircuitBreaker != nil {
							g.CircuitBreaker.RecordError()
						}
					}
				} else {
					g.NotifyLoggers(types.WarnLevel, "%s => level: WARN, result: FAILURE, event: Serve => Skipped starting a plug due to open circuit", g.componentMetadata)
				}
			}(pc)
		}
	}
}
