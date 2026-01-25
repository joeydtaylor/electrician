package meter

import (
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// AddMetricMonitor registers metrics to display alongside the standard output.
func (m *Meter[T]) AddMetricMonitor(metricInfo ...*types.MetricInfo) (*types.MetricInfo, bool) {
	if len(metricInfo) == 0 {
		return nil, false
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.monitoredMetrics = append(m.monitoredMetrics, metricInfo...)

	var found *types.MetricInfo
	for _, info := range metricInfo {
		if info == nil || info.Name == "" {
			continue
		}
		thresholdType := info.ThresholdType
		if thresholdType == 0 {
			thresholdType = metricThresholdType(info.Name)
		}
		existing := m.registerMetricLocked(info.Name, info.DisplayAs, thresholdType)
		if found == nil && existing != nil {
			found = existing
		}
	}

	if found == nil {
		return nil, false
	}
	return found, true
}

// AddMetric registers a metric for tracking.
func (m *Meter[T]) AddMetric(name string, total uint64, threshold float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := m.registerMetricLocked(name, metricDisplay(name), metricThresholdType(name))
	if info == nil {
		return
	}
	info.Total = total
	info.Threshold = threshold
	m.totals[name] = total
	m.thresholds[name] = threshold
}

// AddToMetricTotal increases a metric's total.
func (m *Meter[T]) AddToMetricTotal(metricName string, additionalTotal uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totals[metricName] += additionalTotal
	if info, exists := m.metrics[metricName]; exists && info != nil {
		info.Total = m.totals[metricName]
	}
}

// AddTotalItems adds to the expected total item count.
func (m *Meter[T]) AddTotalItems(additionalTotal uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totalItems += additionalTotal
}

// DecrementCount decrements a metric count, if it exists.
func (m *Meter[T]) DecrementCount(metricName string) {
	counter := m.ensureCounter(metricName)
	if counter == nil {
		return
	}
	currentValue := atomic.LoadUint64(counter)
	if currentValue == 0 {
		return
	}
	atomic.StoreUint64(counter, currentValue-1)
}

// IncrementCount increments a metric count.
func (m *Meter[T]) IncrementCount(metricName string) {
	counter := m.ensureCounter(metricName)
	if counter == nil {
		return
	}
	atomic.AddUint64(counter, 1)
}

// ResetMetrics clears all counters and totals.
func (m *Meter[T]) ResetMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, counter := range m.counts {
		if counter == nil {
			continue
		}
		atomic.StoreUint64(counter, 0)
	}
	for name := range m.totals {
		m.totals[name] = 0
	}
}

// SetComponentMetadata sets the meter metadata.
func (m *Meter[T]) SetComponentMetadata(name string, id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

// SetContextCancelHook wraps the meter's cancel function with a hook.
func (m *Meter[T]) SetContextCancelHook(hook func()) {
	if hook == nil {
		return
	}
	m.mu.Lock()
	originalCancel := m.cancel
	m.cancel = func() {
		hook()
		if originalCancel != nil {
			originalCancel()
		}
	}
	m.mu.Unlock()
}

// SetDynamicMetric creates or updates a dynamic metric definition.
func (m *Meter[T]) SetDynamicMetric(metricName string, total uint64, initialCount uint64, threshold float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := m.registerMetricLocked(metricName, metricDisplay(metricName), metricThresholdType(metricName))
	if info == nil {
		return
	}
	atomic.StoreUint64(info.Count, initialCount)
	info.Total = total
	info.Threshold = threshold
	m.totals[metricName] = total
	m.thresholds[metricName] = threshold
}

// SetDynamicMetricThreshold updates the threshold for a dynamic metric.
func (m *Meter[T]) SetDynamicMetricThreshold(metricName string, threshold float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thresholds[metricName] = threshold
	if info, exists := m.metrics[metricName]; exists && info != nil {
		info.Threshold = threshold
	}
}

// SetDynamicMetricTotal updates the total for a dynamic metric.
func (m *Meter[T]) SetDynamicMetricTotal(metricName string, total uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.totals[metricName] = total
	if info, exists := m.metrics[metricName]; exists && info != nil {
		info.Total = total
	}
}

// SetIdleTimeout updates the idle timeout duration.
func (m *Meter[T]) SetIdleTimeout(to time.Duration) {
	m.mu.Lock()
	m.idleTimeout = to
	m.resetIdleTimerLocked()
	m.mu.Unlock()
}

// SetMetricCount sets the count for a metric.
func (m *Meter[T]) SetMetricCount(metricName string, count uint64) {
	counter := m.ensureCounter(metricName)
	if counter == nil {
		return
	}
	atomic.StoreUint64(counter, count)
}

// SetMetricPeak updates a metric count only if the new value is higher.
func (m *Meter[T]) SetMetricPeak(metricName string, count uint64) {
	counter := m.ensureCounter(metricName)
	if counter == nil {
		return
	}
	for {
		current := atomic.LoadUint64(counter)
		if count <= current {
			return
		}
		if atomic.CompareAndSwapUint64(counter, current, count) {
			return
		}
	}
}

// SetMetricPeakPercentage updates a metric's peak percentage.
func (m *Meter[T]) SetMetricPeakPercentage(metricName string, percentage float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info := m.registerMetricLocked(metricName, metricDisplay(metricName), types.Percentage)
	if info == nil {
		return
	}
	if percentage > info.PeakPercentage {
		info.PeakPercentage = percentage
	}
}

// SetMetricPercentage sets a metric percentage value.
func (m *Meter[T]) SetMetricPercentage(metricName string, percentage float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info := m.registerMetricLocked(metricName, metricDisplay(metricName), types.Percentage)
	if info == nil {
		return
	}
	info.Percentage = percentage
}

// SetMetricTimestamp sets a metric timestamp.
func (m *Meter[T]) SetMetricTimestamp(metricName string, timestamp int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	info := m.registerMetricLocked(metricName, metricDisplay(metricName), metricThresholdType(metricName))
	if info == nil {
		return
	}
	info.Timestamp = timestamp
}

// SetMetricTotal sets a metric's total.
func (m *Meter[T]) SetMetricTotal(metricName string, total uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setMetricTotalLocked(metricName, total)
}

// SetTicker sets the ticker used by the monitor loop.
func (m *Meter[T]) SetTicker(ticker *time.Ticker) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ticker = ticker
}

// SetTimerStartTime manually sets a metric timer's start time.
func (m *Meter[T]) SetTimerStartTime(metricName string, startTime time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startTimes[metricName] = startTime
}

// SetTotal sets a metric total (alias of SetMetricTotal).
func (m *Meter[T]) SetTotal(metricName string, total uint64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.setMetricTotalLocked(metricName, total)
}

// StartTimer starts a metric timer.
func (m *Meter[T]) StartTimer(metricName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.startTimes[metricName] = time.Now()
}

// StopTimer stops a metric timer and returns the elapsed time.
func (m *Meter[T]) StopTimer(metricName string) time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	startTime, exists := m.startTimes[metricName]
	if !exists {
		return 0
	}
	duration := time.Since(startTime)
	delete(m.startTimes, metricName)
	return duration
}

