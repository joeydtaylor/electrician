package meter

import (
	"fmt"
	"runtime"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func (m *Meter[T]) updateDisplay() {
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(m.startTime).Seconds()

	totalSubmitted := float64(m.GetMetricCount(types.MetricTotalSubmittedCount))
	totalTransformed := float64(m.GetMetricCount(types.MetricTotalTransformedCount))
	totalErrors := float64(m.GetMetricCount(types.MetricTotalErrorCount))

	transformsPerSecond := 0.0
	errorsPerSecond := 0.0
	if elapsedTime > 0 {
		transformsPerSecond = totalTransformed / elapsedTime
		errorsPerSecond = totalErrors / elapsedTime

		m.SetMetricCount(types.MetricTransformsPerSecond, uint64(transformsPerSecond))
		m.SetMetricCount(types.MetricTransformationErrorsPerSecond, uint64(errorsPerSecond))
	}

	processedPerSecond := transformsPerSecond + errorsPerSecond
	oldPeak := float64(m.GetMetricCount(types.MetricPeakProcessedPerSecond))
	if processedPerSecond > oldPeak {
		m.SetMetricCount(types.MetricPeakProcessedPerSecond, uint64(processedPerSecond))
	}

	if totalSubmitted > 0 {
		cpuPercentages, _ := cpu.Percent(time.Millisecond*500, false)
		memStats, _ := mem.VirtualMemory()

		transformPercentage := (totalTransformed / totalSubmitted) * 100
		errorPercentage := (totalErrors / totalSubmitted) * 100
		backupWirePercentage := (float64(m.GetMetricCount(types.MetricSurgeProtectorBackupWireSubmissionCount)) / totalSubmitted) * 100

		m.SetMetricPercentage(types.MetricTransformPercentage, transformPercentage)
		m.SetMetricPercentage(types.MetricErrorPercentage, errorPercentage)

		if len(cpuPercentages) > 0 {
			m.SetMetricPercentage(types.MetricCurrentCpuPercentage, cpuPercentages[0])
			m.SetMetricPeakPercentage(types.MetricCurrentCpuPercentage, cpuPercentages[0])
		}

		m.SetMetricPercentage(types.MetricCurrentRamPercentage, memStats.UsedPercent)
		m.SetMetricPeakPercentage(types.MetricCurrentRamPercentage, memStats.UsedPercent)

		m.SetMetricCount(types.MetricCurrentGoRoutinesActive, uint64(runtime.NumGoroutine()))
		m.SetMetricPeak(types.MetricPeakGoRoutinesActive, uint64(runtime.NumGoroutine()))

		m.SetMetricPeak(types.MetricPeakTransformedPerSecond, uint64(transformsPerSecond))
		m.SetMetricPeak(types.MetricPeakTransformationErrorsPerSecond, uint64(errorsPerSecond))

		m.SetMetricPercentage(types.MetricSurgeProtectorBackupWireSubmissionPercentage, backupWirePercentage)
	}

	fmt.Printf("\033[2J\033[H")

	fmt.Printf("Start Time: %v, Elapsed Time: %s\n",
		m.startTime.Format("01-02-2006 15:04:05"),
		time.Duration(elapsedTime)*time.Second,
	)

	fmt.Printf("%s: %d, %s: %d, %s: %.2f%%, Peak: %.2f%%, %s: %.2f%%, Peak: %.2f%%\n",
		m.GetMetricDisplayName(types.MetricCurrentGoRoutinesActive),
		m.GetMetricCount(types.MetricCurrentGoRoutinesActive),
		m.GetMetricDisplayName(types.MetricPeakGoRoutinesActive),
		m.GetMetricCount(types.MetricPeakGoRoutinesActive),
		m.GetMetricDisplayName(types.MetricCurrentCpuPercentage),
		m.GetMetricPercentage(types.MetricCurrentCpuPercentage),
		m.GetMetricPeakPercentage(types.MetricCurrentCpuPercentage),
		m.GetMetricDisplayName(types.MetricCurrentRamPercentage),
		m.GetMetricPercentage(types.MetricCurrentRamPercentage),
		m.GetMetricPeakPercentage(types.MetricCurrentRamPercentage),
	)

	totalItems := m.totalItemsValue()
	if totalItems != 0 {
		fmt.Printf("Total Expected Elements: %d, Progress: %d%%\n",
			totalItems,
			int(float64(m.GetMetricCount(types.MetricTotalProcessedCount))/float64(totalItems)*100),
		)
	}

	fmt.Printf("%s: %d, %s: %d, %s: %d\n",
		m.GetMetricDisplayName(types.MetricTotalSubmittedCount),
		m.GetMetricCount(types.MetricTotalSubmittedCount),
		m.GetMetricDisplayName(types.MetricProcessedPerSecond),
		(m.GetMetricCount(types.MetricTransformsPerSecond) + m.GetMetricCount(types.MetricTransformationErrorsPerSecond)),
		m.GetMetricDisplayName(types.MetricPeakProcessedPerSecond),
		m.GetMetricCount(types.MetricPeakProcessedPerSecond),
	)

	fmt.Printf("%s: %d (%.2f%%), %s: %d, %s: %d\n",
		"Transforms",
		m.GetMetricCount(types.MetricTotalTransformedCount),
		m.GetMetricPercentage(types.MetricTransformPercentage),
		m.GetMetricDisplayName(types.MetricTransformsPerSecond),
		m.GetMetricCount(types.MetricTransformsPerSecond),
		m.GetMetricDisplayName(types.MetricPeakTransformedPerSecond),
		m.GetMetricCount(types.MetricPeakTransformedPerSecond),
	)

	fmt.Printf("%s: %d (%.2f%%), %s: %d, %s: %d\n",
		"Transform Errors",
		m.GetMetricCount(types.MetricTotalErrorCount),
		m.GetMetricPercentage(types.MetricErrorPercentage),
		m.GetMetricDisplayName(types.MetricTransformationErrorsPerSecond),
		m.GetMetricCount(types.MetricTransformationErrorsPerSecond),
		m.GetMetricDisplayName(types.MetricPeakTransformationErrorsPerSecond),
		m.GetMetricCount(types.MetricPeakTransformationErrorsPerSecond),
	)

	for _, name := range m.snapshotMetricNames() {
		switch name {
		case types.MetricTransformsPerSecond,
			types.MetricCurrentGoRoutinesActive,
			types.MetricTransformationErrorsPerSecond,
			types.MetricTotalProcessedCount,
			types.MetricTotalSubmittedCount,
			types.MetricTransformPercentage,
			types.MetricTransformationErrorPercentage,
			types.MetricTotalTransformedCount,
			types.MetricPeakProcessedPerSecond,
			types.MetricPeakTransformedPerSecond,
			types.MetricPeakTransformationErrorsPerSecond:
			continue
		case types.MetricSurgeProtectorBackupWireSubmissionCount:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d (%.2f%%)\n",
					m.GetMetricDisplayName(name),
					count,
					m.GetMetricPercentage(types.MetricSurgeProtectorBackupWireSubmissionPercentage),
				)
			}
		default:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d\n", m.GetMetricDisplayName(name), count)
			}
		}
	}

	for _, metricInfo := range m.snapshotMonitoredMetrics() {
		if metricInfo == nil {
			continue
		}
		fmt.Printf("%s: %d\n", metricInfo.DisplayAs, m.GetMetricCount(metricInfo.Name))
	}
}

