// Package wire provides the public API for configuring and controlling
// data processing wires within the Electrician framework. A "wire" is a
// highly configurable, concurrent unit that accepts, transforms, and forwards
// data streams, and integrates with various components such as circuit breakers,
// generators, sensors, loggers, surge protectors, transformers, encoders, and insulators.
//
// This API enables you to:
//   - Connect and configure supporting components (e.g. circuit breakers, generators, sensors,
//     loggers, surge protectors, and transformers) to modify the behavior of the wire.
//   - Manage input/output channels and internal buffers for efficient, concurrent data handling.
//   - Control the lifecycle of the wire by starting, stopping, and restarting its processing routines.
//   - Retrieve internal state and configuration details for monitoring, debugging, or further adjustments.
//   - Handle errors and implement retry mechanisms using insulators and surge protectors.
//
// All methods are designed with concurrency in mind and include detailed logging for robust
// debugging and traceability.
package wire

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// ConnectCircuitBreaker attaches a circuit breaker to the wire.
// The circuit breaker regulates data flow based on predefined conditions,
// allowing the wire to prevent overload or failure propagation.
//
// Parameters:
//   - cb: The circuit breaker to be connected.
//
// This method is thread-safe and logs its operation at debug level.
func (w *Wire[T]) ConnectCircuitBreaker(cb types.CircuitBreaker[T]) {
	w.CircuitBreaker = cb
	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectCircuitBreaker",
		"component", w.componentMetadata,
		"event", "ConnectCircuitBreaker",
		"result", "SUCCESS",
		"circuitBreakerComponentMetadata", cb.GetComponentMetadata(),
		"circuitBreakerAddress", cb,
	)
}

// ConnectGenerator attaches one or more generator functions to the wire.
// Generators feed data into the wire’s processing pipeline and are automatically connected
// via their ConnectToComponent method.
//
// Parameters:
//   - generator: One or more generator functions to be connected.
func (w *Wire[T]) ConnectGenerator(generator ...types.Generator[T]) {
	w.generators = append(w.generators, generator...)
	for _, g := range generator {
		g.ConnectToComponent(w)
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectGenerator",
			"component", w.componentMetadata,
			"event", "ConnectGenerator",
			"result", "SUCCESS",
			"generatorComponentMetadata", g.GetComponentMetadata(),
			"generatorAddress", g,
		)
	}
}

// ConnectLogger attaches one or more logger components to the wire.
// Loggers capture operational events and errors, which is essential for monitoring and debugging.
//
// Parameters:
//   - logger: One or more logger instances to be connected.
func (w *Wire[T]) ConnectLogger(logger ...types.Logger) {
	w.loggers = append(w.loggers, logger...)
	for _, l := range logger {
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectLogger",
			"component", w.componentMetadata,
			"event", "ConnectLogger",
			"result", "SUCCESS",
			"loggerAddress", l,
		)
	}
}

// ConnectSensor attaches one or more sensor components to the wire.
// Sensors monitor data throughput and various internal metrics of the wire,
// aiding in real-time performance tracking and diagnostics.
//
// Parameters:
//   - sensor: One or more sensor components to be connected.
func (w *Wire[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	w.sensors = append(w.sensors, sensor...)
	for _, m := range sensor {
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectSensor",
			"component", w.componentMetadata,
			"event", "ConnectSensor",
			"result", "SUCCESS",
			"sensorComponentMetadata", m.GetComponentMetadata(),
			"sensorAddress", m,
		)
	}
}

