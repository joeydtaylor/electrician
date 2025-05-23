package sensor

import (
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// Sensor defines a structure for monitoring processing events and managing callbacks in data flow systems.
type Sensor[T any] struct {
	componentMetadata                     types.ComponentMetadata             // Metadata describing the sensor.
	OnStart                               []func(cmd types.ComponentMetadata) // Callbacks for start events.
	OnInsulatorAttempt                    []func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)
	OnInsulatorSuccess                    []func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)
	OnInsulatorFailure                    []func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)
	OnRestart                             []func(cmd types.ComponentMetadata)                            // Callbacks for start events.
	OnSubmit                              []func(cmd types.ComponentMetadata, elem T)                    // Callbacks for start events.
	OnElementProcessed                    []func(cmd types.ComponentMetadata, elem T)                    // Callbacks for each processed element.
	OnError                               []func(cmd types.ComponentMetadata, err error, elem T)         // Callbacks for error handling.
	OnCancel                              []func(cmd types.ComponentMetadata, elem T)                    // Callbacks for cancellation events.
	OnComplete                            []func(cmd types.ComponentMetadata)                            // Callbacks for completion events.
	OnTerminate                           []func(cmd types.ComponentMetadata)                            // Callbacks for termination events.
	OnHTTPClientRequestStart              []func(http types.ComponentMetadata)                           // Callbacks for HTTP client request start events.
	OnHTTPClientResponseReceived          []func(http types.ComponentMetadata)                           // Callbacks for HTTP client response received events.
	OnHTTPClientError                     []func(http types.ComponentMetadata, err error)                // Callbacks for HTTP client error events.
	OnHTTPClientRequestComplete           []func(http types.ComponentMetadata)                           // Callbacks for HTTP client request complete events.
	OnSurgeProtectorTrip                  []func(sp types.ComponentMetadata)                             // Callbacks for when the surge protector trips.
	OnSurgeProtectorReset                 []func(sp types.ComponentMetadata)                             // Callbacks for when the surge protector resets.
	OnSurgeProtectorBackupFailure         []func(sp types.ComponentMetadata, err error)                  // Callbacks for backup system failures.
	OnSurgeProtectorRateLimitExceeded     []func(sp types.ComponentMetadata, elem T)                     // Callbacks for when the surge protector's rate limit is exceeded.
	OnResisterDequeued                    []func(r types.ComponentMetadata, elem T)                      // Callbacks for when elements in the queue are processed.
	OnResisterQueued                      []func(r types.ComponentMetadata, elem T)                      // Callbacks for when elements in the queue are processed.
	OnResisterRequeued                    []func(r types.ComponentMetadata, elem T)                      // Callbacks for when elements in the queue are processed.
	OnResisterEmpty                       []func(r types.ComponentMetadata)                              // Callbacks for when the surge protector's queue becomes empty.
	OnSurgeProtectorReleaseToken          []func(sp types.ComponentMetadata)                             // Callbacks for releasing tokens.
	OnSurgeProtectorConnectResister       []func(sp types.ComponentMetadata, r types.ComponentMetadata)  // Callbacks for connecting a resister.
	OnSurgeProtectorDetachedBackups       []func(sp types.ComponentMetadata, bs types.ComponentMetadata) // Callbacks for detaching backup systems.
	OnSurgeProtectorBackupWireSubmit      []func(sp types.ComponentMetadata, elem T)                     // Callbacks for detaching backup systems.
	OnSurgeProtectorSubmit                []func(sp types.ComponentMetadata, elem T)                     // Callbacks for detaching backup systems.
	OnSurgeProtectorDrop                  []func(sp types.ComponentMetadata, elem T)                     // Callbacks for detaching backup systems.
	OnCircuitBreakerTrip                  []func(sp types.ComponentMetadata, time int64, nextReset int64)
	OnCircuitBreakerReset                 []func(sp types.ComponentMetadata, time int64)
	OnCircuitBreakerNeutralWireSubmission []func(sp types.ComponentMetadata, elem T)
	OnCircuitBreakerRecordError           []func(sp types.ComponentMetadata, time int64)
	OnCircuitBreakerAllow                 []func(sp types.ComponentMetadata)
	OnCircuitBreakerDrop                  []func(sp types.ComponentMetadata, elem T)

	callbackLock sync.Mutex
	loggers      []types.Logger // Attached loggers for event logging.
	loggersLock  sync.Mutex     // Mutex to ensure logger access is thread-safe.
	meters       []types.Meter[T]
}

