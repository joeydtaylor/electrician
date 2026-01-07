// Package conduit facilitates the creation of robust data processing pipelines by providing a conduit component.
// This conduit acts as a bridge between various components in a system, allowing for efficient data transfer and processing.
// It manages wires, circuit breakers, generators, loggers, and sensors to construct flexible and customizable pipelines.
//
// The api.go file within this package implements essential methods and functionalities for the Conduit component:
// - Connection Methods: ConnectCircuitBreaker connects a circuit breaker to control data flow.
//   ConnectConduit establishes connections to the next conduit in the processing chain.
//   ConnectPlug links generator functions to the conduit for data generation.
//   ConnectLogger attaches loggers to the conduit for event logging.
//   ConnectSensor adds a sensor to measure and monitor data flow.
//   ConnectWire appends new wires to the conduit's processing chain.
// - Metadata and State Methods: GetComponentMetadata retrieves conduit metadata.
//   GetCircuitBreaker returns the associated circuit breaker.
//   GetInputChannel and GetOutputChannel provide access to the conduit's input and output channels.
//   IsStarted checks if the conduit has been started.
// - Data Processing Methods: Load retrieves the output buffer after processing.
//   LoadAsJSONArray collects output into a JSON array.
//   Submit submits an element for processing within the conduit.
//   Terminate stops the conduit and its managed wires.
//
// This comprehensive set of methods enables developers to build scalable, resilient, and observable data processing pipelines
// with fine-grained control over data flow and event handling.

package conduit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (c *Conduit[T]) ConnectSurgeProtector(protector types.SurgeProtector[T]) {
	c.surgeProtectorLock.Lock()
	defer c.surgeProtectorLock.Unlock()
	c.surgeProtector = protector
	if len(c.wires) != 0 {
		for _, w := range c.wires {
			c.surgeProtector.ConnectComponent(w)
			w.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectSurgeProtector, target: %s => Surge protector connected", w.GetComponentMetadata(), protector.GetComponentMetadata())
		}
	}
}

// ConnectCircuitBreaker attaches a Circuit Breaker to the Conduit.
// This function connects a circuit breaker to the conduit, allowing it to control the flow of data.
// Parameters:
//   - cb: The circuit breaker to be connected.
func (c *Conduit[T]) ConnectCircuitBreaker(cb types.CircuitBreaker[T]) {
	c.cbLock.Lock()
	defer c.cbLock.Unlock()
	c.CircuitBreaker = cb
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectCircuitBreaker, target: %v => Connected circuit breaker", c.componentMetadata, cb.GetComponentMetadata())
}

// ConnectConduit establishes the next conduit in the chain.
// This function connects the next conduit in the processing chain.
// Parameters:
//   - next: The next conduit to be connected.
func (c *Conduit[T]) ConnectConduit(next types.Conduit[T]) {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.NextConduit = next
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectConduit, target: %v => Connected conduit", c.componentMetadata, next.GetComponentMetadata())
}

// ConnectPlug sets the generator function of the conduit.
// This function connects one or more generator functions to the conduit component,
// enabling it to generate data and feed it into the processing pipeline.
// Parameters:
//   - generator: The generator function(s) to be connected.
func (c *Conduit[T]) ConnectGenerator(generator ...types.Generator[T]) {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.generators = append(c.generators, generator...)
	for _, g := range generator {
		c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectPlug, target: %v => Connected generator", c.componentMetadata, g)
	}
}

func (c *Conduit[T]) GetGenerators() []types.Generator[T] {

	return c.generators
}

// ConnectLogger adds a logger to the conduit.
// This function connects a logger to the conduit component, allowing it to
// log relevant events and information during data processing.
// Parameters:
//   - logger: The logger to be connected.
func (c *Conduit[T]) ConnectLogger(logger ...types.Logger) {
	c.loggersLock.Lock()
	defer c.loggersLock.Unlock()
	c.loggers = append(c.loggers, logger...)
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectLogger, target: %v => Connected Logger", c.componentMetadata, &logger)
}

