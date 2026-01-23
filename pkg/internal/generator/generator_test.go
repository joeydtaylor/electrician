package generator_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/generator"
	"github.com/joeydtaylor/electrician/pkg/internal/plug"
	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestGenerator_StartStop(t *testing.T) {
	ctx := context.Background()

	g := generator.NewGenerator[int](ctx)
	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if !g.IsStarted() {
		t.Fatalf("expected generator to be started")
	}

	if err := g.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
	if g.IsStarted() {
		t.Fatalf("expected generator to be stopped")
	}
}

func TestGenerator_SubmitsViaPlug(t *testing.T) {
	ctx := context.Background()

	submitCh := make(chan int, 1)
	submitter := &recordingSubmitter[int]{ch: submitCh}

	p := plug.NewPlug[int](ctx,
		plug.WithAdapterFunc[int](func(ctx context.Context, submit func(context.Context, int) error) {
			_ = submit(ctx, 42)
		}),
	)

	g := generator.NewGenerator[int](ctx, generator.WithPlug[int](p))
	g.ConnectToComponent(submitter)

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = g.Stop() }()

	select {
	case got := <-submitCh:
		if got != 42 {
			t.Fatalf("expected 42, got %d", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for submission")
	}
}

func TestGenerator_ConfigurationPanicsAfterStart(t *testing.T) {
	ctx := context.Background()

	g := generator.NewGenerator[int](ctx)
	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = g.Stop() }()

	assertPanics(t, "ConnectCircuitBreaker", func() {
		g.ConnectCircuitBreaker(types.CircuitBreaker[int](nil))
	})
	assertPanics(t, "ConnectLogger", func() {
		g.ConnectLogger(types.Logger(nil))
	})
	assertPanics(t, "ConnectSensor", func() {
		g.ConnectSensor(types.Sensor[int](nil))
	})
	assertPanics(t, "ConnectPlug", func() {
		g.ConnectPlug(types.Plug[int](nil))
	})
	assertPanics(t, "ConnectToComponent", func() {
		g.ConnectToComponent(types.Submitter[int](nil))
	})
	assertPanics(t, "SetComponentMetadata", func() {
		g.SetComponentMetadata("generator", "gen-1")
	})
}

func TestGenerator_StartWithNilContext(t *testing.T) {
	ctx := context.Background()

	g := generator.NewGenerator[int](ctx)
	if err := g.Start(nil); err != nil {
		t.Fatalf("Start(nil) error: %v", err)
	}
	if !g.IsStarted() {
		t.Fatalf("expected generator to be started")
	}

	_ = g.Stop()
}

func TestGenerator_StartWithCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	g := generator.NewGenerator[int](context.Background())
	if err := g.Start(ctx); err == nil {
		t.Fatalf("expected Start to return context error")
	}
	if g.IsStarted() {
		t.Fatalf("expected generator to remain stopped")
	}
}

func TestGenerator_StartTwice(t *testing.T) {
	ctx := context.Background()

	g := generator.NewGenerator[int](ctx)
	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = g.Stop() }()

	if err := g.Start(ctx); err == nil {
		t.Fatalf("expected Start to return error when already started")
	}
}

func TestGenerator_UsesFastSubmit(t *testing.T) {
	ctx := context.Background()

	fast := &fastRecordingSubmitter[int]{}

	p := plug.NewPlug[int](ctx,
		plug.WithAdapterFunc[int](func(ctx context.Context, submit func(context.Context, int) error) {
			_ = submit(ctx, 7)
		}),
	)

	g := generator.NewGenerator[int](ctx, generator.WithPlug[int](p))
	g.ConnectToComponent(fast)

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = g.Stop() }()

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&fast.fastCount) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if atomic.LoadInt32(&fast.fastCount) == 0 {
		t.Fatalf("expected FastSubmit to be called")
	}
	if atomic.LoadInt32(&fast.submitCount) != 0 {
		t.Fatalf("expected Submit not to be called")
	}
}

func TestGenerator_SubmitBlockedByCircuitBreaker(t *testing.T) {
	ctx := context.Background()

	cb := &stubCircuitBreaker[int]{}
	cb.SetAllow(false)

	submitted := make(chan int, 1)
	submitter := &recordingSubmitter[int]{ch: submitted}

	errCh := make(chan error, 1)
	p := plug.NewPlug[int](ctx,
		plug.WithAdapterFunc[int](func(ctx context.Context, submit func(context.Context, int) error) {
			errCh <- submit(ctx, 1)
		}),
	)

	g := generator.NewGenerator[int](ctx,
		generator.WithPlug[int](p),
		generator.WithCircuitBreaker[int](cb),
	)
	g.ConnectToComponent(submitter)

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = g.Stop() }()

	select {
	case err := <-errCh:
		if err == nil {
			t.Fatalf("expected submit to be blocked")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for submit error")
	}

	select {
	case <-submitted:
		t.Fatalf("expected no submissions while breaker is open")
	case <-time.After(200 * time.Millisecond):
	}
}

