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

// ConnectCircuitBreaker attaches a circuit breaker that can reject submissions and divert
// elements to neutral wires when open.
//
// Configuration should be finalized before Start; connecting while processing is running
// is not safe unless the rest of the package is made config-race-safe.
func (w *Wire[T]) ConnectCircuitBreaker(cb types.CircuitBreaker[T]) {
	w.cbLock.Lock()
	w.CircuitBreaker = cb
	w.cbLock.Unlock()

	if cb == nil {
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectCircuitBreaker",
			"component", w.componentMetadata,
			"event", "ConnectCircuitBreaker",
			"result", "SUCCESS",
			"circuitBreaker", nil,
		)
		return
	}

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

// ConnectGenerator registers generators that can feed the wire.
func (w *Wire[T]) ConnectGenerator(generators ...types.Generator[T]) {
	if len(generators) == 0 {
		return
	}

	// Compact in-place to drop nils without allocating.
	n := 0
	for _, g := range generators {
		if g != nil {
			generators[n] = g
			n++
		}
	}
	if n == 0 {
		return
	}
	generators = generators[:n]

	w.generators = append(w.generators, generators...)
	for _, g := range generators {
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

// ConnectLogger registers loggers. Nil loggers are ignored.
func (w *Wire[T]) ConnectLogger(loggers ...types.Logger) {
	if len(loggers) == 0 {
		return
	}

	// Compact in-place to drop nils without allocating.
	n := 0
	for _, l := range loggers {
		if l != nil {
			loggers[n] = l
			n++
		}
	}
	if n == 0 {
		return
	}
	loggers = loggers[:n]

	w.loggersLock.Lock()
	w.loggers = append(w.loggers, loggers...)
	w.loggersLock.Unlock()

	atomic.AddInt32(&w.loggerCount, int32(n))

	// Single event to avoid spam at startup.
	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectLogger",
		"component", w.componentMetadata,
		"event", "ConnectLogger",
		"result", "SUCCESS",
		"count", n,
	)
}

// ConnectSensor registers sensors. Nil sensors are ignored.
func (w *Wire[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	if len(sensors) == 0 {
		return
	}

	// Compact in-place to drop nils without allocating.
	n := 0
	for _, s := range sensors {
		if s != nil {
			sensors[n] = s
			n++
		}
	}
	if n == 0 {
		return
	}
	sensors = sensors[:n]

	w.sensorLock.Lock()
	w.sensors = append(w.sensors, sensors...)
	w.sensorLock.Unlock()

	atomic.AddInt32(&w.sensorCount, int32(n))

	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectSensor",
		"component", w.componentMetadata,
		"event", "ConnectSensor",
		"result", "SUCCESS",
		"count", n,
	)
}

// ConnectSurgeProtector attaches a surge protector (rate limiter / queue).
func (w *Wire[T]) ConnectSurgeProtector(protector types.SurgeProtector[T]) {
	w.surgeProtector = protector
	if protector == nil {
		w.NotifyLoggers(
			types.DebugLevel,
			"ConnectSurgeProtector",
			"component", w.componentMetadata,
			"event", "ConnectSurgeProtector",
			"result", "SUCCESS",
			"surgeProtector", nil,
		)
		return
	}

	_, _, maxRetryAttempts, _ := protector.GetRateLimit()
	w.queueRetryMaxAttempt = maxRetryAttempts

	protector.ConnectComponent(w)

	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectSurgeProtector",
		"component", w.componentMetadata,
		"event", "ConnectSurgeProtector",
		"result", "SUCCESS",
		"surgeProtectorComponentMetadata", protector.GetComponentMetadata(),
		"surgeProtectorAddress", protector,
	)
}