// ConnectSensor adds a sensor to the wire.
// This function connects a sensor to the wire component, enabling it to
// measure and monitor the flow of data through the processing pipeline.
// Parameters:
//   - sensor: The sensor to be connected.
func (c *Conduit[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.sensors = append(c.sensors, sensor...)
	for _, m := range sensor {
		c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectSensor, target: %v => Connected sensor", c.componentMetadata, m.GetComponentMetadata())
	}
}

// ConnectWire appends new Wires to the Conduit's processing chain.
// This function connects new wires to the conduit, establishing the processing chain.
// Parameters:
//   - newWires: The new wires to be connected.
func (c *Conduit[T]) ConnectWire(newWires ...types.Wire[T]) {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	for _, newWire := range newWires {

		c.wires = append(c.wires, newWire)
		c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: ConnectWire, target: %v => Connected wire", c.componentMetadata, newWire.GetComponentMetadata())

		/* 		// Connect the new wire to the previous wire in the list
		   		if len(c.wires) > 1 {
		   			prevWire := c.wires[len(c.wires)-2]
		   			newWire.SetInputChannel(prevWire.GetOutputChannel())
		   		}

		   		// Special handling for the first wire if there's a generator
		   		if len(c.wires) == 1 && c.generators != nil {
		   			for _, g := range c.generators {
		   				g.Start(c.ctx)
		   			}
		   		} */
	}

	// Connect the Conduit output channel to the last wire's output if there's no subsequent conduit
	if c.NextConduit == nil && len(c.wires) > 0 {
		lastWire := c.wires[len(c.wires)-1]
		c.OutputChan = lastWire.GetOutputChannel()
	}
}

// GetComponentMetadata returns the metadata of the conduit.
// This function retrieves the metadata associated with the conduit component.
// Returns:
//   - types.ComponentMetadata: The metadata of the conduit.
func (c *Conduit[T]) GetComponentMetadata() types.ComponentMetadata {
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: GetComponentMetadata, return: %s, => GetComponentMetadata called", c.componentMetadata, c.componentMetadata)
	return c.componentMetadata
}

// GetCircuitBreaker returns the circuit breaker of the conduit.
// This function retrieves the circuit breaker associated with the conduit.
// Returns:
//   - types.CircuitBreaker[T]: The circuit breaker of the conduit.
func (c *Conduit[T]) GetCircuitBreaker() types.CircuitBreaker[T] {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: GetCircuitBreaker, return: %v => GetCircuitBreaker called", c.componentMetadata, c.CircuitBreaker.GetComponentMetadata())
	return c.CircuitBreaker
}

// GetInputChannel returns the input channel of the conduit.
// This function retrieves the input channel of the conduit component.
// Returns:
//   - chan T: The input channel of the conduit.
func (c *Conduit[T]) GetInputChannel() chan T {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: GetInputChannel, return: %v => GetInputChannel called", c.componentMetadata, c.InputChan)
	return c.InputChan
}

// GetOutputChannel returns the output channel of the conduit.
// This function retrieves the output channel of the conduit component.
// Returns:
//   - chan T: The output channel of the conduit.
func (c *Conduit[T]) GetOutputChannel() chan T {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: GetOutputChannel, return: %v => GetOutputChannel called", c.componentMetadata, c.OutputChan)
	return c.OutputChan
}

// IsStarted checks if the conduit has been started.
// This function checks whether the conduit has been started or not.
// Returns:
//   - bool: A boolean value indicating whether the conduit has been started.
func (c *Conduit[T]) IsStarted() bool {
	return atomic.LoadInt32(&c.started) == 1
}

// Load waits for all processing within the wire to complete and then returns the output buffer.
// This function waits for all processing within the wire component to complete,
// and then returns a copy of the output buffer containing the processed data.
// Returns:
//   - *bytes.Buffer: A copy of the output buffer containing the processed data.
func (c *Conduit[T]) Load() *bytes.Buffer {
	_ = c.Stop()

	if len(c.wires) > 0 && c.wires[len(c.wires)-1] != nil {
		return c.wires[len(c.wires)-1].GetOutputBuffer()
	}
	return new(bytes.Buffer)
}

// LoadAsJSONArray collects output into a JSON array. Assumes final output is JSON-compatible.
// This function collects the output of the wire component into a JSON array,
// assuming that the final output is JSON-compatible.
// Returns:
//   - []byte: The JSON array containing the output data.
//   - error: An error, if any occurred during the marshalling process.
func (c *Conduit[T]) LoadAsJSONArray() ([]byte, error) {
	_ = c.Stop()

	var items []T
	if c.OutputChan != nil {
		for item := range c.OutputChan {
			items = append(items, item)
		}
	}
	return json.Marshal(items)
}

// NotifyLoggers notifies loggers with the specified log level, format, and arguments.
// This function notifies registered loggers with the specified log level, format, and arguments.
// Parameters:
//   - level: The log level of the message.
//   - format: The format string for the log message.
//   - args: The arguments to be formatted into the log message.
func (c *Conduit[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if c.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range c.loggers {
			if logger == nil {
				continue
			}
			// Ensure we only acquire the lock once per logger to avoid deadlock or excessive locking overhead
			c.loggersLock.Lock()
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
			c.loggersLock.Unlock()
		}
	}
}

// SetComponentMetadata sets overrides for fields that are settable such as name and id.
// through which error messages are communicated.
// Parameters:
//   - name: The name to be set.
//   - id: The id to be set.
func (c *Conduit[T]) SetComponentMetadata(name string, id string) {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.componentMetadata.Name = name
	c.componentMetadata.ID = id
}

// SetConcurrencyControl configures concurrency control parasensors for the conduit.
// This function sets the buffer size and maximum number of concurrent routines for the conduit's wires.
// Parameters:
//   - bufferSize: The buffer size for each wire's channel.
//   - maxRoutines: The maximum number of concurrent routines for each wire.
func (c *Conduit[T]) SetConcurrencyControl(bufferSize int, maxRoutines int) {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.MaxBufferSize = bufferSize
	c.MaxConcurrency = maxRoutines
}

// SetInputChannel sets the input channel of the conduit.
// This function sets the input channel of the conduit component.
// Parameters:
//   - inputChan: The input channel to be set.
func (c *Conduit[T]) SetInputChannel(inputChan chan T) {
	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: SetInputChannel, change: w.inChan -> %v => SetInputChannel called", c.componentMetadata, inputChan)
	c.InputChan = inputChan
}

// Start starts the Conduit component.
// This function starts the Conduit component, initializing its configuration and its managed wires,
// starting processing routines, and managing generators.
func (c *Conduit[T]) Start(_ context.Context) error {
	if !atomic.CompareAndSwapInt32(&c.started, 0, 1) {
		return nil
	}

	// Start wires (use conduit context so Stop() reliably cancels everything).
	for _, w := range c.wires {
		if w == nil {
			continue
		}
		_ = w.Start(c.ctx)
	}

	// Start conduit-level generators (if any).
	if c.generators != nil {
		for _, g := range c.generators {
			if g == nil {
				continue
			}
			g.Start(c.ctx)
		}
	}

	// Forward from conduit InputChan to first wire ONLY if InputChan exists.
	// (Most usage goes through Submit; this is kept for compatibility.)
	if c.InputChan != nil && len(c.wires) > 0 && c.wires[0] != nil {
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
					_ = c.wires[0].Submit(c.ctx, elem)
				}
			}
		}()
	}

	// Forward conduit output into next conduit if chained.
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

