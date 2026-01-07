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
	"sync"
	"sync/atomic"
	"time"

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
	// Track non-nil loggers
	var n int32
	for _, l := range logger {
		if l != nil {
			n++
		}
	}

	w.loggersLock.Lock()
	w.loggers = append(w.loggers, logger...)
	w.loggersLock.Unlock()

	if n != 0 {
		atomic.AddInt32(&w.loggerCount, n)
	}

	// Keep your existing debug logs if you want, but only if loggers exist.
	// (Otherwise you’re paying interface-boxing cost for nothing.)
	if atomic.LoadInt32(&w.loggerCount) != 0 {
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
}

func (w *Wire[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	// Track non-nil sensors
	var n int32
	for _, s := range sensor {
		if s != nil {
			n++
		}
	}

	w.sensorLock.Lock()
	w.sensors = append(w.sensors, sensor...)
	w.sensorLock.Unlock()

	if n != 0 {
		atomic.AddInt32(&w.sensorCount, n)
	}

	// Optional debug log gating
	if atomic.LoadInt32(&w.loggerCount) != 0 {
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

func (w *Wire[T]) LoadAsJSONArray() ([]byte, error) {
	// Ensure processing is finished and OutputChan is closed so the range terminates.
	_ = w.Stop()

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

func (w *Wire[T]) Load() *bytes.Buffer {
	// Ensure processing is finished before reading OutputBuffer.
	_ = w.Stop()

	w.bufferMutex.Lock()
	buf := make([]byte, w.OutputBuffer.Len())
	copy(buf, w.OutputBuffer.Bytes())
	w.bufferMutex.Unlock()

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

// FastSubmit is an optional fast lane used by Generator when available.
// It MUST preserve correctness: if any feature that needs Submit() is enabled,
// it falls back to Submit().
func (w *Wire[T]) FastSubmit(ctx context.Context, elem T) error {
	// If any intercepting feature is enabled, fall back.
	if w.CircuitBreaker != nil || w.surgeProtector != nil || w.insulatorFunc != nil {
		return w.Submit(ctx, elem)
	}
	if atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0 {
		return w.Submit(ctx, elem) // preserve submit notifications semantics
	}

	// Direct enqueue: baseline-like.
	select {
	case w.inChan <- elem:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
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
func (w *Wire[T]) Start(_ context.Context) error {
	if !atomic.CompareAndSwapInt32(&w.started, 0, 1) {
		return nil // or return an error
	}

	w.notifyStart()

	// error handler
	w.wg.Add(1)
	go w.handleErrorElements()

	// resister should be tied to w.ctx and tracked in wg
	if w.surgeProtector != nil && w.surgeProtector.IsResisterConnected() {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.processResisterElements(w.ctx)
		}()
	}

	w.fastPathEnabled = w.computeFastPathEnabled()
	if w.fastPathEnabled {
		w.fastTransform = w.transformations[0]
	}

	// workers
	w.transformElements()

	// generators should also use w.ctx so Stop actually stops them
	for _, g := range w.generators {
		if g != nil && !g.IsStarted() {
			g.Start(w.ctx)
		}
	}

	return nil
}

// Submit sends an element to the wire for processing.
// This method handles the submission of data while integrating with the circuit breaker
// and surge protector mechanisms. Depending on the current state of these components,
// the submission may be handled normally or deferred via the surge protector.
//
// Parameters:
//   - ctx: The context to manage cancellation or timeouts for the submission.
//   - elem: The data element to be processed.
//
// Returns:
//   - error: An error if the submission fails, or nil on success.
func (w *Wire[T]) Submit(ctx context.Context, elem T) error {
	// Circuit breaker: no Element wrapper required.
	if w.CircuitBreaker != nil && !w.CircuitBreaker.Allow() {
		return w.handleCircuitBreakerTrip(ctx, elem)
	}

	// Surge protector path: only now do we allocate Element metadata.
	if w.surgeProtector != nil {
		// Cheap ID by default. Swap to NewElementHashed(elem) only if you really need dedupe-by-content.
		element := types.NewElementFast[T](elem)

		if w.surgeProtector.IsTripped() {
			w.notifySurgeProtectorSubmit(element.Data)
			return w.surgeProtector.Submit(ctx, element)
		}

		if w.surgeProtector.IsBeingRateLimited() && !w.surgeProtector.TryTake() {
			w.notifyRateLimit(elem, time.Now().Add(w.surgeProtector.GetTimeUntilNextRefill()).Format("2006-01-02 15:04:05"))
			return w.surgeProtector.Submit(ctx, element)
		}
		// If rate limiting allows, fall through to normal submit.
	}

	// Fast path: no Element, no hashing, no allocations beyond what user code does.
	return w.submitNormally(ctx, elem)
}

// Stop gracefully terminates the wire's processing routines and cleans up resources.
// It cancels the context, waits for all goroutines to finish, and closes all associated channels.
//
// Returns:
//   - error: An error if the shutdown encounters issues, otherwise nil.
func (w *Wire[T]) Stop() error {
	if !atomic.CompareAndSwapInt32(&w.started, 1, 0) {
		return nil
	}

	w.notifyStop()

	w.terminateOnce.Do(func() {
		// Mark closed BEFORE cancel/close.
		w.closeLock.Lock()
		w.isClosed = true
		w.closeLock.Unlock()

		// Cancel wire context (workers exit).
		w.cancel()

		// Wait for internal goroutines you accounted for in wg.
		w.wg.Wait()

		// Now it’s safe to close outward-facing channels.
		w.closeOutputChanOnce.Do(func() { close(w.OutputChan) })
		w.closeErrorChanOnce.Do(func() { close(w.errorChan) })

		// Optional: do NOT close inChan; it’s the one that causes send panics from external producers.
		// If you insist on closing it, you MUST harden Submit/submitNormally to never panic (see #3).
		// w.closeInputChanOnce.Do(func() { close(w.inChan) })

		// Signal completion exactly once.
		select {
		case <-w.completeSignal:
			// already closed/used (if you implement it as a close; otherwise ignore)
		default:
			// if you want “close channel means complete”, add a once/close here
		}
	})

	return nil
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
	// Stop if running
	_ = w.Stop()

	// Rebind context (critical)
	w.ctx, w.cancel = context.WithCancel(ctx)

	// Reset lifecycle guards so Stop() works again after restart
	w.terminateOnce = sync.Once{}
	w.closeOutputChanOnce = sync.Once{}
	w.closeInputChanOnce = sync.Once{}
	w.closeErrorChanOnce = sync.Once{}

	// Reset closed flag
	w.closeLock.Lock()
	w.isClosed = false
	w.closeLock.Unlock()

	// Rebuild channels/buffers
	w.SetInputChannel(make(chan T, w.maxBufferSize))
	w.SetOutputChannel(make(chan T, w.maxBufferSize))
	w.SetErrorChannel(make(chan types.ElementError[T], w.maxBufferSize))
	w.OutputBuffer = &bytes.Buffer{}

	// Restart CB ticker if you rely on it
	if w.CircuitBreaker != nil {
		go w.startCircuitBreakerTicker()
	}

	return w.Start(w.ctx)
}
