package generator

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Restart stops and restarts the Generator component.
func (g *Generator[T]) Restart(ctx context.Context) error {
	// Stop the generator if it's running.
	if err := g.Stop(); err != nil {
		return fmt.Errorf("failed to stop generator during restart: %w", err)
	}

	g.notifyRestart()
	// Restart the generator with the new context.
	return g.Start(ctx)
}

func (g *Generator[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	g.sensors = append(g.sensors, sensor...)
	for _, m := range sensor {
		g.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectSensor, target: %v => Connected sensor", g.componentMetadata, m.GetComponentMetadata())
	}
}

func (g *Generator[T]) Start(ctx context.Context) error {
	if !atomic.CompareAndSwapInt32(&g.started, 0, 1) {
		return fmt.Errorf("generator already started")
	}
	g.notifyStart()

	// Bind generator context to Start() ctx.
	g.configLock.Lock()
	var cancel context.CancelFunc
	g.ctx, cancel = context.WithCancel(ctx)
	g.cancel = cancel

	// Snapshot plugs so we don't hold configLock while starting goroutines.
	plugs := append([]types.Plug[T](nil), g.plugs...)
	cb := g.CircuitBreaker
	g.configLock.Unlock()

	g.wg.Add(1)
	go g.listenToControlChan()

	// Auto-stop when all plug goroutines exit
	go func() {
		g.wg.Wait()
		_ = g.Stop()
	}()

	if cb != nil {
		go g.startCircuitBreakerTicker()
	}

	// Build once; used by all plugs/connectors.
	submitFunc := g.buildSubmitFunc()

	for _, plug := range plugs {
		if plug == nil {
			continue
		}

		// Adapter funcs
		for _, pf := range plug.GetAdapterFuncs() {
			g.wg.Add(1)
			go func(pf types.AdapterFunc[T]) {
				defer g.wg.Done()
				pf(g.ctx, submitFunc)
			}(pf)
		}

		// Connectors
		for _, pc := range plug.GetConnectors() {
			g.wg.Add(1)
			go func(pc types.Adapter[T]) {
				defer g.wg.Done()

				// Keep your existing “don’t start plugs when CB open” behavior.
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

	return nil
}

func (g *Generator[T]) IsStarted() bool {
	return (atomic.LoadInt32(&g.started)) == 1
}

func (g *Generator[T]) Stop() error {
	// It invokes the OnCancel callback for each sensor.
	g.stopOnce.Do(func() {
		if atomic.CompareAndSwapInt32(&g.started, 1, 0) {
			g.notifyStop()
			// Cancel the context to signal all plugs to stop
			if g.cancel != nil {
				g.cancel()
			}

			// Wait for all plug goroutines to finish
			for atomic.LoadInt32(&g.started) > 0 {
				/* 		fmt.Println("Waiting for all generator plugs to stop...") */
				time.Sleep(100 * time.Millisecond) // Wait a bit before checking again
			}
		}
	})
	// notifyElementProcessed notifies all sensors of the wire that an element has been processed.

	/* 	fmt.Println("All generator plugs have stopped.") */
	return nil
}

func (g *Generator[T]) ConnectCircuitBreaker(cb types.CircuitBreaker[T]) {
	g.CircuitBreaker = cb
	g.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectCircuitBreaker, target: %v => Connected circuit breaker", g.componentMetadata, cb)
}

func (g *Generator[T]) ConnectPlug(plugs ...types.Plug[T]) {
	g.configLock.Lock()
	defer g.configLock.Unlock()
	g.plugs = append(g.plugs, plugs...)
}

// GetComponentMetadata returns the metadata.
func (g *Generator[T]) GetComponentMetadata() types.ComponentMetadata {
	return g.componentMetadata
}

// SetComponentMetadata sets the component metadata.
func (g *Generator[T]) SetComponentMetadata(name string, id string) {
	g.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

func (g *Generator[T]) ConnectLogger(l ...types.Logger) {
	g.loggersLock.Lock()
	defer g.loggersLock.Unlock()
	g.loggers = append(g.loggers, l...)
}

func (g *Generator[T]) ConnectToComponent(submitters ...types.Submitter[T]) {
	g.configLock.Lock()
	defer g.configLock.Unlock()
	g.connectedComponents = append(g.connectedComponents, submitters...)
}

func (g *Generator[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if g.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range g.loggers {
			if logger == nil {
				continue
			}
			// Ensure we only acquire the lock once per logger to avoid deadlock or excessive locking overhead
			g.loggersLock.Lock()
			if logger.GetLevel() <= level {
				switch level {
				case types.DebugLevel:
					logger.Debug(msg)
				case types.InfoLevel:
					logger.Info(msg)
				case types.WarnLevel:
					logger.Warn(msg)
				case types.ErrorLevel:
					logger.Error(msg)
				case types.DPanicLevel:
					logger.DPanic(msg)
				case types.PanicLevel:
					logger.Panic(msg)
				case types.FatalLevel:
					logger.Fatal(msg)
				}
			}
			g.loggersLock.Unlock()
		}
	}
}
