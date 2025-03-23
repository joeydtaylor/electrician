package meter

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (m *Meter[T]) Monitor() {
	m.startTime = time.Now()
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {

		// 1) Idle Timeout
		case <-m.idleTimer.C:
			fmt.Println()
			fmt.Print("\033[2K")
			fmt.Printf("Idled for %ds, shutting down...\n", int(m.idleTimeout.Seconds()))
			fmt.Println()
			m.printFinalProgress()
			m.cancel()
			return

		// 2) Context Done
		case <-m.ctx.Done():
			fmt.Println("Process cancelled.")
			m.printFinalProgress()
			return

		// 3) Periodic Checks
		case <-ticker.C:

			// ---------------------------------------------------------
			// A) Check error_count threshold as a percentage of total
			// ---------------------------------------------------------
			if threshold, hasThreshold := m.thresholds["error_count"]; hasThreshold {
				totalSubmitted := float64(m.GetMetricCount(types.MetricTotalSubmittedCount))
				errorCount := float64(m.GetMetricCount("error_count"))

				// If totalSubmitted is nonzero, interpret threshold as an error ratio (%)
				if totalSubmitted > 0 {
					errRatio := (errorCount / totalSubmitted) * 100
					if errRatio >= threshold {
						m.updateDisplay()
						fmt.Println()
						fmt.Printf("\nThreshold exceeded for error_count: %.2f%%\n", threshold)
						m.printFinalProgress()
						m.cancel()
						return
					}
				}
			}

			// ---------------------------------------------------------
			// B) If we've processed all items, wrap up
			// ---------------------------------------------------------
			if m.totalItems != 0 && m.GetMetricCount(types.MetricTotalProcessedCount) == m.totalItems {
				m.updateDisplay()

				// If you want to detect “any errors occurred,”
				// you can check "error_count" or a percentage metric, if you prefer
				if m.GetMetricCount("error_count") != 0 {
					fmt.Print("\033[2K")
					fmt.Printf("\nFinished with errors\n")
				} else {
					fmt.Print("\033[2K")
					fmt.Printf("\nCompleted all results successfully!\n")
				}
				fmt.Println()

				m.printFinalProgress()
				m.cancel()
				return
			}

			// ---------------------------------------------------------
			// C) Otherwise, just update the display
			// ---------------------------------------------------------
			m.updateDisplay()
		}
	}
}

func (m *Meter[T]) ReportData() {
	select {
	case m.dataChan <- true: // Attempt to send signal
		// Signal sent successfully, do nothing
	default:
		// Channel is full, do nothing
	}
}

// SetErrorThreshold sets an error threshold for a specific metric.
func (m *Meter[T]) SetErrorThreshold(threshold float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.thresholds["error_count"] = threshold
	if info, exists := m.metrics["error_count"]; exists {
		info.Monitored = true
	}
}

// SetDynamicMetric initializes or updates a dynamic metric with the specified values.
func (m *Meter[T]) SetDynamicMetric(metricName string, total uint64, initialCount uint64, threshold float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, exists := m.metrics[metricName]; !exists {
		count := uint64(initialCount)
		m.metrics[metricName] = &types.MetricInfo{
			Count:     &count,
			Total:     total,
			Threshold: threshold,
		}
	} else {
		info := m.metrics[metricName]
		info.Total = total
		if info.Count != nil {
			atomic.StoreUint64(info.Count, initialCount)
		}
		info.Threshold = threshold
	}
}

// GetDynamicMetricInfo retrieves information about a dynamic metric.
func (m *Meter[T]) AddMetricMonitor(metricInfo ...*types.MetricInfo) (*types.MetricInfo, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.monitoredMetrics = append(m.monitoredMetrics, metricInfo...)
	for _, mn := range m.monitoredMetrics {
		info, exists := m.metrics[mn.Name]
		return info, exists
	}
	return nil, false
}

// GetDynamicMetricInfo retrieves information about a dynamic metric.
func (m *Meter[T]) GetDynamicMetricInfo(metricName string) (*types.MetricInfo, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	info, exists := m.metrics[metricName]
	return info, exists
}

// GetDynamicMetricInfo retrieves information about a dynamic metric.
func (m *Meter[T]) GetMetricDisplayName(metricName string) string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	info, exists := m.metrics[metricName]
	if exists {
		return info.DisplayAs
	}
	return ""
}

