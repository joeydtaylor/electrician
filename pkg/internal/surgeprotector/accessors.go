package surgeprotector

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// GetBackupSystems returns the configured backup submitters.
func (s *SurgeProtector[T]) GetBackupSystems() []types.Submitter[T] {
	s.backupLock.Lock()
	backups := append([]types.Submitter[T](nil), s.backupSystems...)
	s.backupLock.Unlock()
	return backups
}

// GetComponentMetadata returns the surge protector metadata.
func (s *SurgeProtector[T]) GetComponentMetadata() types.ComponentMetadata {
	s.metadataLock.Lock()
	metadata := s.componentMetadata
	s.metadataLock.Unlock()
	return metadata
}

// GetResisterQueue returns the current resister queue length.
func (s *SurgeProtector[T]) GetResisterQueue() int {
	resister := s.snapshotResister()
	if resister == nil {
		return 0
	}
	return resister.Len()
}

// GetRateLimit returns the current rate limit configuration.
func (s *SurgeProtector[T]) GetRateLimit() (int32, time.Duration, int, int) {
	s.rateLock.Lock()
	capacity := s.capacity
	refill := s.refillRate
	maxRetry := s.maxRetryAttempts
	backoff := s.backoffJitter
	s.rateLock.Unlock()
	return capacity, refill, maxRetry, backoff
}