// ConnectSurgeProtector attaches a surge protector to the wire.
// A surge protector helps manage and divert data during high-load conditions,
// ensuring system resilience. Note that attaching a surge protector after the wire
// has started processing may lead to a panic, as dynamic configuration changes are not allowed.
//
// Parameters:
//   - protector: The surge protector to be connected.
//
// This method also sets internal rate-limiting parameters based on the surge protector's configuration.
func (w *Wire[T]) ConnectSurgeProtector(protector types.SurgeProtector[T]) {
	w.surgeProtector = protector
	_, _, maxRetryAttempts, _ := w.surgeProtector.GetRateLimit()
	w.queueRetryMaxAttempt = maxRetryAttempts
	w.surgeProtector.ConnectComponent(w)
	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectSurgeProtector",
		"component", w.componentMetadata,
		"event", "ConnectSurgeProtector",
		"result", "SUCCESS",
		"surgeProtectorComponentMetadata", w.GetComponentMetadata(),
		"surgeProtectorAddress", w.surgeProtector,
	)
}

// ConnectTransformer attaches one or more transformation functions to the wire.
// Transformers modify or enrich incoming data before it is further processed.
//
// Parameters:
//   - transformation: One or more transformation functions to be connected.
func (w *Wire[T]) ConnectTransformer(transformation ...types.Transformer[T]) {
	w.transformations = append(w.transformations, transformation...)
	for _, t := range transformation {
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectTransformer",
			"component", w.componentMetadata,
			"event", "ConnectTransformer",
			"result", "SUCCESS",
			"transformerAddress", t,
		)
	}
}

// GetCircuitBreaker retrieves the circuit breaker attached to the wire.
// This method is intended for inspection purposes and logs the retrieval operation.
//
// Returns:
//   - types.CircuitBreaker[T]: The currently attached circuit breaker.
//
// Note: This method uses a locking mechanism to ensure thread safety.
func (w *Wire[T]) GetCircuitBreaker() types.CircuitBreaker[T] {
	w.cbLock.Lock()
	defer w.cbLock.Unlock()
	w.NotifyLoggers(
		types.DebugLevel,
		"GetCircuitBreaker",
		"component", w.componentMetadata,
		"event", "GetCircuitBreaker",
		"result", "SUCCESS",
		"circuitBreakerMetadata", w.GetComponentMetadata(),
		"circuitBreakerAddress", w.CircuitBreaker,
	)
	return w.CircuitBreaker
}

// GetComponentMetadata returns the metadata associated with the wire component.
// This metadata includes identification and descriptive information used for logging and diagnostics.
//
// Returns:
//   - types.ComponentMetadata: The metadata structure for the wire.
func (w *Wire[T]) GetComponentMetadata() types.ComponentMetadata {
	w.NotifyLoggers(
		types.DebugLevel,
		"GetComponentMetadata",
		"component", w.componentMetadata,
		"event", "GetComponentMetadata",
		"result", "SUCCESS",
		"componentMetadata", w.componentMetadata,
	)
	return w.componentMetadata
}

// GetInputChannel returns the input channel through which the wire receives data.
//
// Returns:
//   - chan T: The input channel of the wire.
func (w *Wire[T]) GetInputChannel() chan T {
	w.NotifyLoggers(
		types.DebugLevel,
		"GetInputChannel",
		"component", w.componentMetadata,
		"event", "GetInputChannel",
		"result", "SUCCESS",
		"inputChannelAddress", w.inChan,
	)
	return w.inChan
}

// GetGenerators retrieves the list of generator functions currently connected to the wire.
//
// Returns:
//   - []types.Generator[T]: A slice containing all attached generators.
func (w *Wire[T]) GetGenerators() []types.Generator[T] {
	w.NotifyLoggers(
		types.DebugLevel,
		"GetGenerators",
		"component", w.componentMetadata,
		"event", "GetGenerators",
		"result", "SUCCESS",
		"generatorAddress", w.generators,
	)
	return w.generators
}

// GetOutputChannel returns the output channel used to emit processed data.
//
// Returns:
//   - chan T: The output channel of the wire.
func (w *Wire[T]) GetOutputChannel() chan T {
	w.NotifyLoggers(
		types.DebugLevel,
		"GetOutputChannel",
		"component", w.componentMetadata,
		"event", "GetOutputChannel",
		"result", "SUCCESS",
		"outputChannelAddress", w.OutputChan,
	)
	return w.OutputChan
}

