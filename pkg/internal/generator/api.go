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
	// Ensure that the generator can only be started once
	if !atomic.CompareAndSwapInt32(&g.started, 0, 1) {
		return fmt.Errorf("generator already started")
	}
	g.notifyStart()

	// Safely initialize context and cancel function
	g.configLock.Lock()
	var cancel context.CancelFunc
	g.ctx, cancel = context.WithCancel(ctx)
	g.cancel = cancel
	g.configLock.Unlock()

	g.wg.Add(1)
	go g.listenToControlChan()

	// Start listening to the control channel
	go func() {
		g.wg.Wait()

		g.Stop()
	}()

	// Optionally start the circuit breaker ticker
	if g.CircuitBreaker != nil {
		go g.startCircuitBreakerTicker()
	}

	// Define a submit function that will be used by all plugs and components
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

	// Start all configured plugs
	g.configLock.Lock()
	for _, plug := range g.plugs {
		// Start adapter functions for each plug
		pfs := plug.GetAdapterFuncs()
		for _, pf := range pfs {
			g.wg.Add(1)
			go func(pf types.AdapterFunc[T]) {
				defer g.wg.Done()
				pf(g.ctx, func(ctx context.Context, item T) error {
					return submitFunc(ctx, item)
				})
			}(pf)
		}

		// Start connectors for each plug
		pcs := plug.GetConnectors()
		for _, pc := range pcs {
			g.wg.Add(1)
			go func(pc types.Adapter[T]) {
				defer g.wg.Done()
				if (g.CircuitBreaker != nil && g.CircuitBreaker.Allow()) || g.CircuitBreaker == nil {
					err := pc.Serve(g.ctx, submitFunc)
					if err != nil && g.CircuitBreaker != nil {
						g.CircuitBreaker.RecordError()
					}
				} else {
					g.NotifyLoggers(types.WarnLevel, "%s => level: WARN, result: FAILURE, event: Serve => Skipped starting a plug due to open circuit", g.componentMetadata)
				}
			}(pc)
		}
	}
	g.configLock.Unlock()

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