// NewSensor creates a new Sensor instance with optional configuration.
// It initializes a Sensor with unique metadata and sets up callback lists for various events.
func NewSensor[T any](options ...types.Option[types.Sensor[T]]) types.Sensor[T] {
	m := &Sensor[T]{
		componentMetadata: types.ComponentMetadata{
			ID:   utils.GenerateUniqueHash(), // Generate a unique identifier for the Sensor.
			Type: "SENSOR",
		},
		meters:             make([]types.Meter[T], 0),
		OnStart:            make([]func(cmd types.ComponentMetadata), 0),
		OnRestart:          make([]func(cmd types.ComponentMetadata), 0),
		OnSubmit:           make([]func(cmd types.ComponentMetadata, elem T), 0),
		OnElementProcessed: make([]func(cmd types.ComponentMetadata, elem T), 0),
		OnError:            make([]func(cmd types.ComponentMetadata, err error, elem T), 0),
		OnCancel:           make([]func(cmd types.ComponentMetadata, elem T), 0),
		OnComplete:         make([]func(cmd types.ComponentMetadata), 0),
		OnTerminate:        make([]func(cmd types.ComponentMetadata), 0),
		OnInsulatorAttempt: make([]func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration), 0),
		OnInsulatorFailure: make([]func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration), 0),
		OnInsulatorSuccess: make([]func(c types.ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration), 0),

		OnHTTPClientRequestStart:     make([]func(http types.ComponentMetadata), 0),
		OnHTTPClientResponseReceived: make([]func(http types.ComponentMetadata), 0),
		OnHTTPClientError:            make([]func(http types.ComponentMetadata, err error), 0),
		OnHTTPClientRequestComplete:  make([]func(http types.ComponentMetadata), 0),

		OnSurgeProtectorTrip:              make([]func(sp types.ComponentMetadata), 0),
		OnSurgeProtectorReset:             make([]func(sp types.ComponentMetadata), 0),
		OnSurgeProtectorBackupFailure:     make([]func(sp types.ComponentMetadata, err error), 0),
		OnSurgeProtectorBackupWireSubmit:  make([]func(sp types.ComponentMetadata, elem T), 0),
		OnSurgeProtectorSubmit:            make([]func(sp types.ComponentMetadata, elem T), 0),
		OnSurgeProtectorDrop:              make([]func(sp types.ComponentMetadata, elem T), 0),
		OnSurgeProtectorRateLimitExceeded: make([]func(sp types.ComponentMetadata, elem T), 0),
		OnSurgeProtectorReleaseToken:      make([]func(sp types.ComponentMetadata), 0),
		OnSurgeProtectorConnectResister:   make([]func(sp types.ComponentMetadata, r types.ComponentMetadata), 0),
		OnSurgeProtectorDetachedBackups:   make([]func(sp types.ComponentMetadata, bu types.ComponentMetadata), 0),

		OnCircuitBreakerTrip:                  make([]func(sp types.ComponentMetadata, time int64, nextReset int64), 0),
		OnCircuitBreakerReset:                 make([]func(sp types.ComponentMetadata, time int64), 0),
		OnCircuitBreakerNeutralWireSubmission: make([]func(sp types.ComponentMetadata, elem T), 0),
		OnCircuitBreakerRecordError:           make([]func(sp types.ComponentMetadata, time int64), 0),
		OnCircuitBreakerAllow:                 make([]func(sp types.ComponentMetadata), 0),
		OnCircuitBreakerDrop:                  make([]func(sp types.ComponentMetadata, elem T), 0),

		OnResisterDequeued: make([]func(r types.ComponentMetadata, elem T), 0),
		OnResisterQueued:   make([]func(r types.ComponentMetadata, elem T), 0),
		OnResisterRequeued: make([]func(r types.ComponentMetadata, elem T), 0),
		OnResisterEmpty:    make([]func(r types.ComponentMetadata), 0),
	}

	// Apply configuration options to the Sensor.
	for _, opt := range m.decorateCallbacks(options...) {
		opt(m)
	}

	return m // Return the configured Sensor.
}