func TestGenerator_RecordErrorOnSubmitFailure(t *testing.T) {
	ctx := context.Background()

	cb := &stubCircuitBreaker[int]{}
	cb.SetAllow(true)

	submitErr := errors.New("submit failed")
	submitter := &fastRecordingSubmitter[int]{err: submitErr}

	errCh := make(chan error, 1)
	p := plug.NewPlug[int](ctx,
		plug.WithAdapterFunc[int](func(ctx context.Context, submit func(context.Context, int) error) {
			errCh <- submit(ctx, 2)
		}),
	)

	g := generator.NewGenerator[int](ctx,
		generator.WithPlug[int](p),
		generator.WithCircuitBreaker[int](cb),
	)
	g.ConnectToComponent(submitter)

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = g.Stop() }()

	select {
	case err := <-errCh:
		if !errors.Is(err, submitErr) {
			t.Fatalf("expected submit error, got %v", err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timed out waiting for submit error")
	}

	if cb.RecordCount() == 0 {
		t.Fatalf("expected circuit breaker RecordError to be called")
	}
}

func TestGenerator_RestartConnectorsWhenBreakerCloses(t *testing.T) {
	ctx := context.Background()

	cb := &stubCircuitBreaker[int]{}
	cb.SetAllow(false)

	adapter := &stubAdapter[int]{}

	p := plug.NewPlug[int](ctx,
		plug.WithAdapter[int](adapter),
	)

	g := generator.NewGenerator[int](ctx,
		generator.WithPlug[int](p),
		generator.WithCircuitBreaker[int](cb),
	)

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = g.Stop() }()

	time.Sleep(200 * time.Millisecond)
	if adapter.Calls() != 0 {
		t.Fatalf("expected no connector calls while breaker is open")
	}

	cb.SetAllow(true)

	deadline := time.Now().Add(600 * time.Millisecond)
	for time.Now().Before(deadline) {
		if adapter.Calls() > 0 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	t.Fatalf("expected connector to restart after breaker closed")
}

func TestGenerator_NotifyLoggers(t *testing.T) {
	ctx := context.Background()

	logger := &countingLogger{level: types.DebugLevel}
	g := generator.NewGenerator[int](ctx, generator.WithLogger[int](logger))

	g.NotifyLoggers(types.DebugLevel, "debug")
	g.NotifyLoggers(types.InfoLevel, "info")
	g.NotifyLoggers(types.WarnLevel, "warn")
	g.NotifyLoggers(types.ErrorLevel, "error")
	g.NotifyLoggers(types.DPanicLevel, "dpanic")
	g.NotifyLoggers(types.PanicLevel, "panic")
	g.NotifyLoggers(types.FatalLevel, "fatal")

	if atomic.LoadInt32(&logger.debug) != 1 ||
		atomic.LoadInt32(&logger.info) != 1 ||
		atomic.LoadInt32(&logger.warn) != 1 ||
		atomic.LoadInt32(&logger.err) != 1 ||
		atomic.LoadInt32(&logger.dpanic) != 1 ||
		atomic.LoadInt32(&logger.panic) != 1 ||
		atomic.LoadInt32(&logger.fatal) != 1 {
		t.Fatalf("expected one log per level")
	}
}

func TestGenerator_SensorLifecycleEvents(t *testing.T) {
	ctx := context.Background()

	var startCount int32
	var stopCount int32
	var restartCount int32

	s := sensor.NewSensor[int](
		sensor.WithOnStartFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&startCount, 1)
		}),
		sensor.WithOnStopFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&stopCount, 1)
		}),
		sensor.WithOnRestartFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&restartCount, 1)
		}),
	)

	g := generator.NewGenerator[int](ctx, generator.WithSensor[int](s))

	if err := g.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if err := g.Restart(ctx); err != nil {
		t.Fatalf("Restart() error: %v", err)
	}
	if err := g.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	if atomic.LoadInt32(&startCount) != 2 {
		t.Fatalf("expected 2 start events, got %d", startCount)
	}
	if atomic.LoadInt32(&stopCount) != 2 {
		t.Fatalf("expected 2 stop events, got %d", stopCount)
	}
	if atomic.LoadInt32(&restartCount) != 1 {
		t.Fatalf("expected 1 restart event, got %d", restartCount)
	}
}

