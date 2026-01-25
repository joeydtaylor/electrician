package meter

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithMetricTotal sets an initial total for a metric.
func WithMetricTotal[T any](metricName string, total uint64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetMetricTotal(metricName, total)
	}
}

// WithIdleTimeout updates the idle timeout duration.
func WithIdleTimeout[T any](to time.Duration) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetIdleTimeout(to)
	}
}

// WithInitialMetricCount sets an initial count for a metric.
func WithInitialMetricCount[T any](metricName string, count uint64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetMetricCount(metricName, count)
	}
}

// WithTotalItems sets the expected total items to process.
func WithTotalItems[T any](total uint64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.AddTotalItems(total)
	}
}

// WithTimerStart sets the start time for a metric timer.
func WithTimerStart[T any](metricName string, startTime time.Time) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetTimerStartTime(metricName, startTime)
	}
}

// WithComponentMetadata sets component metadata for the meter.
func WithComponentMetadata[T any](name string, id string) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetComponentMetadata(name, id)
	}
}

// WithLogger attaches loggers to the meter.
func WithLogger[T any](loggers ...types.Logger) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.ConnectLogger(loggers...)
	}
}

// WithDynamicMetric registers a metric with initial values and a threshold.
func WithDynamicMetric[T any](metricName string, total uint64, initialCount uint64, threshold float64) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetDynamicMetric(metricName, total, initialCount, threshold)
	}
}

// WithUpdateFrequency sets the update cadence for Monitor.
func WithUpdateFrequency[T any](duration time.Duration) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.GetTicker().Stop() // Stop the existing ticker
		m.SetTicker(time.NewTicker(duration))
	}
}

// WithCancellationHook adds a custom function to be called on cancellation.
func WithCancellationHook[T any](hook func()) types.Option[types.Meter[T]] {
	return func(m types.Meter[T]) {
		m.SetContextCancelHook(hook)
	}
}
