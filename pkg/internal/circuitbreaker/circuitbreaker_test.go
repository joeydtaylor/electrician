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

	wires[0] = nil
	if cb.GetNeutralWires()[0] == nil {
		t.Fatalf("expected neutral wires slice to be a copy")
	}
}

func TestCircuitBreaker_AllowAutoResetImmediate(t *testing.T) {
	ctx := context.Background()

	var trips int32
	var resets int32
	s := sensor.NewSensor[int](
		sensor.WithCircuitBreakerTripFunc[int](func(types.ComponentMetadata, int64, int64) {
			atomic.AddInt32(&trips, 1)
		}),
		sensor.WithCircuitBreakerResetFunc[int](func(types.ComponentMetadata, int64) {
			atomic.AddInt32(&resets, 1)
		}),
	)

	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, 0, circuitbreaker.WithSensor[int](s))
	cb.RecordError()
	if atomic.LoadInt32(&trips) != 1 {
		t.Fatalf("expected 1 trip notification, got %d", trips)
	}

	if !cb.Allow() {
		t.Fatalf("expected breaker to auto-reset with zero window")
	}
	if atomic.LoadInt32(&resets) != 1 {
		t.Fatalf("expected 1 reset notification, got %d", resets)
	}
}

func TestCircuitBreaker_AllowNotifiesSensor(t *testing.T) {
	ctx := context.Background()

	var allows int32
	s := sensor.NewSensor[int](
		sensor.WithCircuitBreakerAllowFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&allows, 1)
		}),
	)

	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 5, time.Second, circuitbreaker.WithSensor[int](s))
	if !cb.Allow() {
		t.Fatalf("expected breaker to allow")
	}
	if atomic.LoadInt32(&allows) == 0 {
		t.Fatalf("expected allow notification")
	}
}

func TestCircuitBreaker_DebounceDisabled(t *testing.T) {
	ctx := context.Background()

	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 2, time.Second)
	cb.SetDebouncePeriod(0)

	cb.RecordError()
	cb.RecordError()
	if cb.Allow() {
		t.Fatalf("expected breaker to trip without debounce")
	}
}

func TestCircuitBreaker_ResetNoopWhenAlreadyAllowed(t *testing.T) {
	ctx := context.Background()
	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, time.Second)

	resetCh := cb.NotifyOnReset()
	cb.Reset()

	select {
	case <-resetCh:
		t.Fatalf("did not expect reset notification when already allowed")
	default:
	}
}