// GetOutputBuffer returns the internal output buffer where processed data is stored
// prior to being sent through the output channel.
//
// Returns:
//   - *bytes.Buffer: The pointer to the wire's output buffer.
func (w *Wire[T]) GetOutputBuffer() *bytes.Buffer {
	w.NotifyLoggers(
		types.DebugLevel,
		"GetOutputBuffer",
		"component", w.componentMetadata,
		"event", "GetOutputBuffer",
		"result", "SUCCESS",
		"outputBufferAddress", w.OutputBuffer,
	)
	return w.OutputBuffer
}

// IsStarted indicates whether the wire has been started.
//
// Returns:
//   - bool: True if the wire is currently running; otherwise, false.
func (w *Wire[T]) IsStarted() bool {
	return atomic.LoadInt32(&w.started) == 1
}

// Load blocks until all processing is complete, then returns a copy of the output buffer.
// This method is useful for synchronously retrieving the final processed data.
//
// Returns:
//   - *bytes.Buffer: A new buffer containing the processed output data.
func (w *Wire[T]) Load() *bytes.Buffer {
	// Wait for the completion signal indicating that all processing is done.
	<-w.completeSignal

	// Safely terminate the wire's operations.
	w.Stop()

	// Lock the bufferMutex to ensure safe access to the OutputBuffer.
	w.bufferMutex.Lock()
	buf := make([]byte, w.OutputBuffer.Len())
	copy(buf, w.OutputBuffer.Bytes())
	w.bufferMutex.Unlock()
	// Log the buffer copy and return it.
	w.NotifyLoggers(
		types.DebugLevel,
		"Load Called",
		"component", w.componentMetadata,
		"event", "Load",
		"result", "SUCCESS",
		"buffer", buf,
	)
	return bytes.NewBuffer(buf)
}

// LoadAsJSONArray collects all processed output into a JSON array.
// It assumes that the output data is JSON-compatible.
//
// Returns:
//   - []byte: The marshaled JSON array representing the output data.
//   - error: An error value if marshalling fails.
func (w *Wire[T]) LoadAsJSONArray() ([]byte, error) {
	w.wg.Wait() // Ensure all data is processed before marshalling.
	var items []T
	for item := range w.GetOutputChannel() {
		items = append(items, item)
	}
	w.NotifyLoggers(
		types.DebugLevel,
		"LoadAsJSONArray Called",
		"component", w.componentMetadata,
		"event", "LoadAsJSONArray",
		"result", "SUCCESS",
		"itemsInSlice", items,
	)
	return json.Marshal(items)
}

// SetComponentMetadata updates the wire's metadata, such as name and ID.
// This metadata is used for identification and logging purposes.
//
// Parameters:
//   - name: The new name for the component.
//   - id: The new identifier for the component.
func (w *Wire[T]) SetComponentMetadata(name string, id string) {
	old := w.componentMetadata
	w.componentMetadata.Name = name
	w.componentMetadata.ID = id
	w.NotifyLoggers(
		types.DebugLevel,
		"SetComponentMetadata Called",
		"component", w.componentMetadata,
		"event", "SetComponentMetadata",
		"result", "SUCCESS",
		"oldComponentMetadata", old,
		"newComponentMetadata", w.componentMetadata,
	)
}

// SetConcurrencyControl adjusts the wire's buffer size and maximum concurrent routines.
// This allows dynamic tuning of the wire's throughput and parallel processing capacity.
//
// Parameters:
//   - bufferSize: The new size for the internal buffers.
//   - maxConcurrency: The new maximum number of concurrent processing routines.
func (w *Wire[T]) SetConcurrencyControl(bufferSize int, maxConcurrency int) {
	oldBufferSize := w.maxBufferSize
	oldMaxConcurrency := w.maxConcurrency
	w.maxBufferSize = bufferSize
	w.maxConcurrency = maxConcurrency
	w.NotifyLoggers(
		types.DebugLevel,
		"SetConcurrencyControl Called",
		"component", w.componentMetadata,
		"event", "SetConcurrencyControl",
		"result", "SUCCESS",
		"oldMaxBufferSize", oldBufferSize,
		"oldMaxConcurrency", oldMaxConcurrency,
		"newMaxBufferSize", w.maxBufferSize,
		"newMaxConcurrency", w.maxConcurrency,
	)
}

