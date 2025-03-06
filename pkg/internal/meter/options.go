package meter

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithMetricTotal sets an initial total for a specific metric.
func WithMetricTotal[T any](metricName string, total uint64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetMetricTotal(metricName, total)
	}
}

// WithMetricTotal sets an initial total for a specific metric.
func WithIdleTimeout[T any](to time.Duration) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetIdleTimeout(to)
	}
}

// WithMetricTotal sets an initial total for a specific metric.
func WithMetricMonitor[T any](metricInfo ...*types.MetricInfo) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.AddMetricMonitor(metricInfo...)
	}
}

// WithInitialMetricCount sets an initial count for a specific metric.
func WithInitialMetricCount[T any](metricName string, count uint64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetMetricCount(metricName, count)
	}
}

// WithTotalItems sets the initial total number of items to be processed.
func WithTotalItems[T any](total uint64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.AddTotalItems(total)
	}
}

// WithErrorThreshold sets an error threshold percentage for when to take action (e.g., pause or stop processing).
func WithErrorThreshold[T any](threshold float64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetMetricThreshold(types.MetricTransformationErrorPercentage, threshold) // Ensure error threshold is applied to the types.MetricTransformationErrorPercentage metric
	}
}

// WithTimerStart sets the start time for a specified timer metric manually.
func WithTimerStart[T any](metricName string, startTime time.Time) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetTimerStartTime(metricName, startTime)
	}
}

// WithComponentMetadata sets the component metadata for the Meter.
func WithComponentMetadata[T any](name string, id string) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetComponentMetadata(name, id)
	}
}

// WithLogger adds loggers to the Meter for outputting logs.
func WithLogger[T any](loggers ...types.Logger) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.ConnectLogger(loggers...)
	}
}

// WithMetricThreshold sets a threshold for a specific metric that triggers an alert when exceeded.
func WithMetricThreshold[T any](metricName string, threshold float64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetMetricThreshold(metricName, threshold)
	}
}

// WithDynamicMetric adds a new metric to be monitored, with optional initial values.
func WithDynamicMetric[T any](metricName string, total uint64, initialCount uint64, threshold float64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetDynamicMetric(metricName, total, initialCount, threshold)
	}
}

// WithUpdateFrequency sets the frequency at which the Meter updates its readings.
func WithUpdateFrequency[T any](duration time.Duration) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.GetTicker().Stop() // Stop the existing ticker
		m.SetTicker(time.NewTicker(duration))
	}
}

// WithCancellationHook adds a custom function to be called upon cancellation.
func WithCancellationHook[T any](hook func()) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetContextCancelHook(hook)
	}
}