func TestCircuitBreaker_SensorCreationIncrementsMeter(t *testing.T) {
	ctx := context.Background()

	meter := newStubMeter[int]()
	s := sensor.NewSensor[int]()
	s.ConnectMeter(meter)

	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, time.Second)
	cb.ConnectSensor(s)

	if got := meter.GetMetricCount(types.MetricCircuitBreakerCount); got != 1 {
		t.Fatalf("expected circuit breaker count 1, got %d", got)
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

func TestCircuitBreaker_NotifyLoggers_LevelChecker(t *testing.T) {
	ctx := context.Background()
	cb := circuitbreaker.NewCircuitBreaker[int](ctx, 1, time.Second)

	logger := &levelCheckLogger{
		countingLogger: countingLogger{level: types.DebugLevel},
		enabled:        map[types.LogLevel]bool{types.WarnLevel: true},
	}
	cb.ConnectLogger(logger)

	cb.NotifyLoggers(types.InfoLevel, "info")
	cb.NotifyLoggers(types.WarnLevel, "warn")

	if got := atomic.LoadInt32(&logger.info); got != 0 {
		t.Fatalf("expected info logs to be skipped, got %d", got)
	}
	if got := atomic.LoadInt32(&logger.warn); got != 1 {
		t.Fatalf("expected warn logs to be called once, got %d", got)
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

type levelCheckLogger struct {
	countingLogger
	enabled map[types.LogLevel]bool
}

func (l *levelCheckLogger) IsLevelEnabled(level types.LogLevel) bool {
	return l.enabled[level]
}

type stubMeter[T any] struct {
	counts     map[string]uint64
	timestamps map[string]int64
}

func newStubMeter[T any]() *stubMeter[T] {
	return &stubMeter[T]{
		counts:     make(map[string]uint64),
		timestamps: make(map[string]int64),
	}
}

func (s *stubMeter[T]) GetMetricDisplayName(metricName string) string { return metricName }
func (s *stubMeter[T]) GetMetricPercentage(metricName string) float64 { return 0 }
func (s *stubMeter[T]) SetMetricTimestamp(metricName string, time int64) {
	s.timestamps[metricName] = time
}
func (s *stubMeter[T]) GetMetricTimestamp(metricName string) int64          { return s.timestamps[metricName] }
func (s *stubMeter[T]) SetMetricPercentage(name string, percentage float64) {}
func (s *stubMeter[T]) GetMetricNames() []string                            { return nil }
func (s *stubMeter[T]) ReportData()                                         {}
func (s *stubMeter[T]) SetIdleTimeout(to time.Duration)                     {}
func (s *stubMeter[T]) SetContextCancelHook(hook func())                    {}
func (s *stubMeter[T]) SetDynamicMetric(metricName string, total uint64, initialCount uint64, threshold float64) {
}
func (s *stubMeter[T]) GetDynamicMetricInfo(metricName string) (*types.MetricInfo, bool) {
	return nil, false
}
func (s *stubMeter[T]) SetMetricPeak(metricName string, count uint64) {}
func (s *stubMeter[T]) SetDynamicMetricTotal(metricName string, total uint64) {
}
func (s *stubMeter[T]) GetTicker() *time.Ticker                        { return nil }
func (s *stubMeter[T]) SetTicker(ticker *time.Ticker)                  {}
func (s *stubMeter[T]) GetOriginalContext() context.Context            { return nil }
func (s *stubMeter[T]) GetOriginalContextCancel() context.CancelFunc   { return nil }
func (s *stubMeter[T]) Monitor()                                       {}
func (s *stubMeter[T]) AddTotalItems(additionalTotal uint64)           {}
func (s *stubMeter[T]) CheckMetrics() bool                             { return true }
func (s *stubMeter[T]) PauseProcessing()                               {}
func (s *stubMeter[T]) ResumeProcessing()                              {}
func (s *stubMeter[T]) GetMetricCount(metricName string) uint64        { return s.counts[metricName] }
func (s *stubMeter[T]) SetMetricCount(metricName string, count uint64) { s.counts[metricName] = count }
func (s *stubMeter[T]) GetMetricTotal(metricName string) uint64        { return 0 }
func (s *stubMeter[T]) SetMetricTotal(metricName string, total uint64) {}
func (s *stubMeter[T]) IncrementCount(metricName string)               { s.counts[metricName]++ }
func (s *stubMeter[T]) DecrementCount(metricName string)               { s.counts[metricName]-- }
func (s *stubMeter[T]) IsTimerRunning(metricName string) bool          { return false }
func (s *stubMeter[T]) GetTimerStartTime(metricName string) (time.Time, bool) {
	return time.Time{}, false
}
func (s *stubMeter[T]) SetTimerStartTime(metricName string, startTime time.Time) {}
func (s *stubMeter[T]) SetTotal(metricName string, total uint64)                 {}
func (s *stubMeter[T]) StartTimer(metricName string)                             {}
func (s *stubMeter[T]) StopTimer(metricName string) time.Duration                { return 0 }
func (s *stubMeter[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{}
}
func (s *stubMeter[T]) ConnectLogger(...types.Logger) {}
func (s *stubMeter[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {
}
func (s *stubMeter[T]) SetComponentMetadata(name string, id string) {}
func (s *stubMeter[T]) ResetMetrics()                               {}
