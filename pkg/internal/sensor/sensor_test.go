package sensor_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestSensor_WireCallbacks(t *testing.T) {
	s := sensor.NewSensor[int]()

	var startCount int32
	var submitCount int32
	var processedCount int32
	var errorCount int32
	var cancelCount int32
	var completeCount int32
	var stopCount int32
	var restartCount int32

	s.RegisterOnStart(func(types.ComponentMetadata) { atomic.AddInt32(&startCount, 1) })
	s.RegisterOnSubmit(func(types.ComponentMetadata, int) { atomic.AddInt32(&submitCount, 1) })
	s.RegisterOnElementProcessed(func(types.ComponentMetadata, int) { atomic.AddInt32(&processedCount, 1) })
	s.RegisterOnError(func(types.ComponentMetadata, error, int) { atomic.AddInt32(&errorCount, 1) })
	s.RegisterOnCancel(func(types.ComponentMetadata, int) { atomic.AddInt32(&cancelCount, 1) })
	s.RegisterOnComplete(func(types.ComponentMetadata) { atomic.AddInt32(&completeCount, 1) })
	s.RegisterOnTerminate(func(types.ComponentMetadata) { atomic.AddInt32(&stopCount, 1) })
	s.RegisterOnRestart(func(types.ComponentMetadata) { atomic.AddInt32(&restartCount, 1) })

	meta := types.ComponentMetadata{Type: "WIRE"}
	s.InvokeOnStart(meta)
	s.InvokeOnSubmit(meta, 1)
	s.InvokeOnElementProcessed(meta, 1)
	s.InvokeOnError(meta, errSentinel{}, 1)
	s.InvokeOnCancel(meta, 1)
	s.InvokeOnComplete(meta)
	s.InvokeOnStop(meta)
	s.InvokeOnRestart(meta)

	if startCount != 1 {
		t.Fatalf("expected 1 start callback, got %d", startCount)
	}
	if submitCount != 1 {
		t.Fatalf("expected 1 submit callback, got %d", submitCount)
	}
	if processedCount != 1 {
		t.Fatalf("expected 1 processed callback, got %d", processedCount)
	}
	if errorCount != 1 {
		t.Fatalf("expected 1 error callback, got %d", errorCount)
	}
	if cancelCount != 1 {
		t.Fatalf("expected 1 cancel callback, got %d", cancelCount)
	}
	if completeCount != 1 {
		t.Fatalf("expected 1 complete callback, got %d", completeCount)
	}
	if stopCount != 1 {
		t.Fatalf("expected 1 stop callback, got %d", stopCount)
	}
	if restartCount != 1 {
		t.Fatalf("expected 1 restart callback, got %d", restartCount)
	}
}

func TestSensor_CircuitBreakerCallbacks(t *testing.T) {
	s := sensor.NewSensor[int]()

	var tripCount int32
	var resetCount int32
	var allowCount int32
	var dropCount int32

	s.RegisterOnCircuitBreakerTrip(func(types.ComponentMetadata, int64, int64) { atomic.AddInt32(&tripCount, 1) })
	s.RegisterOnCircuitBreakerReset(func(types.ComponentMetadata, int64) { atomic.AddInt32(&resetCount, 1) })
	s.RegisterOnCircuitBreakerAllow(func(types.ComponentMetadata) { atomic.AddInt32(&allowCount, 1) })
	s.RegisterOnCircuitBreakerDrop(func(types.ComponentMetadata, int) { atomic.AddInt32(&dropCount, 1) })

	meta := types.ComponentMetadata{Type: "CIRCUIT_BREAKER"}
	s.InvokeOnCircuitBreakerTrip(meta, 10, 20)
	s.InvokeOnCircuitBreakerReset(meta, 30)
	s.InvokeOnCircuitBreakerAllow(meta)
	s.InvokeOnCircuitBreakerDrop(meta, 1)

	if tripCount != 1 {
		t.Fatalf("expected 1 trip callback, got %d", tripCount)
	}
	if resetCount != 1 {
		t.Fatalf("expected 1 reset callback, got %d", resetCount)
	}
	if allowCount != 1 {
		t.Fatalf("expected 1 allow callback, got %d", allowCount)
	}
	if dropCount != 1 {
		t.Fatalf("expected 1 drop callback, got %d", dropCount)
	}
}

func TestSensor_NotifyLoggers(t *testing.T) {
	s := sensor.NewSensor[int]()

	logger := &countingLogger{level: types.InfoLevel}
	s.ConnectLogger(logger)

	s.NotifyLoggers(types.DebugLevel, "debug")
	s.NotifyLoggers(types.InfoLevel, "info")

	if got := atomic.LoadInt32(&logger.debug); got != 0 {
		t.Fatalf("expected 0 debug logs, got %d", got)
	}
	if got := atomic.LoadInt32(&logger.info); got != 1 {
		t.Fatalf("expected 1 info log, got %d", got)
	}
}