// UpdateMetric sets a metric count.
func (m *Meter[T]) UpdateMetric(metricName string, count uint64) {
	m.SetMetricCount(metricName, count)
}

// UpdateMetricPercentage updates a metric percentage value.
func (m *Meter[T]) UpdateMetricPercentage(metricName string, percentage float64) {
	m.SetMetricPercentage(metricName, percentage)
}

func (m *Meter[T]) ensureCounter(metricName string) *uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	counter := m.counts[metricName]
	if counter != nil {
		return counter
	}
	info := m.registerMetricLocked(metricName, metricDisplay(metricName), metricThresholdType(metricName))
	if info == nil {
		return nil
	}
	return info.Count
}

func (m *Meter[T]) registerMetricLocked(name string, display string, thresholdType types.ThresholdType) *types.MetricInfo {
	if name == "" {
		return nil
	}
	if display == "" {
		display = name
	}

	counter, exists := m.counts[name]
	if !exists || counter == nil {
		counter = new(uint64)
		m.counts[name] = counter
	}

	info, exists := m.metrics[name]
	if !exists || info == nil {
		info = &types.MetricInfo{
			Name:          name,
			DisplayAs:     display,
			ThresholdType: thresholdType,
			Count:         counter,
			Timestamp:     time.Now().Unix(),
		}
		m.metrics[name] = info
		return info
	}

	if info.Count == nil {
		info.Count = counter
	} else {
		m.counts[name] = info.Count
	}

	if display != "" {
		info.DisplayAs = display
	}
	if thresholdType != info.ThresholdType {
		info.ThresholdType = thresholdType
	}

	return info
}

func (m *Meter[T]) resetIdleTimerLocked() {
	if m.idleTimer == nil {
		m.idleTimer = time.NewTimer(m.idleTimeout)
		return
	}
	if !m.idleTimer.Stop() {
		select {
		case <-m.idleTimer.C:
		default:
		}
	}
	m.idleTimer.Reset(m.idleTimeout)
}

func (m *Meter[T]) setMetricTotalLocked(metricName string, total uint64) {
	info := m.registerMetricLocked(metricName, metricDisplay(metricName), metricThresholdType(metricName))
	if info == nil {
		return
	}
	info.Total = total
	m.totals[metricName] = total
}

func metricDisplay(name string) string {
	if display, ok := metricDisplayNames[name]; ok {
		return display
	}
	return name
}

func metricThresholdType(name string) types.ThresholdType {
	if _, ok := percentageMetrics[name]; ok {
		return types.Percentage
	}
	return types.Absolute
}
