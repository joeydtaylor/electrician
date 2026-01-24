package circuitbreaker_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/circuitbreaker"
	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/wire"
)

func TestCircuitBreaker_DefaultAllow(t *testing.T) {
	ctx := context.Background()
	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, 50*time.Millisecond)

	if !cb.Allow() {
		t.Fatalf("expected breaker to allow by default")
	}

	meta := cb.GetComponentMetadata()
	if meta.ID == "" {
		t.Fatalf("expected component id to be set")
	}
	if meta.Type != "CIRCUIT_BREAKER" {
		t.Fatalf("expected component type to be CIRCUIT_BREAKER, got %q", meta.Type)
	}
}

func TestCircuitBreaker_RecordErrorTripsAndAutoResets(t *testing.T) {
	ctx := context.Background()

	var trips int32
	var resets int32
	var records int32
	s := sensor.NewSensor[int](
		sensor.WithCircuitBreakerTripFunc[int](func(types.ComponentMetadata, int64, int64) {
			atomic.AddInt32(&trips, 1)
		}),
		sensor.WithCircuitBreakerResetFunc[int](func(types.ComponentMetadata, int64) {
			atomic.AddInt32(&resets, 1)
		}),
		sensor.WithCircuitBreakerRecordErrorFunc[int](func(types.ComponentMetadata, int64) {
			atomic.AddInt32(&records, 1)
		}),
	)

	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, 50*time.Millisecond, circuitbreaker.WithSensor[int](s))
	cb.RecordError()

	if cb.Allow() {
		t.Fatalf("expected breaker to be open after error threshold")
	}
	if got := atomic.LoadInt32(&records); got != 1 {
		t.Fatalf("expected 1 recorded error, got %d", got)
	}
	if got := atomic.LoadInt32(&trips); got != 1 {
		t.Fatalf("expected 1 trip, got %d", got)
	}

	time.Sleep(75 * time.Millisecond)
	if !cb.Allow() {
		t.Fatalf("expected breaker to allow after time window")
	}
	if got := atomic.LoadInt32(&resets); got != 1 {
		t.Fatalf("expected 1 reset, got %d", got)
	}
}

func TestCircuitBreaker_Debounce(t *testing.T) {
	ctx := context.Background()

	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 2, 2*time.Second)
	cb.SetDebouncePeriod(1)

	cb.RecordError()
	cb.RecordError()
	if !cb.Allow() {
		t.Fatalf("expected breaker to remain closed due to debounce")
	}

	time.Sleep(1100 * time.Millisecond)
	cb.RecordError()
	if cb.Allow() {
		t.Fatalf("expected breaker to trip after debounce window")
	}
}

func TestCircuitBreaker_NotifyOnReset(t *testing.T) {
	ctx := context.Background()
	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, time.Second)

	cb.Trip()
	resetCh := cb.NotifyOnReset()
	cb.Reset()

	select {
	case <-resetCh:
	case <-time.After(250 * time.Millisecond):
		t.Fatalf("expected reset notification")
	}
}

func TestCircuitBreaker_ConnectNeutralWire(t *testing.T) {
	ctx := context.Background()
	w1 := wire.NewWire[int](ctx)
	w2 := wire.NewWire[int](ctx)

	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, time.Second)
	cb.ConnectNeutralWire(w1, nil, w2)

	wires := cb.GetNeutralWires()
	if len(wires) != 2 {
		t.Fatalf("expected 2 neutral wires, got %d", len(wires))
	}
	if wires[0] != w1 || wires[1] != w2 {
		t.Fatalf("unexpected neutral wires ordering")
	}
}

func TestCircuitBreaker_NotifyLoggers(t *testing.T) {
	ctx := context.Background()
	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, time.Second)

	logger := &countingLogger{level: types.InfoLevel}
	cb.ConnectLogger(logger)

	cb.NotifyLoggers(types.DebugLevel, "debug")
	cb.NotifyLoggers(types.InfoLevel, "info")

	if got := atomic.LoadInt32(&logger.debug); got != 0 {
		t.Fatalf("expected 0 debug logs, got %d", got)
	}
	if got := atomic.LoadInt32(&logger.info); got != 1 {
		t.Fatalf("expected 1 info log, got %d", got)
	}
}

func TestCircuitBreaker_SetComponentMetadata(t *testing.T) {
	ctx := context.Background()
	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, time.Second)

	cb.SetComponentMetadata("breaker", "id-123")
	meta := cb.GetComponentMetadata()
	if meta.Name != "breaker" || meta.ID != "id-123" {
		t.Fatalf("expected metadata to be updated, got %+v", meta)
	}
	if meta.Type != "CIRCUIT_BREAKER" {
		t.Fatalf("expected component type to be CIRCUIT_BREAKER, got %q", meta.Type)
	}
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
