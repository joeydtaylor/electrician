package surgeprotector

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func WithBackupSystem[T any](bs ...types.Submitter[T]) types.Option[types.SurgeProtector[T]] {
	return func(sp types.SurgeProtector[T]) {
		sp.AttachBackup(bs...)
	}
}

func WithBlackoutPeriod[T any](start, end time.Time) types.Option[types.SurgeProtector[T]] {
	return func(sp types.SurgeProtector[T]) {
		sp.SetBlackoutPeriod(start, end)
	}
}

func WithLogger[T any](logger ...types.Logger) types.Option[types.SurgeProtector[T]] {
	return func(sp types.SurgeProtector[T]) {
		sp.ConnectLogger(logger...)
	}
}

func WithComponentMetadata[T any](name, id string) types.Option[types.SurgeProtector[T]] {
	return func(sp types.SurgeProtector[T]) {
		sp.SetComponentMetadata(name, id)
	}
}

func WithRateLimit[T any](capacity int, refillRate time.Duration, maxRetryAttempts int) types.Option[types.SurgeProtector[T]] {
	return func(sp types.SurgeProtector[T]) {
		sp.SetRateLimit(capacity, refillRate, maxRetryAttempts)
	}
}

func WithResister[T any](r types.Resister[T]) types.Option[types.SurgeProtector[T]] {
	return func(sp types.SurgeProtector[T]) {
		sp.ConnectResister(r)
	}
}

func WithSensor[T any](s ...types.Sensor[T]) types.Option[types.SurgeProtector[T]] {
	return func(sp types.SurgeProtector[T]) {
		sp.ConnectSensor(s...)
	}
}