// Submit submits an element to the conduit for processing.
// This function submits an element to the conduit component for processing.
// Parameters:
//   - ctx: The context for the submission operation.
//   - elem: The element to be submitted.
//
// Returns:
//   - error: An error, if any occurred during the submission process.
func (c *Conduit[T]) Submit(ctx context.Context, elem T) error {
	if len(c.wires) == 0 {
		return fmt.Errorf("no wires connected in conduit")
	}

	if c.CircuitBreaker != nil && !c.CircuitBreaker.Allow() {
		if len(c.CircuitBreaker.GetNeutralWires()) != 0 {
			for _, gw := range c.CircuitBreaker.GetNeutralWires() {
				err := gw.Submit(ctx, elem)
				if err != nil {
					c.NotifyLoggers(types.ErrorLevel, "component: %s, level: ERROR, result: FAILURE, event: Submit, element: %v, target: %s => Failure submitting to GroundWire!", c.componentMetadata, elem, gw.GetComponentMetadata())
					return err
				}
				c.NotifyLoggers(types.InfoLevel, "component: %s, level: INFO, result: SUCCESS, event: Submit, element: %v, target: %s => Element submitted to GroundWire...", c.componentMetadata, elem, gw.GetComponentMetadata())
			}
		}
	} else {
		err := c.wires[0].Submit(ctx, elem)
		if err != nil {
			return err
		}
		c.NotifyLoggers(types.InfoLevel, "component: %s, level: INFO, result: SUCCESS, event: Submit, element: %v, target: %s => Element submitted", c.componentMetadata, elem, c.wires[0].GetComponentMetadata())
	}

	return nil
}

// Terminate stops the Conduit component.
// This function stops the Conduit component, and cleanly shuts down its managed wires
func (c *Conduit[T]) Stop() error {
	if !atomic.CompareAndSwapInt32(&c.started, 1, 0) {
		return nil
	}

	c.terminateOnce.Do(func() {
		// Cancel first so forwarders exit.
		c.cancel()

		// Stop wires (this should close last wire output => closes c.OutputChan if last wire was wired to it).
		for _, w := range c.wires {
			if w == nil {
				continue
			}
			_ = w.Stop()
		}

		// Wait for internal forwarders (InputChan->wire, OutputChan->NextConduit).
		c.completeSignal.Wait()
	})

	return nil
}

// Restart stops, reinitializes, and restarts the Wire component.
func (c *Conduit[T]) Restart(ctx context.Context) error {
	// Stop the current operations if it is running.
	if atomic.LoadInt32(&c.started) == 1 {
		c.Stop()
	}

	c.configLock.Lock()
	defer c.configLock.Unlock()
	c.InputChan = (make(chan T, c.MaxBufferSize))
	c.OutputChan = (make(chan T, c.MaxBufferSize))

	for _, w := range c.wires {
		w.Start(ctx)
	}

	if c.NextConduit != nil {
		c.NextConduit.Restart(ctx)
	}
	// Restart the wire with the new context.
	return c.Start(ctx)
}
