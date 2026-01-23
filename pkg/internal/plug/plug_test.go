package plug_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/plug"
	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestPlug_AddAdapterFuncAndConnectors(t *testing.T) {
	ctx := context.Background()

	p := plug.NewPlug[int](ctx).(*plug.Plug[int])

	adapter := &stubAdapter[int]{}
	adapterFn := func(context.Context, func(context.Context, int) error) {}

	p.AddAdapterFunc(nil, adapterFn)
	p.ConnectAdapter(nil, adapter)

	funcs := p.GetAdapterFuncs()
	if len(funcs) != 1 {
		t.Fatalf("expected 1 adapter func, got %d", len(funcs))
	}

	connectors := p.GetConnectors()
	if len(connectors) != 1 {
		t.Fatalf("expected 1 adapter, got %d", len(connectors))
	}
}

func TestPlug_SetComponentMetadata(t *testing.T) {
	ctx := context.Background()

	p := plug.NewPlug[int](ctx).(*plug.Plug[int])
	p.SetComponentMetadata("ingest", "plug-1")

	meta := p.GetComponentMetadata()
	if meta.Name != "ingest" || meta.ID != "plug-1" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}
	if meta.Type != "PLUG" {
		t.Fatalf("expected Type to remain PLUG, got %q", meta.Type)
	}
}

func TestPlug_NotifyLoggers(t *testing.T) {
	ctx := context.Background()

	logger := &countingLogger{level: types.DebugLevel}
	p := plug.NewPlug[int](ctx, plug.WithLogger[int](logger)).(*plug.Plug[int])

	p.NotifyLoggers(types.DebugLevel, "debug")
	p.NotifyLoggers(types.InfoLevel, "info")
	p.NotifyLoggers(types.WarnLevel, "warn")
	p.NotifyLoggers(types.ErrorLevel, "error")
	p.NotifyLoggers(types.DPanicLevel, "dpanic")
	p.NotifyLoggers(types.PanicLevel, "panic")
	p.NotifyLoggers(types.FatalLevel, "fatal")

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

func TestPlug_ConnectSensorLogs(t *testing.T) {
	ctx := context.Background()

	logger := &countingLogger{level: types.DebugLevel}
	p := plug.NewPlug[int](ctx, plug.WithLogger[int](logger)).(*plug.Plug[int])
	s := sensor.NewSensor[int]()

	p.ConnectSensor(s)

	if atomic.LoadInt32(&logger.debug) == 0 {
		t.Fatalf("expected ConnectSensor to log")
	}
}

func TestPlug_ImmutabilityAfterFreeze(t *testing.T) {
	ctx := context.Background()

	p := plug.NewPlug[int](ctx).(*plug.Plug[int])
	p.Freeze()

	assertPanics(t, "AddAdapterFunc", func() {
		p.AddAdapterFunc(func(context.Context, func(context.Context, int) error) {})
	})
	assertPanics(t, "ConnectAdapter", func() {
		p.ConnectAdapter(&stubAdapter[int]{})
	})
	assertPanics(t, "ConnectLogger", func() {
		p.ConnectLogger(&countingLogger{})
	})
	assertPanics(t, "ConnectSensor", func() {
		p.ConnectSensor(sensor.NewSensor[int]())
	})
	assertPanics(t, "SetComponentMetadata", func() {
		p.SetComponentMetadata("name", "id")
	})
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

type stubAdapter[T any] struct{}

func (s *stubAdapter[T]) Serve(context.Context, func(context.Context, T) error) error { return nil }

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