func TestGenerator_AttachesSensorsToPlugs(t *testing.T) {
	ctx := context.Background()

	p1 := &stubPlug[int]{}
	p2 := &stubPlug[int]{}

	s := sensor.NewSensor[int]()

	g := generator.NewGenerator[int](ctx)
	g.ConnectPlug(p1)
	if p1.SensorCalls() != 0 {
		t.Fatalf("expected no sensor calls before sensor registration")
	}

	g.ConnectSensor(s)
	if p1.SensorCalls() == 0 {
		t.Fatalf("expected sensors to attach to existing plugs")
	}

	g.ConnectPlug(p2)
	if p2.SensorCalls() == 0 {
		t.Fatalf("expected sensors to attach to new plugs")
	}
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic: %s", name)
		}
	}()
	fn()
}

type recordingSubmitter[T any] struct {
	ch      chan T
	started int32
}

func (r *recordingSubmitter[T]) Submit(_ context.Context, elem T) error {
	r.ch <- elem
	return nil
}

func (r *recordingSubmitter[T]) ConnectGenerator(...types.Generator[T]) {}

func (r *recordingSubmitter[T]) GetGenerators() []types.Generator[T] { return nil }

func (r *recordingSubmitter[T]) ConnectLogger(...types.Logger) {}

func (r *recordingSubmitter[T]) Restart(ctx context.Context) error { return r.Start(ctx) }

