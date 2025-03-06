package surgeprotector

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Implement scheduled or conditional activation
func (s *SurgeProtector[T]) ConnectComponent(c ...types.Submitter[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.managedComponents = append(s.managedComponents, c...)
}

// SetBlackoutPeriod configures the surge protector to trip during a specified period and reset afterwards.
func (s *SurgeProtector[T]) SetBlackoutPeriod(start, end time.Time) {
	go func() {
		timeUntilStart := time.Until(start)
		if timeUntilStart > 0 {
			// Wait until the start of the blackout period.
			time.Sleep(timeUntilStart)
		}

		// Trip the surge protector at the start time.
		s.Trip()
		s.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Trip => Surge protector tripped at start of blackout period.", s.componentMetadata)

		// Calculate remaining time until the end if we're already in the blackout period when this is set.
		timeUntilEnd := time.Until(end)
		if timeUntilEnd > 0 {
			time.Sleep(timeUntilEnd)
		}

		// Reset the surge protector on

		s.Reset()
		s.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Reset => Surge protector reset at end of blackout period.", s.componentMetadata)
	}()
}

// Conditional activation based on system metrics or external signals
// This requires integration with system monitoring tools or APIs
func (s *SurgeProtector[T]) WithConditionalBlackout(check func() bool) {
	go func() {
		for {
			if check() {
				s.Trip()
			} else {
				s.Reset()
			}
			time.Sleep(1 * time.Minute) // Check every minute
		}
	}()
}

// Conditional activation based on system metrics or external signals
// This requires integration with system monitoring tools or APIs
func (s *SurgeProtector[T]) SetRateLimit(capacity int, refillRate time.Duration, maxRetryAttempts int) {
	s.capacity = int32(capacity)
	s.refillRate = refillRate
	s.maxRetryAttempts = maxRetryAttempts
	atomic.StoreInt32(&s.rateLimited, 1)
}

// Conditional activation based on system metrics or external signals
// This requires integration with system monitoring tools or APIs
func (s *SurgeProtector[T]) GetRateLimit() (int32, time.Duration, int, int) {
	return s.capacity, s.refillRate, s.maxRetryAttempts, s.backoffJitter
}

// GetTimeUntilNextRefill calculates and returns the duration until the next refill.
func (s *SurgeProtector[T]) GetTimeUntilNextRefill() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	if now.After(s.nextRefill) {
		return 0
	}
	return s.nextRefill.Sub(now)
}

func (s *SurgeProtector[T]) ReleaseToken() {
	atomic.AddInt32(&s.tokens, 1) // Increase token count atomically
	s.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: ReleaseToken => Released token", s.componentMetadata)
}

// AttachBackup connects a backup system to the surge protector.
func (s *SurgeProtector[T]) AttachBackup(backup ...types.Submitter[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backupSystems = append(s.backupSystems, backup...)
}

// DetachBackupSystem removes the attached backup system.
func (s *SurgeProtector[T]) DetachBackups() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, b := range s.backupSystems {
		s.NotifyLoggers(types.WarnLevel, "%s => level: WARN, result: SUCCESS, event: DetachBackups, target: %v => Detached backup from Surge Protector!", s.componentMetadata, b.GetComponentMetadata())
	}
	s.notifySurgeProtectorDetachedBackups(s.backupSystems...)
	s.backupSystems = nil
}

func (s *SurgeProtector[T]) Trip() {
	atomic.StoreInt32(&s.active, 1) // Set to true atomically
	s.notifySurgeProtectorTrip()
}

func (s *SurgeProtector[T]) Reset() {
	s.notifySurgeProtectorReset()
	atomic.StoreInt32(&s.active, 0) // Set to false atomically
	for _, mc := range s.managedComponents {
		s.NotifyLoggers(types.InfoLevel, "%s => level: INFO, result: PENDING, event: Reset, target: %v => Surge protector reset. Trying to restart managed components gracefully...", s.componentMetadata, mc.GetComponentMetadata())
		if mc.GetComponentMetadata().Type == "CONDUIT" {
			mc.Restart(s.ctx)
		} else {
			if !mc.IsStarted() {
				for _, ss := range s.sensors {
					ss.InvokeOnRestart(mc.GetComponentMetadata())
				}
				mc.Start(s.ctx)
			}
		}
		if len(mc.GetGenerators()) != 0 {
			for _, g := range mc.GetGenerators() {
				for _, ss := range s.sensors {
					ss.InvokeOnRestart(g.GetComponentMetadata())
				}
				g.Start(s.ctx)
			}
		}
	}

}

func (s *SurgeProtector[T]) IsTripped() bool {
	return atomic.LoadInt32(&s.active) == 1 // Check value atomically
}

func (s *SurgeProtector[T]) IsBeingRateLimited() bool {
	return atomic.LoadInt32(&s.active) == 1 // Check value atomically
}

func (s *SurgeProtector[T]) IsResisterConnected() bool {
	return atomic.LoadInt32(&s.resisterEnabled) == 1 // Check value atomically
}

func (s *SurgeProtector[T]) ConnectResister(r types.Resister[T]) {
	atomic.StoreInt32(&s.resisterEnabled, 1) // Check value atomically
	s.resister = r
	s.notifySurgeProtectorConnectResister(r)
}

