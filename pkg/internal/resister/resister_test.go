package resister_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/resister"
	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestResister_PushPopPriorityOrder(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	low := types.NewElementFast(1)
	low.QueuePriority = 1

	high := types.NewElementFast(2)
	high.QueuePriority = 5

	mid := types.NewElementFast(3)
	mid.QueuePriority = 3

	if err := r.Push(low); err != nil {
		t.Fatalf("push low: %v", err)
	}
	if err := r.Push(high); err != nil {
		t.Fatalf("push high: %v", err)
	}
	if err := r.Push(mid); err != nil {
		t.Fatalf("push mid: %v", err)
	}

	if got := r.Pop().Data; got != 2 {
		t.Fatalf("expected high priority first, got %d", got)
	}
	if got := r.Pop().Data; got != 3 {
		t.Fatalf("expected mid priority second, got %d", got)
	}
	if got := r.Pop().Data; got != 1 {
		t.Fatalf("expected low priority third, got %d", got)
	}
}

func TestResister_RequeueUpdatesAndNotifies(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	var queued int32
	var requeued int32
	var dequeued int32
	var empty int32

	s := sensor.NewSensor[int](
		sensor.WithResisterQueuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&queued, 1)
		}),
		sensor.WithResisterRequeuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&requeued, 1)
		}),
		sensor.WithResisterDequeuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&dequeued, 1)
		}),
		sensor.WithResisterEmptyFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&empty, 1)
		}),
	)
	r.ConnectSensor(s)

	element := types.NewElementFast(10)
	element.QueuePriority = 1
	if err := r.Push(element); err != nil {
		t.Fatalf("push element: %v", err)
	}

	updated := types.NewElementFast(20)
	updated.ID = element.ID
	updated.QueuePriority = 7

	if err := r.Push(updated); err != nil {
		t.Fatalf("requeue element: %v", err)
	}

	if got := r.Len(); got != 1 {
		t.Fatalf("expected queue length 1, got %d", got)
	}

	popped := r.Pop()
	if popped == nil {
		t.Fatalf("expected element to be popped")
	}
	if popped.QueuePriority != 7 {
		t.Fatalf("expected priority to be updated, got %d", popped.QueuePriority)
	}
	if popped.Data != 20 {
		t.Fatalf("expected data to be updated, got %d", popped.Data)
	}

	if atomic.LoadInt32(&queued) != 1 {
		t.Fatalf("expected 1 queued notification, got %d", queued)
	}
	if atomic.LoadInt32(&requeued) != 1 {
		t.Fatalf("expected 1 requeued notification, got %d", requeued)
	}
	if atomic.LoadInt32(&dequeued) != 1 {
		t.Fatalf("expected 1 dequeued notification, got %d", dequeued)
	}
	if atomic.LoadInt32(&empty) != 1 {
		t.Fatalf("expected 1 empty notification, got %d", empty)
	}
}

func TestResister_RequeueSamePointer(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	var queued int32
	var requeued int32

	s := sensor.NewSensor[int](
		sensor.WithResisterQueuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&queued, 1)
		}),
		sensor.WithResisterRequeuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&requeued, 1)
		}),
	)
	r.ConnectSensor(s)

	element := types.NewElementFast(1)
	if err := r.Push(element); err != nil {
		t.Fatalf("push element: %v", err)
	}
	if err := r.Push(element); err != nil {
		t.Fatalf("push same element: %v", err)
	}

	if got := r.Len(); got != 1 {
		t.Fatalf("expected queue length 1, got %d", got)
	}
	if atomic.LoadInt32(&queued) != 1 {
		t.Fatalf("expected 1 queued notification, got %d", queued)
	}
	if atomic.LoadInt32(&requeued) != 1 {
		t.Fatalf("expected 1 requeued notification, got %d", requeued)
	}
}

func TestResister_PushAfterPopTreatsAsNew(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	var queued int32
	var requeued int32

	s := sensor.NewSensor[int](
		sensor.WithResisterQueuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&queued, 1)
		}),
		sensor.WithResisterRequeuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&requeued, 1)
		}),
	)
	r.ConnectSensor(s)

	first := types.NewElementFast(5)
	if err := r.Push(first); err != nil {
		t.Fatalf("push first: %v", err)
	}
	_ = r.Pop()

	second := types.NewElementFast(9)
	second.ID = first.ID
	if err := r.Push(second); err != nil {
		t.Fatalf("push second: %v", err)
	}

	if atomic.LoadInt32(&queued) != 2 {
		t.Fatalf("expected 2 queued notifications, got %d", queued)
	}
	if atomic.LoadInt32(&requeued) != 0 {
		t.Fatalf("expected 0 requeued notifications, got %d", requeued)
	}
}

