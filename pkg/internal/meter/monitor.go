package meter

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Monitor prints periodic snapshots until completion, cancellation, or idle timeout.
func (m *Meter[T]) Monitor() {
	m.startTime = time.Now()
	ticker := m.currentTicker()
	if ticker == nil {
		ticker = time.NewTicker(defaultUpdateInterval)
		defer ticker.Stop()
	}

	for {
		select {
		case <-m.idleTimer.C:
			idleTimeout := m.idleTimeoutValue()
			fmt.Println()
			fmt.Print("\033[2K")
			fmt.Printf("Idled for %ds, shutting down...\n", int(idleTimeout.Seconds()))
			fmt.Println()
			m.printFinalProgress()
			m.cancelContext()
			return
		case <-m.ctx.Done():
			fmt.Println("Process cancelled.")
			m.printFinalProgress()
			return
		case <-ticker.C:
			threshold, hasThreshold := m.thresholdFor(types.MetricTransformationErrorPercentage)
			if hasThreshold {
				totalSubmitted := float64(m.GetMetricCount(types.MetricTotalSubmittedCount))
				errorCount := float64(m.GetMetricCount(types.MetricTransformationErrorPercentage))
				if totalSubmitted > 0 {
					errRatio := (errorCount / totalSubmitted) * 100
					if errRatio >= threshold {
						m.updateDisplay()
						fmt.Println()
						fmt.Printf("\nThreshold exceeded for error_count: %.2f%%\n", threshold)
						m.printFinalProgress()
						m.cancelContext()
						return
					}
				}
			}

			totalItems := m.totalItemsValue()
			if totalItems != 0 && m.GetMetricCount(types.MetricTotalProcessedCount) == totalItems {
				m.updateDisplay()
				if m.GetMetricCount(types.MetricTransformationErrorPercentage) != 0 {
					fmt.Print("\033[2K")
					fmt.Printf("\nFinished with errors\n")
				} else {
					fmt.Print("\033[2K")
					fmt.Printf("\nCompleted all results successfully!\n")
				}
				fmt.Println()

				m.printFinalProgress()
				m.cancelContext()
				return
			}

			m.updateDisplay()
		}
	}
}

// ReportData signals activity to reset the idle timer.
func (m *Meter[T]) ReportData() {
	select {
	case m.dataCh <- struct{}{}:
	default:
	}
}

// CheckMetrics returns true if any metric meets its total or threshold criteria.
func (m *Meter[T]) CheckMetrics() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	for metric, countPtr := range m.counts {
		if countPtr == nil {
			continue
		}
		count := atomic.LoadUint64(countPtr)
		total, hasTotal := m.totals[metric]
		threshold, hasThreshold := m.thresholds[metric]

		if hasTotal && count >= total {
			fmt.Printf("%s completed: %d/%d\n", metric, count, total)
			return true
		}

		if hasThreshold && hasTotal && total > 0 && float64(count)/float64(total)*100 >= threshold {
			fmt.Printf("%s threshold exceeded: %.2f%% (Count: %d, Total: %d)\n", metric, threshold, count, total)
			return true
		}
	}
	return false
}

// PauseProcessing emits a pause signal for consumers that honor it.
func (m *Meter[T]) PauseProcessing() {
	select {
	case m.pauseCh <- struct{}{}:
	default:
	}
}

// ResumeProcessing emits a resume signal for consumers that honor it.
func (m *Meter[T]) ResumeProcessing() {
	select {
	case m.pauseCh <- struct{}{}:
	default:
	}
}

func (m *Meter[T]) cancelContext() {
	m.mu.Lock()
	cancel := m.cancel
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}

func (m *Meter[T]) currentTicker() *time.Ticker {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.ticker
}

func (m *Meter[T]) idleTimeoutValue() time.Duration {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.idleTimeout
}

func (m *Meter[T]) totalItemsValue() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalItems
}

func (m *Meter[T]) thresholdFor(metricName string) (float64, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	threshold, hasThreshold := m.thresholds[metricName]
	return threshold, hasThreshold
}

func (m *Meter[T]) monitorIdleTime(ctx context.Context) {
	defer func() {
		m.mu.Lock()
		if m.idleTimer != nil {
			m.idleTimer.Stop()
		}
		m.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-m.dataCh:
			m.mu.Lock()
			m.resetIdleTimerLocked()
			m.mu.Unlock()
		}
	}
}
