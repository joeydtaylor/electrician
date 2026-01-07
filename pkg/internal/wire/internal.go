package wire

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (w *Wire[T]) attemptRecovery(elem T, originalElem T, originalErr error) (T, error) {
	ins := w.insulatorFunc
	if ins == nil {
		return elem, originalErr
	}

	threshold := w.retryThreshold
	if threshold <= 0 {
		return elem, originalErr
	}
	interval := w.retryInterval

	// Avoid per-attempt allocations from time.After.
	var timer *time.Timer
	if interval > 0 && threshold > 1 {
		timer = time.NewTimer(interval)
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		defer timer.Stop()
	}

	var retryErr error
	for attempt := 1; attempt <= threshold; attempt++ {
		if w.ctx.Err() != nil {
			return elem, originalErr
		}

		retryElem, err := ins(w.ctx, elem, originalErr)
		retryErr = err
		if retryErr == nil {
			w.notifyInsulatorAttemptSuccess(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
			return retryElem, nil
		}

		if attempt == threshold {
			if cb := w.CircuitBreaker; cb != nil {
				cb.RecordError()
			}
			w.notifyInsulatorFinalRetryFailure(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
			return retryElem, fmt.Errorf("retry threshold of %d reached with error: %v", threshold, retryErr)
		}

		w.notifyInsulatorAttempt(retryElem, originalElem, retryErr, originalErr, attempt, threshold, interval)
		elem = retryElem

		if interval > 0 {
			// Wait before the next attempt, cancellable via wire context.
			if timer != nil {
				if !timer.Stop() {
					select {
					case <-timer.C:
					default:
					}
				}
				timer.Reset(interval)
				select {
				case <-timer.C:
				case <-w.ctx.Done():
					return elem, originalErr
				}
			} else {
				// interval>0 but threshold<=1 implies we never actually wait.
			}
		}
	}

	return elem, retryErr
}

func (w *Wire[T]) startCircuitBreakerTicker() {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			w.cbLock.Lock()
			cb := w.CircuitBreaker
			w.cbLock.Unlock()

			allowed := true
			if cb != nil {
				allowed = cb.Allow()
			}

			// Non-blocking publish. Consumer (if any) should poll.
			select {
			case w.controlChan <- allowed:
			default:
			}
		}
	}
}

func (w *Wire[T]) encodeElement(elem T) {
	enc := w.encoder
	if enc == nil {
		return
	}

	w.bufferMutex.Lock()
	err := enc.Encode(w.OutputBuffer, elem)
	w.bufferMutex.Unlock()

	if err != nil {
		w.handleError(elem, err)
	}
}

func (w *Wire[T]) handleCircuitBreakerTrip(ctx context.Context, elem T) error {
	cb := w.CircuitBreaker
	if cb == nil {
		return nil
	}

	neutralWires := cb.GetNeutralWires()
	if len(neutralWires) == 0 {
		w.notifyCircuitBreakerDropElement(elem)
		return nil
	}

	for _, gw := range neutralWires {
		if gw == nil {
			continue
		}
		if err := gw.Submit(ctx, elem); err != nil {
			continue
		}
		w.notifyNeutralWireSubmission(elem)
	}

	return nil
}

func (w *Wire[T]) handleError(elem T, err error) {
	ch := w.errorChan
	if ch == nil {
		return
	}

	// Best-effort: telemetry should not stall the pipeline.
	select {
	case ch <- types.ElementError[T]{Err: err, Elem: elem}:
	default:
	}
}

