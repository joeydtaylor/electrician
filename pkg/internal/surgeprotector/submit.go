package surgeprotector

import (
	"context"
	"errors"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Submit handles an element based on current surge protection settings.
func (s *SurgeProtector[T]) Submit(ctx context.Context, elem *types.Element[T]) error {
	if elem == nil {
		return nil
	}

	backups := s.snapshotBackups()
	if len(backups) != 0 {
		for _, bs := range backups {
			if bs == nil {
				continue
			}
			if !bs.IsStarted() {
				_ = bs.Start(s.ctx)
			}
			if err := bs.Submit(ctx, elem.Data); err != nil {
				s.notifySurgeProtectorBackupFailure(err)
				return err
			}
			s.notifySurgeProtectorBackupSubmission(elem.Data)
		}
		return nil
	}

	resister := s.snapshotResister()
	if resister == nil || !s.IsResisterConnected() {
		return s.submitToComponents(ctx, elem.Data)
	}

	for resister.Len() > 0 && s.TryTake() {
		nextElem := resister.Pop()
		if nextElem == nil {
			break
		}
		if err := s.submitToComponents(ctx, nextElem.Data); err != nil {
			nextElem.IncrementRetryCount()
			nextElem.AdjustPriority()
			_ = resister.Push(nextElem)
			return err
		}
	}

	if s.TryTake() {
		return s.submitToComponents(ctx, elem.Data)
	}

	_ = resister.Push(elem)
	s.notifySurgeProtectorSubmit(elem.Data)
	return nil
}

// Enqueue adds an element to the resister queue.
func (s *SurgeProtector[T]) Enqueue(element *types.Element[T]) error {
	resister := s.snapshotResister()
	if resister == nil {
		return errors.New("surgeprotector: resister not configured")
	}
	return resister.Push(element)
}

// Dequeue removes the next element from the resister queue.
func (s *SurgeProtector[T]) Dequeue() (*types.Element[T], error) {
	resister := s.snapshotResister()
	if resister == nil {
		return nil, errors.New("surgeprotector: resister not configured")
	}
	if resister.Len() == 0 {
		return nil, errors.New("surgeprotector: queue is empty")
	}

	element := resister.Pop()
	if element == nil {
		return nil, errors.New("surgeprotector: failed to retrieve element")
	}
	return element, nil
}

func (s *SurgeProtector[T]) submitToComponents(ctx context.Context, elem T) error {
	components := s.snapshotManagedComponents()
	for _, component := range components {
		if component == nil {
			continue
		}
		if err := component.Submit(ctx, elem); err != nil {
			return err
		}
	}
	return nil
}