func (m *Meter[T]) printFinalProgress() {
	m.endTime = time.Now()
	elapsedTime := m.endTime.Sub(m.startTime).Seconds()

	totalSubmitted := m.GetMetricCount(types.MetricTotalSubmittedCount)
	totalErrors := m.GetMetricCount(types.MetricTotalErrorCount)
	totalTransformed := m.GetMetricCount(types.MetricTotalTransformedCount)
	totalPending := totalSubmitted - (totalErrors + totalTransformed)

	var pendingPercentage float64
	if totalSubmitted > 0 {
		pendingPercentage = float64(totalPending) / float64(totalSubmitted) * 100
	}

	m.resetRunningMetrics()

	fmt.Printf("\033[2J\033[H")
	fmt.Printf("Start Time: %v, Elapsed Time: %s\n", m.startTime.Format("01-02-2006 15:04:05"), time.Duration(elapsedTime)*time.Second)
	fmt.Printf("%s: %d, %s: %d, %s: %.2f%%, %s: %.2f%%, %s: %.2f%%, %s: %.2f%%\n",
		"Last Recorded Active Go Routines",
		m.GetMetricCount(types.MetricCurrentGoRoutinesActive),
		"Peak",
		m.GetMetricCount(types.MetricPeakGoRoutinesActive),
		"Last Recorded CPU Percentage",
		m.GetMetricPercentage(types.MetricCurrentCpuPercentage),
		"Peak",
		m.GetMetricPeakPercentage(types.MetricCurrentCpuPercentage),
		"Last Recorded RAM Percentage",
		m.GetMetricPercentage(types.MetricCurrentRamPercentage),
		"Peak",
		m.GetMetricPeakPercentage(types.MetricCurrentRamPercentage),
	)

	totalItems := m.totalItemsValue()
	if totalItems != 0 {
		fmt.Printf("Total Expected Elements: %d, Progress: %d%%, Pending: %d (%.2f%%)\n",
			totalItems,
			int(float64(m.GetMetricCount(types.MetricTotalProcessedCount))/float64(totalItems)*100),
			totalPending,
			pendingPercentage,
		)
	}

	fmt.Printf("%s: %d, Last Recorded %s: %d, Peak %s: %d\n",
		m.GetMetricDisplayName(types.MetricTotalSubmittedCount),
		m.GetMetricCount(types.MetricTotalSubmittedCount),
		m.GetMetricDisplayName(types.MetricProcessedPerSecond),
		(m.GetMetricCount(types.MetricTransformsPerSecond) + m.GetMetricCount(types.MetricTransformationErrorsPerSecond)),
		m.GetMetricDisplayName(types.MetricPeakProcessedPerSecond),
		m.GetMetricCount(types.MetricPeakProcessedPerSecond),
	)

	fmt.Printf("%s: %d (%.2f%%), Last Recorded %s: %d, Peak %s: %d\n",
		"Transform Success",
		m.GetMetricCount(types.MetricTotalTransformedCount),
		m.GetMetricPercentage(types.MetricTransformPercentage),
		m.GetMetricDisplayName(types.MetricTransformsPerSecond),
		m.GetMetricCount(types.MetricTransformsPerSecond),
		m.GetMetricDisplayName(types.MetricPeakTransformedPerSecond),
		m.GetMetricCount(types.MetricPeakTransformedPerSecond),
	)

	fmt.Printf("%s: %d (%.2f%%), Last Recorded %s: %d, Peak %s: %d\n",
		"Transform Errors",
		m.GetMetricCount(types.MetricTotalErrorCount),
		m.GetMetricPercentage(types.MetricErrorPercentage),
		m.GetMetricDisplayName(types.MetricTransformationErrorsPerSecond),
		m.GetMetricCount(types.MetricTransformationErrorsPerSecond),
		m.GetMetricDisplayName(types.MetricPeakTransformationErrorsPerSecond),
		m.GetMetricCount(types.MetricPeakTransformationErrorsPerSecond),
	)

	for _, name := range m.snapshotMetricNames() {
		switch name {
		case types.MetricTotalProcessedCount,
			types.MetricTotalSubmittedCount,
			types.MetricTransformPercentage,
			types.MetricTransformationErrorPercentage:
			continue
		case types.MetricSurgeProtectorBackupWireSubmissionCount:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d (%.2f%%)\n", m.GetMetricDisplayName(name), count, m.GetMetricPercentage(types.MetricSurgeProtectorBackupWireSubmissionPercentage))
			}
		default:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d\n", m.GetMetricDisplayName(name), count)
			}
		}
	}

	for _, metricInfo := range m.snapshotMonitoredMetrics() {
		if metricInfo == nil {
			continue
		}
		fmt.Printf("%s: %d\n", metricInfo.DisplayAs, m.GetMetricCount(metricInfo.Name))
	}

	fmt.Println()
}

func (m *Meter[T]) resetRunningMetrics() {
	m.SetMetricCount(types.MetricComponentRunningCount, 0)
	m.SetMetricCount(types.MetricWireRunningCount, 0)
}

func (m *Meter[T]) snapshotMetricNames() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.metricNames))
	copy(out, m.metricNames)
	return out
}

func (m *Meter[T]) snapshotMonitoredMetrics() []*types.MetricInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.monitoredMetrics) == 0 {
		return nil
	}
	out := make([]*types.MetricInfo, len(m.monitoredMetrics))
	copy(out, m.monitoredMetrics)
	return out
}