// ConnectTransformer appends transformations to the processing chain. Nil transformers are ignored.
//
// Order matters.
func (w *Wire[T]) ConnectTransformer(transformers ...types.Transformer[T]) {
	if len(transformers) == 0 {
		return
	}
	if atomic.LoadInt32(&w.started) == 1 {
		panic("wire: ConnectTransformer called after Start")
	}
	if w.transformerFactory != nil {
		panic("wire: cannot use ConnectTransformer when TransformerFactory is set; wrap transforms inside the factory closure instead")
	}

	// Compact in-place to drop nils without allocating.
	n := 0
	for _, t := range transformers {
		if t != nil {
			transformers[n] = t
			n++
		}
	}
	if n == 0 {
		return
	}
	transformers = transformers[:n]

	w.transformations = append(w.transformations, transformers...)
	w.NotifyLoggers(
		types.DebugLevel,
		"ConnectTransformer",
		"component", w.componentMetadata,
		"event", "ConnectTransformer",
		"result", "SUCCESS",
		"count", n,
	)
}

func (w *Wire[T]) GetCircuitBreaker() types.CircuitBreaker[T] {
	w.cbLock.Lock()
	cb := w.CircuitBreaker
	w.cbLock.Unlock()
	return cb
}

func (w *Wire[T]) GetComponentMetadata() types.ComponentMetadata {
	return w.componentMetadata
}

func (w *Wire[T]) GetInputChannel() chan T {
	return w.inChan
}

func (w *Wire[T]) GetGenerators() []types.Generator[T] {
	return w.generators
}

func (w *Wire[T]) GetOutputChannel() chan T {
	return w.OutputChan
}

func (w *Wire[T]) GetOutputBuffer() *bytes.Buffer {
	return w.OutputBuffer
}

func (w *Wire[T]) IsStarted() bool {
	return atomic.LoadInt32(&w.started) == 1
}

// LoadAsJSONArray stops the wire (best-effort) and drains whatever is currently available
// from the output channel into a JSON array.
//
// This method never blocks waiting for future output.
func (w *Wire[T]) LoadAsJSONArray() ([]byte, error) {
	_ = w.Stop()

	out := w.OutputChan
	if out == nil {
		return []byte("[]"), nil
	}

	items := make([]T, 0, len(out))
	for {
		select {
		case v, ok := <-out:
			if !ok {
				return json.Marshal(items)
			}
			items = append(items, v)
		default:
			// Channel is open but currently empty (or never started). Do not block.
			return json.Marshal(items)
		}
	}
}

// Load stops the wire (best-effort) and returns a copy of the current output buffer.
func (w *Wire[T]) Load() *bytes.Buffer {
	_ = w.Stop()

	w.bufferMutex.Lock()
	buf := make([]byte, w.OutputBuffer.Len())
	copy(buf, w.OutputBuffer.Bytes())
	w.bufferMutex.Unlock()

	w.NotifyLoggers(
		types.DebugLevel,
		"Load",
		"component", w.componentMetadata,
		"event", "Load",
		"result", "SUCCESS",
		"bytes", len(buf),
	)

	return bytes.NewBuffer(buf)
}

// SetComponentMetadata updates the component metadata. Intended for configuration time.
func (w *Wire[T]) SetComponentMetadata(name string, id string) {
	old := w.componentMetadata
	w.componentMetadata.Name = name
	w.componentMetadata.ID = id

	w.NotifyLoggers(
		types.DebugLevel,
		"SetComponentMetadata",
		"component", w.componentMetadata,
		"event", "SetComponentMetadata",
		"result", "SUCCESS",
		"oldComponentMetadata", old,
		"newComponentMetadata", w.componentMetadata,
	)
}

// SetConcurrencyControl sets the internal channel buffer size and worker concurrency.
// Values are sanitized to avoid panics on Restart.
func (w *Wire[T]) SetConcurrencyControl(bufferSize int, maxConcurrency int) {
	if bufferSize < 0 {
		bufferSize = 0
	}
	if maxConcurrency < 1 {
		maxConcurrency = 1
	}

	oldBufferSize := w.maxBufferSize
	oldMaxConcurrency := w.maxConcurrency

	w.maxBufferSize = bufferSize
	w.maxConcurrency = maxConcurrency

	w.NotifyLoggers(
		types.DebugLevel,
		"SetConcurrencyControl",
		"component", w.componentMetadata,
		"event", "SetConcurrencyControl",
		"result", "SUCCESS",
		"oldMaxBufferSize", oldBufferSize,
		"oldMaxConcurrency", oldMaxConcurrency,
		"newMaxBufferSize", w.maxBufferSize,
		"newMaxConcurrency", w.maxConcurrency,
	)
}

