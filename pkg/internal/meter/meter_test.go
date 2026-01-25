package meter

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type stubLogger struct {
	level       types.LogLevel
	infoCount   int32
	debugCount  int32
	warnCount   int32
	errorCount  int32
	panicCount  int32
	lastMessage string
}

func (s *stubLogger) GetLevel() types.LogLevel {
	return s.level
}

func (s *stubLogger) SetLevel(level types.LogLevel) {
	s.level = level
}

func (s *stubLogger) Debug(msg string, _ ...interface{}) {
	atomic.AddInt32(&s.debugCount, 1)
	s.lastMessage = msg
}

func (s *stubLogger) Info(msg string, _ ...interface{}) {
	atomic.AddInt32(&s.infoCount, 1)
	s.lastMessage = msg
}

func (s *stubLogger) Warn(msg string, _ ...interface{}) {
	atomic.AddInt32(&s.warnCount, 1)
	s.lastMessage = msg
}

func (s *stubLogger) Error(msg string, _ ...interface{}) {
	atomic.AddInt32(&s.errorCount, 1)
	s.lastMessage = msg
}

func (s *stubLogger) DPanic(msg string, _ ...interface{}) {
	atomic.AddInt32(&s.panicCount, 1)
	s.lastMessage = msg
}

func (s *stubLogger) Panic(msg string, _ ...interface{}) {
	atomic.AddInt32(&s.panicCount, 1)
	s.lastMessage = msg
}

func (s *stubLogger) Fatal(msg string, _ ...interface{}) {
	atomic.AddInt32(&s.panicCount, 1)
	s.lastMessage = msg
}

func (s *stubLogger) Flush() error { return nil }

func (s *stubLogger) AddSink(string, types.SinkConfig) error { return nil }

func (s *stubLogger) RemoveSink(string) error { return nil }

func (s *stubLogger) ListSinks() ([]string, error) { return nil, nil }

func TestMetricCountsTotalsAndPeaks(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])

	m.SetMetricCount(types.MetricTotalSubmittedCount, 2)
	m.IncrementCount(types.MetricTotalSubmittedCount)
	m.DecrementCount(types.MetricTotalSubmittedCount)

	if got := m.GetMetricCount(types.MetricTotalSubmittedCount); got != 2 {
		t.Fatalf("expected count 2, got %d", got)
	}

	m.SetMetricTotal(types.MetricTotalSubmittedCount, 10)
	m.AddToMetricTotal(types.MetricTotalSubmittedCount, 5)
	if got := m.GetMetricTotal(types.MetricTotalSubmittedCount); got != 15 {
		t.Fatalf("expected total 15, got %d", got)
	}

	m.SetMetricPeak(types.MetricTotalSubmittedCount, 5)
	m.SetMetricPeak(types.MetricTotalSubmittedCount, 4)
	if got := m.GetMetricCount(types.MetricTotalSubmittedCount); got != 5 {
		t.Fatalf("expected peak 5, got %d", got)
	}

	m.SetMetricPeakPercentage(types.MetricCurrentCpuPercentage, 50)
	m.SetMetricPeakPercentage(types.MetricCurrentCpuPercentage, 25)
	if got := m.GetMetricPeakPercentage(types.MetricCurrentCpuPercentage); got != 50 {
		t.Fatalf("expected peak percentage 50, got %.2f", got)
	}

	m.ResetMetrics()
	if got := m.GetMetricCount(types.MetricTotalSubmittedCount); got != 0 {
		t.Fatalf("expected count reset to 0, got %d", got)
	}
}

func TestDynamicMetricLifecycle(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])

	m.SetDynamicMetric("dynamic", 10, 3, 25)
	if got := m.GetMetricCount("dynamic"); got != 3 {
		t.Fatalf("expected initial count 3, got %d", got)
	}

	info, ok := m.GetDynamicMetricInfo("dynamic")
	if !ok {
		t.Fatal("expected dynamic metric info")
	}
	if info.Total != 10 || info.Threshold != 25 {
		t.Fatalf("unexpected metric info: total=%d threshold=%.2f", info.Total, info.Threshold)
	}

	m.SetDynamicMetricTotal("dynamic", 15)
	if got := m.GetMetricTotal("dynamic"); got != 15 {
		t.Fatalf("expected total 15, got %d", got)
	}

	m.SetDynamicMetricThreshold("dynamic", 50)
	info, _ = m.GetDynamicMetricInfo("dynamic")
	if info.Threshold != 50 {
		t.Fatalf("expected threshold 50, got %.2f", info.Threshold)
	}

	m.UpdateMetric("dynamic", 9)
	if got := m.GetMetricCount("dynamic"); got != 9 {
		t.Fatalf("expected count 9, got %d", got)
	}
}

func TestAddMetricMonitorRegistersMetric(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])

	metric := &types.MetricInfo{Name: "custom_metric", DisplayAs: "Custom Metric"}
	info, ok := m.AddMetricMonitor(metric)
	if !ok || info == nil {
		t.Fatal("expected metric to be registered")
	}

	if got := m.GetMetricDisplayName("custom_metric"); got != "Custom Metric" {
		t.Fatalf("expected display name to match, got %q", got)
	}
}

func TestContextCancelHook(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])

	var called int32
	m.SetContextCancelHook(func() {
		atomic.AddInt32(&called, 1)
	})

	cancel := m.GetOriginalContextCancel()
	cancel()

	if atomic.LoadInt32(&called) != 1 {
		t.Fatalf("expected hook to be called once")
	}

	if err := m.GetOriginalContext().Err(); err == nil {
		t.Fatal("expected context to be canceled")
	}
}

func TestNotifyLoggers(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])
	log := &stubLogger{level: types.InfoLevel}
	m.ConnectLogger(log)

	m.NotifyLoggers(types.InfoLevel, "hello %s", "world")
	if atomic.LoadInt32(&log.infoCount) != 1 {
		t.Fatalf("expected info log")
	}
	if log.lastMessage != "hello world" {
		t.Fatalf("unexpected log message: %q", log.lastMessage)
	}

	m.NotifyLoggers(types.DebugLevel, "debug")
	if atomic.LoadInt32(&log.debugCount) != 0 {
		t.Fatalf("expected debug log to be skipped")
	}
}

func TestMonitorIdleTimeout(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])
	m.SetIdleTimeout(5 * time.Millisecond)

	done := make(chan struct{})
	go func() {
		m.Monitor()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("monitor did not exit on idle timeout")
	}

	if err := m.GetOriginalContext().Err(); err == nil {
		t.Fatal("expected context to be canceled after idle timeout")
	}

	m.GetTicker().Stop()
}

func TestUpdateDisplayDoesNotPanic(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])
	m.SetMetricCount(types.MetricTotalSubmittedCount, 10)
	m.SetMetricCount(types.MetricTotalTransformedCount, 5)
	m.SetMetricCount(types.MetricTotalErrorCount, 1)
	m.startTime = time.Now().Add(-time.Second)

	m.updateDisplay()
}

func TestMetricNamesAreCopied(t *testing.T) {
	m := NewMeter[struct{}](context.Background()).(*Meter[struct{}])
	first := m.GetMetricNames()
	if len(first) == 0 {
		t.Fatal("expected metric names")
	}

	first[0] = "mutated"
	second := m.GetMetricNames()
	if second[0] == "mutated" {
		t.Fatal("expected GetMetricNames to return a copy")
	}
}
