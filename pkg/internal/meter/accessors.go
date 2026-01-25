package meter

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// GetComponentMetadata returns the meter metadata.
func (m *Meter[T]) GetComponentMetadata() types.ComponentMetadata {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.componentMetadata
}

// GetDynamicMetricInfo returns the MetricInfo entry for a metric, if present.
func (m *Meter[T]) GetDynamicMetricInfo(metricName string) (*types.MetricInfo, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, exists := m.metrics[metricName]
	return info, exists
}

// GetMetricCount returns the current count for a metric.
func (m *Meter[T]) GetMetricCount(metricName string) uint64 {
	m.mu.Lock()
	counter, exists := m.counts[metricName]
	m.mu.Unlock()
	if !exists || counter == nil {
		return 0
	}
	return atomic.LoadUint64(counter)
}

// GetMetricDisplayName returns the display label for a metric.
func (m *Meter[T]) GetMetricDisplayName(metricName string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, exists := m.metrics[metricName]
	if !exists || info == nil || info.DisplayAs == "" {
		return metricName
	}
	return info.DisplayAs
}

// GetMetricNames returns the ordered metric names list.
func (m *Meter[T]) GetMetricNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.metricNames))
	copy(out, m.metricNames)
	return out
}

// GetMetricPercentage returns the current percentage value for a metric.
func (m *Meter[T]) GetMetricPercentage(metricName string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, exists := m.metrics[metricName]
	if !exists || info == nil {
		return 0
	}
	return info.Percentage
}

// GetMetricPeakPercentage returns the peak percentage value for a metric.
func (m *Meter[T]) GetMetricPeakPercentage(metricName string) float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	info, exists := m.metrics[metricName]
	if !exists || info == nil {
		return 0
	}
	return info.PeakPercentage
}

// GetMetricTotal returns the total for a metric.
func (m *Meter[T]) GetMetricTotal(metricName string) uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	total, exists := m.totals[metricName]
	if !exists {
		return 0
	}
	return total
}

// GetOriginalContext returns the meter's base context.
func (m *Meter[T]) GetOriginalContext() context.Context {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ctx
}

// GetOriginalContextCancel returns the cancel function tied to the meter context.
func (m *Meter[T]) GetOriginalContextCancel() context.CancelFunc {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.cancel
}

// GetTicker returns the ticker used by the monitor loop.
func (m *Meter[T]) GetTicker() *time.Ticker {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ticker
}

// GetTimerStartTime returns the start time of a metric timer, if running.
func (m *Meter[T]) GetTimerStartTime(metricName string) (time.Time, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	startTime, exists := m.startTimes[metricName]
	return startTime, exists
}

// IsTimerRunning reports whether a metric timer is active.
func (m *Meter[T]) IsTimerRunning(metricName string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, running := m.startTimes[metricName]
	return running
}