// FastSubmit is the minimal-overhead submit path when no circuit breaker, surge protector,
// insulator, or per-element telemetry is attached.
//
// It still respects both the caller context and the wire context.
func (w *Wire[T]) FastSubmit(ctx context.Context, elem T) error {
	if w.CircuitBreaker != nil || w.surgeProtector != nil || w.insulatorFunc != nil {
		return w.Submit(ctx, elem)
	}
	if atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0 {
		return w.Submit(ctx, elem)
	}

	in := w.inChan
	select {
	case in <- elem:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-w.ctx.Done():
		return w.ctx.Err()
	}
}

func (w *Wire[T]) SetEncoder(e types.Encoder[T]) {
	oldEncoder := w.encoder
	w.encoder = e

	w.NotifyLoggers(
		types.DebugLevel,
		"SetEncoder",
		"component", w.componentMetadata,
		"event", "SetEncoder",
		"result", "SUCCESS",
		"oldEncoderAddress", oldEncoder,
		"newEncoderAddress", w.encoder,
	)
}

func (w *Wire[T]) SetInsulator(retryFunc func(ctx context.Context, elem T, err error) (T, error), threshold int, interval time.Duration) {
	oldInsulatorFunc := w.insulatorFunc
	oldRetryInterval := w.retryInterval
	oldRetryThreshold := w.retryThreshold

	w.insulatorFunc = retryFunc
	w.retryInterval = interval
	w.retryThreshold = threshold

	w.NotifyLoggers(
		types.DebugLevel,
		"SetInsulator",
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

func (w *Wire[T]) SetOutputBuffer(b bytes.Buffer) {
	oldOutputBuffer := w.OutputBuffer
	w.OutputBuffer = &b

	w.NotifyLoggers(
		types.DebugLevel,
		"SetOutputBuffer",
		"component", w.componentMetadata,
		"event", "SetOutputBuffer",
		"result", "SUCCESS",
		"oldOutputBufferAddress", oldOutputBuffer,
		"newOutputBufferAddress", w.OutputBuffer,
	)
}

func (w *Wire[T]) SetErrorChannel(errorChan chan types.ElementError[T]) {
	oldErrorChan := w.errorChan
	w.errorChan = errorChan

	w.NotifyLoggers(
		types.DebugLevel,
		"SetErrorChannel",
		"component", w.componentMetadata,
		"event", "SetErrorChannel",
		"result", "SUCCESS",
		"oldErrorChannelAddress", oldErrorChan,
		"newErrorChannelAddress", w.errorChan,
	)
}

func (w *Wire[T]) SetInputChannel(inChan chan T) {
	oldInputChan := w.inChan
	w.inChan = inChan

	w.NotifyLoggers(
		types.DebugLevel,
		"SetInputChannel",
		"component", w.componentMetadata,
		"event", "SetInputChannel",
		"result", "SUCCESS",
		"oldInputChannelAddress", oldInputChan,
		"newInputChannelAddress", w.inChan,
	)
}

func (w *Wire[T]) SetOutputChannel(outChan chan T) {
	oldOutputChan := w.OutputChan
	w.OutputChan = outChan

	w.NotifyLoggers(
		types.DebugLevel,
		"SetOutputChannel",
		"component", w.componentMetadata,
		"event", "SetOutputChannel",
		"result", "SUCCESS",
		"oldOutputChannelAddress", oldOutputChan,
		"newOutputChannelAddress", w.OutputChan,
	)
}

func (w *Wire[T]) Start(_ context.Context) error {
	if !atomic.CompareAndSwapInt32(&w.started, 0, 1) {
		return nil
	}

	w.notifyStart()

	// error handler
	w.wg.Add(1)
	go w.handleErrorElements()

	// resister
	if sp := w.surgeProtector; sp != nil && sp.IsResisterConnected() {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			w.processResisterElements(w.ctx)
		}()
	}

	// ---- transformer config ----
	// Allow "pass-through" wires (no transforms) by injecting identity once.
	// Still forbid mixing factory + explicit transforms.
	if w.transformerFactory != nil {
		if len(w.transformations) != 0 {
			panic("wire: invalid config: transformerFactory set but transformations also present")
		}
	} else {
		if len(w.transformations) == 0 {
			// identity transform: preserves old behavior (wire can be used just for buffering/concurrency/surge/etc)
			w.transformations = []types.Transformer[T]{
				func(v T) (T, error) { return v, nil },
			}
		}
	}

	// fast path selection
	w.fastPathEnabled = w.computeFastPathEnabled()
	if w.fastPathEnabled && w.transformerFactory == nil {
		w.fastTransform = w.transformations[0]
	}

	// workers
	w.transformElements()

	// generators
	for _, g := range w.generators {
		if g != nil && !g.IsStarted() {
			g.Start(w.ctx)
		}
	}

	return nil
}

func (w *Wire[T]) Submit(ctx context.Context, elem T) error {
	if cb := w.CircuitBreaker; cb != nil && !cb.Allow() {
		return w.handleCircuitBreakerTrip(ctx, elem)
	}

	if sp := w.surgeProtector; sp != nil {
		element := types.NewElementFast[T](elem)

		if sp.IsTripped() {
			w.notifySurgeProtectorSubmit(element.Data)
			return sp.Submit(ctx, element)
		}

		if sp.IsBeingRateLimited() && !sp.TryTake() {
			w.notifyRateLimit(elem, time.Now().Add(sp.GetTimeUntilNextRefill()).Format("2006-01-02 15:04:05"))
			return sp.Submit(ctx, element)
		}
	}

	return w.submitNormally(ctx, elem)
}

func (w *Wire[T]) Stop() error {
	if !atomic.CompareAndSwapInt32(&w.started, 1, 0) {
		return nil
	}

	w.notifyStop()

	w.terminateOnce.Do(func() {
		w.closeLock.Lock()
		w.isClosed = true
		w.closeLock.Unlock()

		w.cancel()
		w.wg.Wait()

		select {
		case <-w.completeSignal:
			// If another part of the system uses completeSignal, honor it.
		default:
			w.closeOutputChanOnce.Do(func() { close(w.OutputChan) })
			w.closeErrorChanOnce.Do(func() { close(w.errorChan) })
		}
	})

	return nil
}

func (w *Wire[T]) Restart(ctx context.Context) error {
	_ = w.Stop()

	w.ctx, w.cancel = context.WithCancel(ctx)

	w.terminateOnce = sync.Once{}
	w.closeOutputChanOnce = sync.Once{}
	w.closeInputChanOnce = sync.Once{}
	w.closeErrorChanOnce = sync.Once{}

	w.closeLock.Lock()
	w.isClosed = false
	w.closeLock.Unlock()

	w.SetInputChannel(make(chan T, w.maxBufferSize))
	w.SetOutputChannel(make(chan T, w.maxBufferSize))
	w.SetErrorChannel(make(chan types.ElementError[T], w.maxBufferSize))
	w.OutputBuffer = &bytes.Buffer{}

	if w.CircuitBreaker != nil {
		go w.startCircuitBreakerTicker()
	}

	return w.Start(w.ctx)
}

func (w *Wire[T]) SetTransformerFactory(factory func() types.Transformer[T]) {
	// Config-time only.
	if atomic.LoadInt32(&w.started) == 1 {
		panic("wire: SetTransformerFactory called after Start")
	}

	if factory == nil {
		w.transformerFactory = nil
		return
	}

	// Prevent silent misconfig: factory and transformers are mutually exclusive.
	if len(w.transformations) != 0 {
		panic("wire: cannot use TransformerFactory with ConnectTransformer; wrap transforms inside the factory closure instead")
	}

	w.transformerFactory = factory
}