func (w *Wire[T]) handleErrorElements() {
	defer w.wg.Done()

	// Common case: no telemetry attached. Drain and exit on cancellation.
	if !w.hasLoggers() && !w.hasSensors() {
		ch := w.errorChan
		if ch == nil {
			return
		}
		for {
			select {
			case <-w.ctx.Done():
				return
			case _, ok := <-ch:
				if !ok {
					return
				}
			}
		}
	}

	for {
		select {
		case e, ok := <-w.errorChan:
			if !ok {
				return
			}
			w.notifyElementTransformError(e.Elem, e.Err)
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *Wire[T]) handleProcessingError(elem T, originalElem T, originalErr error) (T, error) {
	if w.insulatorFunc != nil {
		cb := w.CircuitBreaker
		if cb == nil || cb.Allow() {
			return w.attemptRecovery(elem, originalElem, originalErr)
		}
	}

	if cb := w.CircuitBreaker; cb != nil {
		cb.RecordError()
	}
	return elem, originalErr
}

func (w *Wire[T]) transformElement(elem T, telemetryEnabled bool) (T, error) {
	originalElem := elem
	transforms := w.transformations

	switch len(transforms) {
	case 0:
		// No-op.
	case 1:
		var err error
		elem, err = transforms[0](elem)
		if err != nil {
			return w.handleProcessingError(elem, originalElem, err)
		}
	default:
		for i := 0; i < len(transforms); i++ {
			var err error
			elem, err = transforms[i](elem)
			if err != nil {
				return w.handleProcessingError(elem, originalElem, err)
			}
		}
	}

	if telemetryEnabled {
		w.notifyElementProcessed(elem)
	}
	return elem, nil
}

func (w *Wire[T]) processChannelFast() {
	defer w.wg.Done()

	in := w.inChan
	out := w.OutputChan
	done := w.ctx.Done()
	transform := w.fastTransform

	for {
		select {
		case <-done:
			return
		case elem, ok := <-in:
			if !ok {
				return
			}

			processed, err := transform(elem)
			if err != nil {
				// Fast path is only enabled when there is no circuit breaker, insulator,
				// encoder, or per-element telemetry. Best-effort error reporting.
				w.handleError(elem, err)
				continue
			}

			select {
			case out <- processed:
			case <-done:
				return
			}
		}
	}
}

func (w *Wire[T]) transformElements() {
	workers := w.maxConcurrency
	if workers <= 0 {
		workers = 1
	}

	for i := 0; i < workers; i++ {
		w.wg.Add(1)

		if w.fastPathEnabled {
			// Factory fast path: each worker gets its own transformer instance.
			if w.transformerFactory != nil {
				tf := w.transformerFactory()
				if tf == nil {
					panic("wire: transformerFactory returned nil transformer")
				}
				go w.processChannelFastWith(tf)
				continue
			}

			// Static fast path
			go w.processChannelFastWith(w.fastTransform)
			continue
		}

		// Normal path
		go w.processChannel()
	}
}

func (w *Wire[T]) processChannel() {
	defer w.wg.Done()

	in := w.inChan
	done := w.ctx.Done()

	// Treat telemetry as configuration-time. If you attach loggers/sensors after Start,
	// you're outside the supported contract for this type.
	telemetryEnabled := atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0

	for {
		select {
		case <-done:
			return
		case elem, ok := <-in:
			if !ok {
				return
			}
			w.processElementOrDivert(elem, telemetryEnabled)
		}
	}
}

func (w *Wire[T]) processElementOrDivert(elem T, telemetryEnabled bool) {
	if w.ctx.Err() != nil {
		return
	}

	processedElem, err := w.transformElement(elem, telemetryEnabled)

	if w.shouldDivertError(err) {
		w.divertToGroundWires(elem)
		return
	}

	if err != nil {
		w.handleError(elem, err)
		return
	}

	w.submitProcessedElement(processedElem)
}

func (w *Wire[T]) shouldDivertError(err error) bool {
	return err != nil && w.CircuitBreaker != nil && !w.CircuitBreaker.Allow()
}

func (w *Wire[T]) divertToGroundWires(elem T) {
	cb := w.CircuitBreaker
	if cb == nil {
		w.handleError(elem, fmt.Errorf("circuit breaker diversion requested but CircuitBreaker is nil"))
		return
	}

	neutral := cb.GetNeutralWires()
	if len(neutral) == 0 {
		w.handleError(elem, fmt.Errorf("circuit breaker tripped but no ground wires to handle"))
		return
	}

	for _, gw := range neutral {
		if gw == nil {
			continue
		}
		w.notifyNeutralWireSubmission(elem)
		_ = gw.Submit(w.ctx, elem)
	}
}

func (w *Wire[T]) submitProcessedElement(processedElem T) {
	if w.encoder != nil {
		w.encodeElement(processedElem)
	}

	if w.OutputChan == nil {
		return
	}

	select {
	case w.OutputChan <- processedElem:
	case <-w.ctx.Done():
		w.notifyCancel(processedElem)
	}
}

func (w *Wire[T]) processResisterElements(ctx context.Context) {
	sp := w.surgeProtector
	if sp == nil {
		return
	}

	// If not rate-limited, poll at a low cadence.
	if !sp.IsBeingRateLimited() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		w.processItems(ctx, ticker)
		return
	}

	_, fillrate, _, _ := sp.GetRateLimit()
	if fillrate <= 0 {
		// No refill interval given; fall back to a conservative cadence.
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		w.processItems(ctx, ticker)
		return
	}

	ticker := time.NewTicker(fillrate)
	defer ticker.Stop()
	w.processItems(ctx, ticker)
}

func (w *Wire[T]) processItems(ctx context.Context, ticker *time.Ticker) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sp := w.surgeProtector
			if sp == nil {
				return
			}

			if !sp.TryTake() {
				if w.shouldStopProcessing() {
					return
				}
				continue
			}

			item, err := sp.Dequeue()
			if err != nil {
				if w.shouldStopProcessing() {
					return
				}
				continue
			}

			// Do not feed dequeued work back into the surge protector.
			_ = w.submitFromResisterQueue(ctx, item.Data)
		}
	}
}