func TestResister_DecayPriorities(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx).(*resister.Resister[int])

	element := types.NewElementFast(5)
	element.QueuePriority = 10
	element.RetryCount = 6

	if err := r.Push(element); err != nil {
		t.Fatalf("push element: %v", err)
	}

	r.DecayPriorities()

	if element.QueuePriority != 8 {
		t.Fatalf("expected priority to decay to 8, got %d", element.QueuePriority)
	}
}

func TestResister_NotifyLoggers(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	logger := &countingLogger{level: types.InfoLevel}
	r.ConnectLogger(logger)

	r.NotifyLoggers(types.DebugLevel, "debug")
	r.NotifyLoggers(types.InfoLevel, "info")

	if got := atomic.LoadInt32(&logger.debug); got != 0 {
		t.Fatalf("expected 0 debug logs, got %d", got)
	}
	if got := atomic.LoadInt32(&logger.info); got != 1 {
		t.Fatalf("expected 1 info log, got %d", got)
	}
}

func TestResister_NotifyLoggers_LevelChecker(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	logger := &levelCheckLogger{
		countingLogger: countingLogger{level: types.DebugLevel},
		enabled:        map[types.LogLevel]bool{types.WarnLevel: true},
	}
	r.ConnectLogger(logger)

	r.NotifyLoggers(types.InfoLevel, "info")
	r.NotifyLoggers(types.WarnLevel, "warn")

	if got := atomic.LoadInt32(&logger.info); got != 0 {
		t.Fatalf("expected info logs to be skipped, got %d", got)
	}
	if got := atomic.LoadInt32(&logger.warn); got != 1 {
		t.Fatalf("expected warn logs to be called once, got %d", got)
	}
}

func TestResister_SetComponentMetadata(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	r.SetComponentMetadata("resister", "id-1")
	meta := r.GetComponentMetadata()
	if meta.Name != "resister" || meta.ID != "id-1" {
		t.Fatalf("expected metadata to be updated, got %+v", meta)
	}
	if meta.Type != "RESISTER" {
		t.Fatalf("expected component type RESISTER, got %q", meta.Type)
	}
}

func TestResister_PushNilReturnsError(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	if err := r.Push(nil); err == nil {
		t.Fatalf("expected error when pushing nil element")
	}
}

func TestResister_WithSensorOption(t *testing.T) {
	ctx := context.Background()

	var queued int32
	s := sensor.NewSensor[int](
		sensor.WithResisterQueuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&queued, 1)
		}),
	)

	r := resister.NewResister[int](ctx, resister.WithSensor[int](s))
	if err := r.Push(types.NewElementFast(1)); err != nil {
		t.Fatalf("push element: %v", err)
	}

	if got := atomic.LoadInt32(&queued); got != 1 {
		t.Fatalf("expected queued callback, got %d", got)
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

type levelCheckLogger struct {
	countingLogger
	enabled map[types.LogLevel]bool
}

func (l *levelCheckLogger) IsLevelEnabled(level types.LogLevel) bool {
	return l.enabled[level]
}

func TestResister_GetResisterQueueLen(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx).(*resister.Resister[int])

	if got := r.GetResisterQueueLen(); got != 0 {
		t.Fatalf("expected empty queue length, got %d", got)
	}

	element := types.NewElementFast(1)
	if err := r.Push(element); err != nil {
		t.Fatalf("push element: %v", err)
	}

	if got := r.GetResisterQueueLen(); got != 1 {
		t.Fatalf("expected queue length 1, got %d", got)
	}
}

func TestResister_PopEmpty(t *testing.T) {
	ctx := context.Background()
	r := resister.NewResister[int](ctx)

	done := make(chan struct{})
	s := sensor.NewSensor[int](
		sensor.WithResisterEmptyFunc[int](func(types.ComponentMetadata) {
			close(done)
		}),
	)
	r.ConnectSensor(s)

	if got := r.Pop(); got != nil {
		t.Fatalf("expected nil pop from empty queue")
	}

	element := types.NewElementFast(7)
	if err := r.Push(element); err != nil {
		t.Fatalf("push element: %v", err)
	}
	_ = r.Pop()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("expected empty notification")
	}
}