// SetSemaphore replaces the concurrency semaphore used by the wire.
// The semaphore controls the number of concurrent routines allowed during processing.
//
// Parameters:
//   - sem: A pointer to the new semaphore channel.
func (w *Wire[T]) SetSemaphore(sem *chan struct{}) {
	oldSemaphore := w.concurrencySem
	w.concurrencySem = *sem
	w.NotifyLoggers(
		types.DebugLevel,
		"SetSemaphore Called",
		"component", w.componentMetadata,
		"event", "SetSemaphore",
		"result", "SUCCESS",
		"oldSemaphoreAddress", oldSemaphore,
		"newSemaphoreAddress", w.concurrencySem,
	)
}

// SetEncoder assigns a new encoder to the wire.
// The encoder is responsible for serializing data before it is transmitted out of the wire.
//
// Parameters:
//   - e: The encoder instance to be set.
func (w *Wire[T]) SetEncoder(e types.Encoder[T]) {
	oldEncoder := w.encoder
	w.encoder = e
	w.NotifyLoggers(
		types.DebugLevel,
		"SetEncoder Called",
		"component", w.componentMetadata,
		"event", "SetEncoder",
		"result", "SUCCESS",
		"oldEncoderAddress", oldEncoder,
		"newEncoderAddress", w.encoder,
	)
}

// SetInsulator configures the error recovery mechanism for the wire.
// The insulator function provides a retry mechanism to handle transient errors during processing.
//
// Parameters:
//   - retryFunc: The function invoked for retrying failed operations.
//   - threshold: The maximum number of retry attempts before giving up.
//   - interval: The delay duration between consecutive retry attempts.
func (w *Wire[T]) SetInsulator(retryFunc func(ctx context.Context, elem T, err error) (T, error), threshold int, interval time.Duration) {
	oldInsulatorFunc := w.insulatorFunc
	oldRetryInterval := w.retryInterval
	oldRetryThreshold := w.retryThreshold
	w.insulatorFunc = retryFunc
	w.retryInterval = interval
	w.retryThreshold = threshold
	w.NotifyLoggers(
		types.DebugLevel,
		"SetInsulator Called",
		"component", w.componentMetadata,
		"event", "SetInsulator",
		"result", "SUCCESS",
		"oldInsulatorFunc", oldInsulatorFunc,
		"oldRetryInterval", oldRetryInterval,
		"oldRetryThreshold", oldRetryThreshold,
		"newInsulatorFunc", w.insulatorFunc,
		"newRetryInterval", w.retryInterval,
		"newRetryThreshold", w.retryThreshold,
	)
}

// SetOutputBuffer assigns a new output buffer to the wire.
// The output buffer temporarily stores processed data before it is emitted through the output channel.
//
// Parameters:
//   - b: The bytes.Buffer to be used as the new output buffer.
func (w *Wire[T]) SetOutputBuffer(b bytes.Buffer) {
	oldOutputBuffer := w.OutputBuffer
	w.OutputBuffer = &b
	w.NotifyLoggers(
		types.DebugLevel,
		"SetOutputBuffer Called",
		"component", w.componentMetadata,
		"event", "SetOutputBuffer",
		"result", "SUCCESS",
		"oldOutputBufferAddress", oldOutputBuffer,
		"newOutputBufferAddress", w.OutputBuffer,
	)
}

// SetErrorChannel assigns a new error channel to the wire.
// Errors encountered during processing are sent through this channel.
//
// Parameters:
//   - errorChan: The error channel to be set.
func (w *Wire[T]) SetErrorChannel(errorChan chan types.ElementError[T]) {
	oldErrorChan := w.errorChan
	w.errorChan = errorChan
	w.NotifyLoggers(
		types.DebugLevel,
		"SetErrorChannel Called",
		"component", w.componentMetadata,
		"event", "SetErrorChannel",
		"result", "SUCCESS",
		"oldErrorChannelAddress", oldErrorChan,
		"newErrorChannelAddress", w.errorChan,
	)
}