func (s *SurgeProtector[T]) GetBackupSystems() []types.Submitter[T] {
	return s.backupSystems
}

func (s *SurgeProtector[T]) Submit(ctx context.Context, elem *types.Element[T]) error {

	if len(s.backupSystems) > 0 {
		for _, bs := range s.backupSystems {
			if !bs.IsStarted() {
				bs.Start(s.ctx)
			}
			s.NotifyLoggers(types.WarnLevel, "%s => level: WARN, result: PENDING, event: Submit, element: %v, target: %v => Routing to surge protector backup!", s.componentMetadata, elem, bs.GetComponentMetadata())
			err := bs.Submit(s.ctx, elem.Data)
			if err != nil {
				s.notifySurgeProtectorBackupFailure(err)
				return err
			}
			s.notifySurgeProtectorBackupSubmission(elem.Data)
		}
		return nil
	}

	if s.IsResisterConnected() {
		for s.resister.Len() > 0 && s.TryTake() {
			nextElem := s.resister.Pop()
			if nextElem != nil {
				if err := s.submitToComponents(s.ctx, nextElem.Data); err != nil {
					nextElem.IncrementRetryCount() // Properly increment retry count
					nextElem.AdjustPriority()      // Adjust priority based on current logic
					s.NotifyLoggers(types.ErrorLevel, "component: %s, level: ERROR, result: FAILURE, event: Submit, element: %v, error: %v => Failed to processed queued item, requeuing...", s.componentMetadata, nextElem, err)
					fmt.Printf("Failed to process queued item: %s, requeuing\n", nextElem.ID)
					s.resister.Push(nextElem)
					return err
				}
			}
		}

		if s.TryTake() {
			return s.submitToComponents(s.ctx, elem.Data)
		} else {
			s.NotifyLoggers(types.WarnLevel, "component: %s, level: WARN, result: PENDING, event: Submit, element: %v => Hit rate limit, queueing element", s.componentMetadata, elem) // Adjust initial priority
			s.Enqueue(elem)
		}
	}

	return nil
}

// ConnectLogger attaches loggers to the SurgeProtector.
func (s *SurgeProtector[T]) ConnectLogger(loggers ...types.Logger) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.loggers = append(s.loggers, loggers...)
}

// ConnectLogger attaches loggers to the SurgeProtector.
func (s *SurgeProtector[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sensors = append(s.sensors, sensors...)
}

func (s *SurgeProtector[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if s.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range s.loggers {
			if logger == nil {
				continue
			}
			// Ensure we only acquire the lock once per logger to avoid deadlock or excessive locking overhead
			s.loggersLock.Lock()
			if logger.GetLevel() <= level {
				switch level {
				case types.DebugLevel:
					logger.Debug(msg)
				case types.InfoLevel:
					logger.Info(msg)
				case types.WarnLevel:
					logger.Warn(msg)
				case types.ErrorLevel:
					logger.Error(msg)
				case types.DPanicLevel:
					logger.DPanic(msg)
				case types.PanicLevel:
					logger.Panic(msg)
				case types.FatalLevel:
					logger.Fatal(msg)
				}
			}
			s.loggersLock.Unlock()
		}
	}
}

// GetComponentMetadata returns the metadata.
func (s *SurgeProtector[T]) GetComponentMetadata() types.ComponentMetadata {
	return s.componentMetadata
}

// GetComponentMetadata returns the metadata.
func (s *SurgeProtector[T]) GetResisterQueue() int {
	return s.resister.Len()
}

// SetComponentMetadata sets the component metadata.
func (s *SurgeProtector[T]) SetComponentMetadata(name string, id string) {
	s.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

func (s *SurgeProtector[T]) TryTake() bool {

	s.tokenLock.Lock()
	defer s.tokenLock.Unlock()
	if time.Now().After(s.nextRefill) {
		s.tokens = s.capacity
		s.nextRefill = time.Now().Add(s.refillRate)
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: SUCCESS, event: TryTake, currentTokens: %v => Tokens refilled", s.componentMetadata, s.tokens)
		s.notifySurgeProtectorReleaseToken()
	}

	if s.tokens >= 0 {
		s.tokens--
		s.NotifyLoggers(types.DebugLevel, "component: %s, level: DEBUG, result: PENDING, event: TryTake, currentTokens: %v => Token taken, processing allowed...", s.componentMetadata, s.tokens)
		return true
	}

	return false
}

// Enqueue adds an element to the queue.
func (s *SurgeProtector[T]) Enqueue(element *types.Element[T]) error {
	return s.resister.Push(element)
}

// Dequeue retrieves and removes the highest priority element from the queue.
func (s *SurgeProtector[T]) Dequeue() (*types.Element[T], error) {
	// Ensure thread safety if required
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.resister.Len() == 0 { // Check if the queue is empty
		return nil, errors.New("queue is empty")
	}

	element := s.resister.Pop() // Assumes Pop returns the highest priority item

	if element == nil {
		return nil, errors.New("failed to retrieve an element")
	}
	return element, nil
}