func TestSensor_MeterDecorators(t *testing.T) {
	s := sensor.NewSensor[int]()
	meter := newStubMeter[int]()
	s.ConnectMeter(meter)

	meta := types.ComponentMetadata{Type: "WIRE"}
	s.InvokeOnStart(meta)
	if got := meter.GetMetricCount(types.MetricComponentRunningCount); got != 1 {
		t.Fatalf("expected running count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricWireRunningCount); got != 1 {
		t.Fatalf("expected wire running count 1, got %d", got)
	}

	s.InvokeOnStop(meta)
	if got := meter.GetMetricCount(types.MetricComponentRunningCount); got != 0 {
		t.Fatalf("expected running count 0, got %d", got)
	}

	s.InvokeOnCircuitBreakerTrip(types.ComponentMetadata{Type: "CIRCUIT_BREAKER"}, 100, 200)
	if meter.GetMetricCount(types.MetricCircuitBreakerTripCount) == 0 {
		t.Fatalf("expected circuit breaker trip count to increment")
	}
	if meter.GetMetricTimestamp(types.MetricCircuitBreakerLastTripTime) != 100 {
		t.Fatalf("expected last trip timestamp to be recorded")
	}
}

type errSentinel struct{}

func (errSentinel) Error() string { return "sentinel" }

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

func (s *stubMeter[T]) GetMetricTimestamp(metricName string) int64 { return s.timestamps[metricName] }

func (s *stubMeter[T]) SetMetricPercentage(name string, percentage float64) {}

func (s *stubMeter[T]) GetMetricNames() []string { return nil }

func (s *stubMeter[T]) ReportData() {}

func (s *stubMeter[T]) SetIdleTimeout(to time.Duration) {}

func (s *stubMeter[T]) SetContextCancelHook(hook func()) {}

func (s *stubMeter[T]) SetDynamicMetric(metricName string, total uint64, initialCount uint64, threshold float64) {
}

func (s *stubMeter[T]) GetDynamicMetricInfo(metricName string) (*types.MetricInfo, bool) {
	return nil, false
}

func (s *stubMeter[T]) SetMetricPeak(metricName string, count uint64) {}

func (s *stubMeter[T]) SetDynamicMetricTotal(metricName string, total uint64) {}

func (s *stubMeter[T]) GetTicker() *time.Ticker { return nil }

func (s *stubMeter[T]) SetTicker(ticker *time.Ticker) {}

func (s *stubMeter[T]) GetOriginalContext() context.Context { return nil }

func (s *stubMeter[T]) GetOriginalContextCancel() context.CancelFunc { return nil }

func (s *stubMeter[T]) Monitor() {}

func (s *stubMeter[T]) AddTotalItems(additionalTotal uint64) {}

func (s *stubMeter[T]) CheckMetrics() bool { return true }

func (s *stubMeter[T]) PauseProcessing() {}

func (s *stubMeter[T]) ResumeProcessing() {}

func (s *stubMeter[T]) GetMetricCount(metricName string) uint64 { return s.counts[metricName] }

func (s *stubMeter[T]) SetMetricCount(metricName string, count uint64) { s.counts[metricName] = count }

func (s *stubMeter[T]) GetMetricTotal(metricName string) uint64 { return 0 }

func (s *stubMeter[T]) SetMetricTotal(metricName string, total uint64) {}

func (s *stubMeter[T]) IncrementCount(metricName string) { s.counts[metricName]++ }

func (s *stubMeter[T]) DecrementCount(metricName string) {
	if s.counts[metricName] > 0 {
		s.counts[metricName]--
	}
}

func (s *stubMeter[T]) IsTimerRunning(metricName string) bool { return false }

func (s *stubMeter[T]) GetTimerStartTime(metricName string) (time.Time, bool) {
	return time.Time{}, false
}

func (s *stubMeter[T]) SetTimerStartTime(metricName string, startTime time.Time) {}

func (s *stubMeter[T]) SetTotal(metricName string, total uint64) {}

func (s *stubMeter[T]) StartTimer(metricName string) {}

func (s *stubMeter[T]) StopTimer(metricName string) time.Duration { return 0 }

func (s *stubMeter[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{}
}

func (s *stubMeter[T]) ConnectLogger(...types.Logger) {}

func (s *stubMeter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {}

func (s *stubMeter[T]) SetComponentMetadata(name string, id string) {}

func (s *stubMeter[T]) ResetMetrics() {}

var _ types.Meter[int] = (*stubMeter[int])(nil)