// SetInputChannel replaces the input channel used by the wire.
// The input channel is where data destined for processing is received.
//
// Parameters:
//   - inChan: The new input channel to be used.
func (w *Wire[T]) SetInputChannel(inChan chan T) {
	oldInputChannelAddress := w.inChan
	w.inChan = inChan
	w.NotifyLoggers(
		types.DebugLevel,
		"SetInputChannel Called",
		"component", w.componentMetadata,
		"event", "SetInputChannel",
		"result", "SUCCESS",
		"oldInputChannelAddress", oldInputChannelAddress,
		"newInputChannelAddress", w.inChan,
	)
}

// SetOutputChannel replaces the output channel used by the wire.
// Processed data is sent through this channel for further consumption.
//
// Parameters:
//   - outChan: The new output channel to be set.
func (w *Wire[T]) SetOutputChannel(outChan chan T) {
	oldOutputChan := w.OutputChan
	w.OutputChan = outChan
	w.NotifyLoggers(
		types.DebugLevel,
		"SetOutputChannel Called",
		"component", w.componentMetadata,
		"event", "SetOutputChannel",
		"result", "SUCCESS",
		"oldOutputChannelAddress", oldOutputChan,
		"newOutputChannelAddress", w.OutputChan,
	)
}

// Start initiates the wire's processing routines.
// It marks the wire as started, triggers asynchronous routines for error handling,
// surge protection, and data transformation, and also starts any connected generators.
//
// Parameters:
//   - ctx: The context used to manage cancellation and timeouts.
//
// Returns:
//   - error: An error if startup fails, otherwise nil.
func (w *Wire[T]) Start(ctx context.Context) error {
	atomic.StoreInt32(&w.started, 1)
	w.notifyStart()

	// Account for the error handling goroutine.
	w.wg.Add(1)
	go w.handleErrorElements()

	if w.surgeProtector != nil && w.surgeProtector.IsResisterConnected() {
		go w.processResisterElements(ctx)
	}

	// Start processing elements (transformElements spawns routines as needed).
	w.transformElements()

	go func() {
		w.wg.Wait()
		w.Stop()
		w.notifyComplete()
	}()

	if w.generators != nil {
		for _, g := range w.generators {
			if !g.IsStarted() {
				g.Start(ctx)
			}
		}
	}

	return nil
}

// Submit sends an element to the wire for processing.
// If DecryptOptions are enabled, the wire assumes this inbound 'elem' is ciphertext
// and decrypts it before continuing the normal flow.
// Submit sends an element to the wire for processing.
// If DecryptOptions are enabled, the wire assumes inbound 'elem' is ciphertext.
// We decrypt it before continuing normal circuit breaker / surge checks.
func (w *Wire[T]) Submit(ctx context.Context, elem T) error {
	// If decryption is enabled, decrypt now
	if w.DecryptOptions != nil && w.DecryptOptions.Enabled {
		decrypted, err := w.decryptItem(elem)
		if err != nil {
			// Log or handle decryption error
			w.NotifyLoggers(types.ErrorLevel,
				"component: %v, level: ERROR, event: Submit => decryption failed: %v",
				w.componentMetadata, err)
			return fmt.Errorf("failed to decrypt inbound data: %w", err)
		}
		elem = decrypted
	}

	// Create an Element wrapper
	element := types.NewElement[T](elem)

	// Circuit breaker check
	if w.CircuitBreaker != nil && !w.CircuitBreaker.Allow() {
		return w.handleCircuitBreakerTrip(ctx, element.Data)
	}
	// Surge protector checks
	if w.surgeProtector != nil && w.surgeProtector.IsTripped() {
		w.notifySurgeProtectorSubmit(element.Data)
		return w.surgeProtector.Submit(ctx, element)
	}
	if w.surgeProtector != nil && !w.surgeProtector.TryTake() {
		w.notifyRateLimit(elem, time.Now().Add(w.surgeProtector.GetTimeUntilNextRefill()).Format("2006-01-02 15:04:05"))
		return w.surgeProtector.Submit(ctx, element)
	}

	// Finally, queue the item for the wire’s transform process
	return w.submitNormally(ctx, element.Data)
}

