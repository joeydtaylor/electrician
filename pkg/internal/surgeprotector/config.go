package surgeprotector

import (
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetComponentMetadata updates the component metadata.
func (s *SurgeProtector[T]) SetComponentMetadata(name string, id string) {
	s.metadataLock.Lock()
	s.componentMetadata = types.ComponentMetadata{Name: name, ID: id, Type: s.componentMetadata.Type}
	s.metadataLock.Unlock()
}

// SetBlackoutPeriod trips the surge protector within the specified time window.
func (s *SurgeProtector[T]) SetBlackoutPeriod(start, end time.Time) {
	go func() {
		if sleep := time.Until(start); sleep > 0 {
			time.Sleep(sleep)
		}

		s.Trip()
		s.NotifyLoggers(types.InfoLevel, "Surge protector tripped for blackout period", "component", s.snapshotMetadata())

		if sleep := time.Until(end); sleep > 0 {
			time.Sleep(sleep)
		}

		s.Reset()
		s.NotifyLoggers(types.InfoLevel, "Surge protector reset after blackout period", "component", s.snapshotMetadata())
	}()
}

// WithConditionalBlackout repeatedly checks a condition to trip or reset the protector.
func (s *SurgeProtector[T]) WithConditionalBlackout(check func() bool) {
	if check == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for {
			if check() {
				s.Trip()
			} else {
				s.Reset()
			}
			<-ticker.C
		}
	}()
}

// SetRateLimit configures the token bucket limits.
func (s *SurgeProtector[T]) SetRateLimit(capacity int, refillRate time.Duration, maxRetryAttempts int) {
	if capacity < 0 {
		capacity = 0
	}

	s.rateLock.Lock()
	s.capacity = int32(capacity)
	s.refillRate = refillRate
	s.maxRetryAttempts = maxRetryAttempts
	s.tokens = int32(capacity)
	s.nextRefill = time.Time{}
	s.rateLock.Unlock()

	atomic.StoreInt32(&s.rateLimited, 1)
}
