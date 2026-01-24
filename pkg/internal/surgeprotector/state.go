package surgeprotector

import (
	"sync/atomic"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Trip switches the surge protector to active mode.
func (s *SurgeProtector[T]) Trip() {
	if !atomic.CompareAndSwapInt32(&s.active, 0, 1) {
		return
	}
	s.notifySurgeProtectorTrip()
}

// Reset clears the tripped state and restarts managed components.
func (s *SurgeProtector[T]) Reset() {
	if !atomic.CompareAndSwapInt32(&s.active, 1, 0) {
		return
	}

	s.notifySurgeProtectorReset()

	managed := s.snapshotManagedComponents()
	sensors := s.snapshotSensors()

	for _, mc := range managed {
		if mc == nil {
			continue
		}

		meta := mc.GetComponentMetadata()
		if meta.Type == "CONDUIT" {
			_ = mc.Restart(s.ctx)
		} else if !mc.IsStarted() {
			for _, sensor := range sensors {
				sensor.InvokeOnRestart(meta)
			}
			_ = mc.Start(s.ctx)
		}

		for _, g := range mc.GetGenerators() {
			if g == nil {
				continue
			}
			for _, sensor := range sensors {
				sensor.InvokeOnRestart(g.GetComponentMetadata())
			}
			_ = g.Start(s.ctx)
		}
	}

	s.NotifyLoggers(types.InfoLevel, "Surge protector reset", "component", s.snapshotMetadata())
}

// IsTripped reports whether the surge protector is tripped.
func (s *SurgeProtector[T]) IsTripped() bool {
	return atomic.LoadInt32(&s.active) == 1
}