// Stop gracefully terminates the wire's processing routines and cleans up resources.
// It cancels the context, waits for all goroutines to finish, and then closes all associated channels.
//
// Returns:
//   - error: An error if the shutdown encounters issues, otherwise nil.
func (w *Wire[T]) Stop() error {
	// Ensure termination logic is executed only once.
	if atomic.CompareAndSwapInt32(&w.started, 1, 0) {
		w.notifyStop()
		w.terminateOnce.Do(func() {
			// First, mark the wire as closed so new sends don't occur.
			w.closeLock.Lock()
			w.isClosed = true
			w.closeLock.Unlock()

			// Cancel the context so that generators and other components can shut down.
			w.cancel()

			// Wait for all worker goroutines (transformers, error handlers, etc.) to finish.
			w.wg.Wait()

			// Now it is safe to close all channels.
			w.closeLock.Lock()
			defer w.closeLock.Unlock()

			w.closeOutputChanOnce.Do(func() { close(w.OutputChan) })
			w.closeInputChanOnce.Do(func() { close(w.inChan) })
			w.closeErrorChanOnce.Do(func() { close(w.errorChan) })

			// Reset the waitgroup in case of a future restart.
			w.wg = sync.WaitGroup{}
		})
	}
	return nil
}

// SetDecryptOptions configures AES-GCM decryption for inbound items.
// If called, the wire assumes *all* data submitted is ciphertext.
func (w *Wire[T]) SetDecryptOptions(opts *relay.SecurityOptions, key string) {
	if w.IsStarted() {
		panic(fmt.Sprintf("cannot set decrypt options after wire has started: %s", w.componentMetadata))
	}
	w.DecryptOptions = opts
	w.DecryptKey = key

	w.NotifyLoggers(types.DebugLevel,
		"component: %v, level: DEBUG, event: SetDecryptOptions => inbound decryption configured",
		w.componentMetadata,
	)
}

// SetEncryptOptions configures AES-GCM encryption for outbound items.
// If called, the wire will encrypt data right before placing it on OutputChan (or errorChan).
func (w *Wire[T]) SetEncryptOptions(opts *relay.SecurityOptions, key string) {
	if w.IsStarted() {
		panic(fmt.Sprintf("cannot set encrypt options after wire has started: %s", w.componentMetadata))
	}
	w.EncryptOptions = opts
	w.EncryptKey = key

	w.NotifyLoggers(types.DebugLevel,
		"component: %v, level: DEBUG, event: SetEncryptOptions => outbound encryption configured",
		w.componentMetadata,
	)
}

// Restart stops the wire, reinitializes its internal state, and starts it again.
// This is useful for applying new configuration settings or recovering from certain error conditions.
//
// Parameters:
//   - ctx: The context used to manage cancellation and timeouts during restart.
//
// Returns:
//   - error: An error if the restart process fails, otherwise nil.
func (w *Wire[T]) Restart(ctx context.Context) error {
	if atomic.LoadInt32(&w.started) == 1 {
		if err := w.Stop(); err != nil {
			return fmt.Errorf("failed to stop wire during restart: %w", err)
		}
	}

	w.concurrencySem = make(chan struct{}, w.maxConcurrency)
	w.SetInputChannel(make(chan T, w.maxBufferSize))
	w.SetOutputChannel(make(chan T, w.maxBufferSize))
	w.SetErrorChannel(make(chan types.ElementError[T], w.maxBufferSize))
	w.OutputBuffer = &bytes.Buffer{}

	return w.Start(ctx)
}
