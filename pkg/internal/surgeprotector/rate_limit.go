package surgeprotector

import (
	"sync/atomic"
	"time"
)

// IsBeingRateLimited reports whether rate limiting is enabled.
func (s *SurgeProtector[T]) IsBeingRateLimited() bool {
	return atomic.LoadInt32(&s.rateLimited) == 1
}

// GetTimeUntilNextRefill returns time remaining until the next refill.
func (s *SurgeProtector[T]) GetTimeUntilNextRefill() time.Duration {
	if !s.IsBeingRateLimited() {
		return 0
	}

	s.rateLock.Lock()
	next := s.nextRefill
	s.rateLock.Unlock()

	if next.IsZero() {
		return 0
	}
	now := time.Now()
	if !now.Before(next) {
		return 0
	}
	return next.Sub(now)
}

// ReleaseToken returns a token to the bucket when rate limiting is active.
func (s *SurgeProtector[T]) ReleaseToken() {
	if !s.IsBeingRateLimited() {
		return
	}

	s.rateLock.Lock()
	if s.tokens < s.capacity {
		s.tokens++
	}
	s.rateLock.Unlock()

	s.notifySurgeProtectorReleaseToken()
}

// TryTake attempts to take a token and returns whether it succeeded.
func (s *SurgeProtector[T]) TryTake() bool {
	if !s.IsBeingRateLimited() {
		return true
	}

	s.rateLock.Lock()
	defer s.rateLock.Unlock()

	now := time.Now()
	if s.nextRefill.IsZero() || !now.Before(s.nextRefill) {
		if s.capacity < 0 {
			s.capacity = 0
		}
		s.tokens = s.capacity
		if s.refillRate > 0 {
			s.nextRefill = now.Add(s.refillRate)
		} else {
			s.nextRefill = now
		}
	}

	if s.tokens <= 0 {
		return false
	}
	s.tokens--
	return true
}

// IsResisterConnected reports whether a resister has been attached.
func (s *SurgeProtector[T]) IsResisterConnected() bool {
	return atomic.LoadInt32(&s.resisterEnabled) == 1
}