func (r *recordingSubmitter[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}

func (r *recordingSubmitter[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{Type: "SUBMITTER"}
}

func (r *recordingSubmitter[T]) SetComponentMetadata(name string, id string) {
	_ = name
	_ = id
}

func (r *recordingSubmitter[T]) Stop() error {
	atomic.StoreInt32(&r.started, 0)
	return nil
}

func (r *recordingSubmitter[T]) IsStarted() bool {
	return atomic.LoadInt32(&r.started) == 1
}

func (r *recordingSubmitter[T]) Start(context.Context) error {
	atomic.StoreInt32(&r.started, 1)
	return nil
}

type fastRecordingSubmitter[T any] struct {
	submitCount int32
	fastCount   int32
	err         error
}

func (r *fastRecordingSubmitter[T]) Submit(_ context.Context, _ T) error {
	atomic.AddInt32(&r.submitCount, 1)
	return r.err
}

func (r *fastRecordingSubmitter[T]) FastSubmit(_ context.Context, _ T) error {
	atomic.AddInt32(&r.fastCount, 1)
	return r.err
}

func (r *fastRecordingSubmitter[T]) ConnectGenerator(...types.Generator[T]) {}

func (r *fastRecordingSubmitter[T]) GetGenerators() []types.Generator[T] { return nil }

func (r *fastRecordingSubmitter[T]) ConnectLogger(...types.Logger) {}

func (r *fastRecordingSubmitter[T]) Restart(ctx context.Context) error { return r.Start(ctx) }

func (r *fastRecordingSubmitter[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}

func (r *fastRecordingSubmitter[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{Type: "SUBMITTER"}
}

func (r *fastRecordingSubmitter[T]) SetComponentMetadata(name string, id string) {
	_ = name
	_ = id
}

func (r *fastRecordingSubmitter[T]) Stop() error {
	atomic.StoreInt32(&r.submitCount, 0)
	return nil
}

func (r *fastRecordingSubmitter[T]) IsStarted() bool {
	return atomic.LoadInt32(&r.submitCount) > 0
}

func (r *fastRecordingSubmitter[T]) Start(context.Context) error { return nil }

type stubCircuitBreaker[T any] struct {
	allow     int32
	recordErr int32
}

func (s *stubCircuitBreaker[T]) Allow() bool {
	return atomic.LoadInt32(&s.allow) == 1
}

func (s *stubCircuitBreaker[T]) SetAllow(allow bool) {
	if allow {
		atomic.StoreInt32(&s.allow, 1)
		return
	}
	atomic.StoreInt32(&s.allow, 0)
}

func (s *stubCircuitBreaker[T]) RecordError() {
	atomic.AddInt32(&s.recordErr, 1)
}

func (s *stubCircuitBreaker[T]) RecordCount() int32 {
	return atomic.LoadInt32(&s.recordErr)
}

func (s *stubCircuitBreaker[T]) ConnectSensor(...types.Sensor[T]) {}

func (s *stubCircuitBreaker[T]) SetDebouncePeriod(int) {}

func (s *stubCircuitBreaker[T]) ConnectNeutralWire(...types.Wire[T]) {}

func (s *stubCircuitBreaker[T]) ConnectLogger(types.Logger) {}

func (s *stubCircuitBreaker[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{Type: "CIRCUITBREAKER"}
}

func (s *stubCircuitBreaker[T]) GetNeutralWires() []types.Wire[T] { return nil }

func (s *stubCircuitBreaker[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}

func (s *stubCircuitBreaker[T]) NotifyOnReset() <-chan struct{} { return nil }

func (s *stubCircuitBreaker[T]) Reset() {}

func (s *stubCircuitBreaker[T]) SetComponentMetadata(name string, id string) {
	_ = name
	_ = id
}

func (s *stubCircuitBreaker[T]) Trip() {}

type stubAdapter[T any] struct {
	serveFunc func(context.Context, func(context.Context, T) error) error
	calls     int32
}

func (s *stubAdapter[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	atomic.AddInt32(&s.calls, 1)
	if s.serveFunc != nil {
		return s.serveFunc(ctx, submit)
	}
	return nil
}

func (s *stubAdapter[T]) Calls() int32 {
	return atomic.LoadInt32(&s.calls)
}

func (s *stubAdapter[T]) ConnectSensor(...types.Sensor[T]) {}

func (s *stubAdapter[T]) ConnectLogger(...types.Logger) {}

func (s *stubAdapter[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}

func (s *stubAdapter[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{Type: "ADAPTER"}
}

func (s *stubAdapter[T]) SetComponentMetadata(name string, id string) {
	_ = name
	_ = id
}

type stubPlug[T any] struct {
	adapterFuncs []types.AdapterFunc[T]
	adapters     []types.Adapter[T]
	sensorCalls  int32
	metadata     types.ComponentMetadata
}

func (s *stubPlug[T]) AddAdapterFunc(funcs ...types.AdapterFunc[T]) {
	s.adapterFuncs = append(s.adapterFuncs, funcs...)
}

func (s *stubPlug[T]) GetAdapterFuncs() []types.AdapterFunc[T] {
	return s.adapterFuncs
}

func (s *stubPlug[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	atomic.AddInt32(&s.sensorCalls, int32(len(sensors)))
}

func (s *stubPlug[T]) SensorCalls() int32 {
	return atomic.LoadInt32(&s.sensorCalls)
}

func (s *stubPlug[T]) ConnectLogger(...types.Logger) {}

func (s *stubPlug[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}

func (s *stubPlug[T]) ConnectAdapter(adapters ...types.Adapter[T]) {
	s.adapters = append(s.adapters, adapters...)
}

func (s *stubPlug[T]) GetComponentMetadata() types.ComponentMetadata {
	return s.metadata
}

func (s *stubPlug[T]) SetComponentMetadata(name string, id string) {
	s.metadata = types.ComponentMetadata{Name: name, ID: id}
}

func (s *stubPlug[T]) GetConnectors() []types.Adapter[T] {
	return s.adapters
}

type countingLogger struct {
	level  types.LogLevel
	debug  int32
	info   int32
	warn   int32
	err    int32
	dpanic int32
	panic  int32
	fatal  int32
}

func (c *countingLogger) GetLevel() types.LogLevel { return c.level }

func (c *countingLogger) SetLevel(level types.LogLevel) { c.level = level }

func (c *countingLogger) Debug(string, ...interface{})  { atomic.AddInt32(&c.debug, 1) }
func (c *countingLogger) Info(string, ...interface{})   { atomic.AddInt32(&c.info, 1) }
func (c *countingLogger) Warn(string, ...interface{})   { atomic.AddInt32(&c.warn, 1) }
func (c *countingLogger) Error(string, ...interface{})  { atomic.AddInt32(&c.err, 1) }
func (c *countingLogger) DPanic(string, ...interface{}) { atomic.AddInt32(&c.dpanic, 1) }
func (c *countingLogger) Panic(string, ...interface{})  { atomic.AddInt32(&c.panic, 1) }
func (c *countingLogger) Fatal(string, ...interface{})  { atomic.AddInt32(&c.fatal, 1) }

func (c *countingLogger) Flush() error { return nil }

func (c *countingLogger) AddSink(string, types.SinkConfig) error { return nil }

func (c *countingLogger) RemoveSink(string) error { return nil }

func (c *countingLogger) ListSinks() ([]string, error) { return nil, nil }