func (w *Wire[T]) submitFromResisterQueue(ctx context.Context, elem T) error {
	if cb := w.CircuitBreaker; cb != nil && !cb.Allow() {
		return w.handleCircuitBreakerTrip(ctx, elem)
	}
	return w.submitNormally(ctx, elem)
}

func (w *Wire[T]) shouldStopProcessing() bool {
	sp := w.surgeProtector
	if sp == nil {
		return true
	}
	return sp.GetResisterQueue() == 0
}

func (w *Wire[T]) submitNormally(ctx context.Context, elem T) error {
	select {
	case w.inChan <- elem:
		if atomic.LoadInt32(&w.loggerCount) != 0 || atomic.LoadInt32(&w.sensorCount) != 0 {
			w.notifySubmit(elem)
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-w.ctx.Done():
		return w.ctx.Err()
	}
}

func (w *Wire[T]) computeFastPathEnabled() bool {
	// Fast path only when "extras" are completely off and there is exactly 1 transform source.
	// Either:
	//   - transformerFactory != nil and len(transformations)==0
	//   - transformerFactory == nil and len(transformations)==1
	if w.CircuitBreaker != nil ||
		w.surgeProtector != nil ||
		w.insulatorFunc != nil ||
		w.encoder != nil ||
		atomic.LoadInt32(&w.loggerCount) != 0 ||
		atomic.LoadInt32(&w.sensorCount) != 0 {
		return false
	}

	if w.transformerFactory != nil {
		return len(w.transformations) == 0
	}
	return len(w.transformations) == 1
}

func (w *Wire[T]) processChannelFastWith(transform types.Transformer[T]) {
	defer w.wg.Done()

	in := w.inChan
	out := w.OutputChan
	done := w.ctx.Done()

	for {
		select {
		case <-done:
			return
		case elem, ok := <-in:
			if !ok {
				return
			}

			processed, err := transform(elem)
			if err != nil {
				// Preserve semantics: report error (non-blocking on shutdown).
				w.handleError(elem, err)
				continue
			}

			select {
			case out <- processed:
			case <-done:
				// Optional: notify cancel (no Stop() here).
				w.notifyCancel(processed)
				return
			}
		}
	}
}

func (w *Wire[T]) hasLoggers() bool {
	return atomic.LoadInt32(&w.loggerCount) != 0
}

func (w *Wire[T]) hasSensors() bool {
	return atomic.LoadInt32(&w.sensorCount) != 0
}

// NotifyLoggers sends a log event to all connected loggers.
//
// Telemetry is treated as configuration-time wiring. Do not mutate the logger set
// concurrently with processing (e.g., calling ConnectLogger while the wire is running).
func (w *Wire[T]) NotifyLoggers(level types.LogLevel, msg string, keysAndValues ...interface{}) {
	if atomic.LoadInt32(&w.loggerCount) == 0 {
		return
	}

	loggers := w.loggers
	if len(loggers) == 0 {
		return
	}

	type levelChecker interface {
		IsLevelEnabled(types.LogLevel) bool
	}

	switch level {
	case types.DebugLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Debug(msg, keysAndValues...)
		}
	case types.InfoLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Info(msg, keysAndValues...)
		}
	case types.WarnLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Warn(msg, keysAndValues...)
		}
	case types.ErrorLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Error(msg, keysAndValues...)
		}
	case types.DPanicLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.DPanic(msg, keysAndValues...)
		}
	case types.PanicLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Panic(msg, keysAndValues...)
		}
	case types.FatalLevel:
		for _, l := range loggers {
			if l == nil {
				continue
			}
			if lc, ok := l.(levelChecker); ok && !lc.IsLevelEnabled(level) {
				continue
			}
			l.Fatal(msg, keysAndValues...)
		}
	}
}