// SetDynamicMetricTotal sets the total for a specified dynamic metric.
func (m *Meter[T]) SetDynamicMetricTotal(metricName string, total uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.metrics[metricName]; exists {
		if total > info.Total {
			info.Total = total
		}
	}
}

// SetDynamicMetricTotal sets the total for a specified dynamic metric.
func (m *Meter[T]) SetMetricTimestamp(metricName string, time int64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.metrics[metricName]; exists {
		info.Timestamp = time
	}
}

func (m *Meter[T]) SetMetricPeak(metricName string, count uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// Check if the metric already exists
	if current, exists := m.counts[metricName]; exists {
		// If it exists, update it only if the new count is higher
		if count > *current {
			*current = count
		}
	} else {
		// If it doesn't exist, initialize it with the new count
		m.counts[metricName] = new(uint64)
		*m.counts[metricName] = count
	}
}

// SetMetricPeak sets the count for a specified dynamic metric.
func (m *Meter[T]) SetIdleTimeout(to time.Duration) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.idleTimeout = to
	m.idleTimer = time.NewTimer(to)
}

// SetDynamicMetricThreshold sets the threshold for a specified dynamic metric.
func (m *Meter[T]) SetDynamicMetricThreshold(metricName string, threshold float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.metrics[metricName]; exists {
		info.Threshold = threshold
	}
}

func (m *Meter[T]) GetTicker() *time.Ticker {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.ticker
}

func (m *Meter[T]) GetOriginalContext() context.Context {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.ctx
}

func (m *Meter[T]) GetOriginalContextCancel() context.CancelFunc {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.cancel
}

func (m *Meter[T]) SetContextCancelHook(hook func()) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	originalCancel := m.GetOriginalContextCancel()
	m.cancel = func() {
		hook()           // Call the custom hook
		originalCancel() // Ensure the original cancellation is still processed
	}
}

func (m *Meter[T]) SetTicker(ticker *time.Ticker) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.ticker = ticker
}

func (m *Meter[T]) AddMetric(name string, total uint64, threshold float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, exists := m.metrics[name]; !exists {
		count := uint64(0)
		m.metrics[name] = &types.MetricInfo{
			Count:     &count,
			Total:     total,
			Threshold: threshold,
		}
	}
}

func (m *Meter[T]) UpdateMetric(name string, count uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.metrics[name]; exists {
		atomic.StoreUint64(info.Count, count)
	}
}

func (m *Meter[T]) UpdateMetricPercentage(name string, percentage float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.metrics[name]; exists {
		info.Percentage = percentage
	}
}

func (m *Meter[T]) SetMetricTotal(name string, total uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.metrics[name]; exists {
		info.Total = total
	}
}

func (m *Meter[T]) SetMetricPercentage(name string, percentage float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if info, exists := m.metrics[name]; exists {
		info.Percentage = percentage
	}
}

func (m *Meter[T]) SetMetricPeakPercentage(name string, percentage float64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	// Check if the metric already exists
	// Set the current percentage

	if info, exists := m.metrics[name]; exists {

		if percentage > info.PeakPercentage {
			info.PeakPercentage = percentage
		}

	}

}

// GetMetricTotal retrieves the total for a specified metric.
func (m *Meter[T]) GetMetricTotal(metricName string) uint64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	total, exists := m.totals[metricName]
	if !exists {
		return 0 // Return 0 if the metric does not exist
	}
	return total
}

func (m *Meter[T]) GetMetricNames() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.metricNames
}

func (m *Meter[T]) CheckMetrics() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for metric, countPtr := range m.counts {
		count := atomic.LoadUint64(countPtr)
		total, hasTotal := m.totals[metric]
		threshold, hasThreshold := m.thresholds[metric]

		// Debugging outputs to trace metric checks

		// Check if the total count completion condition is met
		if hasTotal && count >= total {
			fmt.Printf("%s completed: %d/%d\n", metric, count, total)
			return true
		}

		// Check if the threshold condition is met (only if total is specified and non-zero to avoid division by zero)
		if hasThreshold && hasTotal && total > 0 && float64(count)/float64(total)*100 >= threshold {
			fmt.Printf("%s threshold exceeded: %.2f%% (Count: %d, Total: %d)\n", metric, threshold, count, total)
			return true
		}
	}
	return false
}

func (m *Meter[T]) PauseProcessing() {
	m.pauseChan <- true
}

func (m *Meter[T]) ResumeProcessing() {
	m.pauseChan <- true
}

