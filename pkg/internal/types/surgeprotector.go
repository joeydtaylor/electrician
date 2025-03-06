package types

import (
	"context"
	"time"
)

// SurgeProtector defines the operations required for managing data during surge conditions.
type SurgeProtector[T any] interface {
	// AttachBackupSystem connects a backup system where data can be redirected during surge conditions.
	AttachBackup(...Submitter[T])
	ConnectSensor(sensors ...Sensor[T])

	// DetachBackupSystem removes the attached backup system, ceasing any redirection of data.
	DetachBackups()

	// Activate enables the surge protection, starting redirection of data to the backup system if necessary.
	Trip()

	// Deactivate disables the surge protection, stopping any redirection and resuming normal operation.
	Reset()

	IsBeingRateLimited() bool

	ConnectComponent(c ...Submitter[T])
	GetTimeUntilNextRefill() time.Duration
	ConnectResister(Resister[T])

	Enqueue(element *Element[T]) error
	Dequeue() (*Element[T], error)

	IsTripped() bool
	IsResisterConnected() bool
	GetBackupSystems() []Submitter[T]
	GetResisterQueue() int

	SetBlackoutPeriod(start, end time.Time)

	WithConditionalBlackout(check func() bool)

	SetRateLimit(capacity int, refillRate time.Duration, maxRetryAttempts int)
	GetRateLimit() (int32, time.Duration, int, int)

	// Submit handles the incoming data by deciding whether to process it normally or divert it based on the active state.
	Submit(ctx context.Context, elem *Element[T]) error

	// ConnectLogger attaches loggers to the SurgeProtector for operational logging.
	ConnectLogger(...Logger)

	// NotifyLoggers allows sending formatted log messages to attached loggers.
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	// GetComponentMetadata retrieves details about the SurgeProtector, such as name or ID.
	GetComponentMetadata() ComponentMetadata

	// SetComponentMetadata sets details like name and ID for the SurgeProtector.
	SetComponentMetadata(name string, id string)

	TryTake() bool
}