func (w *Wire[T]) notifyInsulatorFinalRetryFailure(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.InfoLevel,
		"All insulator retries exhausted",
		"component", w.componentMetadata,
		"event", "InsulatorAttempt",
		"result", "FAILURE",
		"totalAttempts", currentAttempt,
		"threshold", maxThreshold,
		"interval_ms", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"currentElement", currentElement,
		"currentErr", currentErr,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnInsulatorFailure(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

func (w *Wire[T]) notifyComplete() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire completed a cycle",
		"component", w.componentMetadata,
		"event", "Complete",
		"result", "SUCCESS",
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnComplete(w.componentMetadata)
	}
}

func (w *Wire[T]) notifyElementProcessed(elem T) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Successfully processed element",
		"component", w.componentMetadata,
		"event", "ElementProcessed",
		"result", "SUCCESS",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnElementProcessed(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyInsulatorAttemptSuccess(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Successfully recovered element",
		"component", w.componentMetadata,
		"event", "InsulatorAttempt",
		"result", "SUCCESS",
		"totalAttempts", currentAttempt,
		"threshold", maxThreshold,
		"interval_ms", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"element", currentElement,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnInsulatorSuccess(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

func (w *Wire[T]) notifyInsulatorAttempt(currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Error during processing; attempting recovery",
		"component", w.componentMetadata,
		"event", "InsulatorAttempt",
		"result", "PENDING",
		"currentAttempt", currentAttempt,
		"threshold", maxThreshold,
		"interval_ms", interval.Milliseconds(),
		"originalElement", originalElement,
		"originalErr", originalErr,
		"currentElement", currentElement,
		"currentErr", currentErr,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnInsulatorAttempt(w.componentMetadata, currentElement, originalElement, currentErr, originalErr, currentAttempt, maxThreshold, interval)
	}
}

func (w *Wire[T]) notifyElementTransformError(elem T, err error) {
	w.NotifyLoggers(
		types.ErrorLevel,
		"Error occurred during element transformation",
		"component", w.componentMetadata,
		"event", "ElementProcessError",
		"result", "FAILURE",
		"element", elem,
		"error", err,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnError(w.componentMetadata, err, elem)
	}
}

func (w *Wire[T]) notifyCancel(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Element processing cancelled due to context cancellation",
		"component", w.componentMetadata,
		"event", "Submit",
		"result", "CANCELLED",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnCancel(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyNeutralWireSubmission(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Submitting element to neutral wire due to circuit breaker trip",
		"component", w.componentMetadata,
		"event", "NeutralWireSubmit",
		"result", "PENDING",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnCircuitBreakerNeutralWireSubmission(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyCircuitBreakerDropElement(elem T) {
	w.NotifyLoggers(
		types.WarnLevel,
		"No neutral wires available and circuit breaker tripped; dropping element",
		"component", w.componentMetadata,
		"event", "CircuitBreakerDropElement",
		"result", "DROPPED",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnCircuitBreakerDrop(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyStart() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire started",
		"component", w.componentMetadata,
		"event", "Start",
		"result", "SUCCESS",
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnStart(w.componentMetadata)
	}
}

func (w *Wire[T]) notifyStop() {
	w.NotifyLoggers(
		types.InfoLevel,
		"Wire stopped",
		"component", w.componentMetadata,
		"event", "Stop",
		"result", "SUCCESS",
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnStop(w.componentMetadata)
	}
}

func (w *Wire[T]) notifySubmit(elem T) {
	w.NotifyLoggers(
		types.InfoLevel,
		"Element submitted",
		"component", w.componentMetadata,
		"event", "Submit",
		"result", "SUCCESS",
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnSubmit(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifySurgeProtectorSubmit(elem T) {
	sp := w.surgeProtector
	meta := types.ComponentMetadata{}
	if sp != nil {
		meta = sp.GetComponentMetadata()
	}

	w.NotifyLoggers(
		types.WarnLevel,
		"Surge protector tripped; submitting element for handling",
		"component", w.componentMetadata,
		"event", "SurgeProtectorSubmit",
		"result", "PENDING",
		"surge_protector", meta,
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnSurgeProtectorSubmit(w.componentMetadata, elem)
	}
}

func (w *Wire[T]) notifyRateLimit(elem T, nextAttempt string) {
	w.NotifyLoggers(
		types.WarnLevel,
		"Hitting rate limit; submission deferred",
		"component", w.componentMetadata,
		"event", "Submit",
		"result", "PENDING",
		"nextAttempt", nextAttempt,
		"element", elem,
	)

	if atomic.LoadInt32(&w.sensorCount) == 0 {
		return
	}
	sensors := w.sensors
	for _, s := range sensors {
		if s == nil {
			continue
		}
		s.InvokeOnSurgeProtectorRateLimitExceeded(w.componentMetadata, elem)
	}
}
