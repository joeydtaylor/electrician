package builder

import (
	"context"
	"time"

	sp "github.com/joeydtaylor/electrician/pkg/internal/surgeprotector"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// NewSurgeProtector creates a new instance of a SurgeProtector.
func NewSurgeProtector[T any](ctx context.Context, options ...types.Option[types.SurgeProtector[T]]) types.SurgeProtector[T] {
	return sp.NewSurgeProtector[T](ctx, options...)
}

func SurgeProtectorWithBackupSystem[T any](bs ...types.Submitter[T]) types.Option[types.SurgeProtector[T]] {
	return sp.WithBackupSystem[T](bs...)
}

func SurgeProtectorWithBlackoutPeriod[T any](start, end time.Time) types.Option[types.SurgeProtector[T]] {
	return sp.WithBlackoutPeriod[T](start, end)
}

func SurgeProtectorWithLogger[T any](logger ...types.Logger) types.Option[types.SurgeProtector[T]] {
	return sp.WithLogger[T](logger...)
}

func SurgeProtectorWithRateLimit[T any](capacity int, fillRate time.Duration, maxRetryAttempts int) types.Option[types.SurgeProtector[T]] {
	return sp.WithRateLimit[T](capacity, fillRate, maxRetryAttempts)
}

func SurgeProtectorWithComponentMetadata[T any](name, id string) types.Option[types.SurgeProtector[T]] {
	return sp.WithComponentMetadata[T](name, id)
}

func SurgeProtectorWithResister[T any](r types.Resister[T]) types.Option[types.SurgeProtector[T]] {
	return sp.WithResister[T](r)
}

func SurgeProtectorWithSensor[T any](s ...types.Sensor[T]) types.Option[types.SurgeProtector[T]] {
	return sp.WithSensor[T](s...)
}