// GetMetricCount retrieves the current count for a specified metric.
func (m *Meter[T]) GetMetricCount(metricName string) uint64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if count, exists := m.counts[metricName]; exists {
		return atomic.LoadUint64(count)
	}
	return 0
}

// SetMetricCount sets the count for a specified metric.
func (m *Meter[T]) SetMetricCount(metricName string, count uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if _, exists := m.counts[metricName]; !exists {
		m.counts[metricName] = new(uint64)
	}
	atomic.StoreUint64(m.counts[metricName], count)
}

// AddTotalItems adds to the total count of items to be processed.
func (m *Meter[T]) AddTotalItems(additionalTotal uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totalItems += additionalTotal
}

// GetMetricTotal retrieves the total for a specified metric.
func (m *Meter[T]) GetMetricPercentage(metricName string) float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	total, exists := m.metrics[metricName]
	if !exists {
		return 0 // Return 0 if the metric does not exist
	}
	return total.Percentage
}

// GetMetricTotal retrieves the total for a specified metric.
func (m *Meter[T]) GetMetricPeakPercentage(metricName string) float64 {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	total, exists := m.metrics[metricName]
	if !exists {
		return 0 // Return 0 if the metric does not exist
	}
	return total.PeakPercentage
}

// AddToMetricTotal adds to the total count for a specified metric.
func (m *Meter[T]) AddToMetricTotal(metricName string, additionalTotal uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totals[metricName] += additionalTotal
}

// ResetMetrics clears all counters and totals for all metrics.
func (m *Meter[T]) ResetMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	for k := range m.counts {
		m.counts[k] = new(uint64) // Reset counters to zero
	}
	for k := range m.totals {
		m.totals[k] = 0 // Reset totals to zero
	}
}

// IsTimerRunning checks if the timer for a specified metric is running.
func (m *Meter[T]) IsTimerRunning(metricName string) bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	_, running := m.startTimes[metricName]
	return running
}

// GetTimerStartTime returns the start time of the timer for the specified metric.
func (m *Meter[T]) GetTimerStartTime(metricName string) (time.Time, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	startTime, exists := m.startTimes[metricName]
	return startTime, exists
}

// SetTimerStartTime manually sets the start time for a specified metric's timer.
func (m *Meter[T]) SetTimerStartTime(metricName string, startTime time.Time) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.startTimes[metricName] = startTime
}

func (m *Meter[T]) IncrementCount(metricName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if counter, exists := m.counts[metricName]; exists {
		atomic.AddUint64(counter, 1)
	}
}

func (m *Meter[T]) DecrementCount(metricName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if counter, exists := m.counts[metricName]; exists {
		currentValue := atomic.LoadUint64(counter)
		if currentValue > 0 {
			newValue := currentValue - 1
			atomic.StoreUint64(counter, newValue)
		}
	}
}

func (m *Meter[T]) SetTotal(metricName string, total uint64) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.totals[metricName] = total
}

func (m *Meter[T]) StartTimer(metricName string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.startTimes[metricName] = time.Now()
}

func (m *Meter[T]) StopTimer(metricName string) time.Duration {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	startTime, exists := m.startTimes[metricName]
	if !exists {
		return 0
	}
	duration := time.Since(startTime)
	delete(m.startTimes, metricName)
	return duration
}

// GetComponentMetadata returns the metadata.
func (m *Meter[T]) GetComponentMetadata() types.ComponentMetadata {
	return m.componentMetadata
}

// SetComponentMetadata sets the component metadata.
func (m *Meter[T]) SetComponentMetadata(name string, id string) {
	m.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

func (m *Meter[T]) ConnectLogger(l ...types.Logger) {
	m.loggersLock.Lock()
	defer m.loggersLock.Unlock()
	m.loggers = append(m.loggers, l...)
}

func (m *Meter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if m.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range m.loggers {
			if logger == nil {
				continue
			}
			// Ensure we only acquire the lock once per logger to avoid deadlock or excessive locking overhead
			m.loggersLock.Lock()
			if logger.GetLevel() <= level {
				switch level {
				case types.DebugLevel:
					logger.Debug(msg)
				case types.InfoLevel:
					logger.Info(msg)
				case types.WarnLevel:
					logger.Warn(msg)
				case types.ErrorLevel:
					logger.Error(msg)
				case types.DPanicLevel:
					logger.DPanic(msg)
				case types.PanicLevel:
					logger.Panic(msg)
				case types.FatalLevel:
					logger.Fatal(msg)
				}
			}
			m.loggersLock.Unlock()
		}
	}
}
