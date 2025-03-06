package builder

import (
	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SensorWithOnHTTPClientRequestStartFunc registers a callback for the OnHTTPClientRequestStart event.
func SensorWithOnHTTPClientRequestStartFunc[T any](callback ...func(ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithOnHTTPClientRequestStartFunc[T](callback...)
}

// SensorWithOnHTTPClientResponseReceivedFunc registers a callback for the OnHTTPClientResponseReceived event.
func SensorWithOnHTTPClientResponseReceivedFunc[T any](callback ...func(ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithOnHTTPClientResponseReceivedFunc[T](callback...)
}

// SensorWithOnHTTPClientErrorFunc registers a callback for the OnHTTPClientError event.
func SensorWithOnHTTPClientErrorFunc[T any](callback ...func(http ComponentMetadata, err error)) types.Option[types.Sensor[T]] {
	return sensor.WithOnHTTPClientErrorFunc[T](callback...)
}

// SensorWithOnHTTPClientRequestCompleteFunc registers a callback for the OnHTTPClientRequestComplete event.
func SensorWithOnHTTPClientRequestCompleteFunc[T any](callback ...func(http ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithOnHTTPClientRequestCompleteFunc[T](callback...)
}

// SensorWithComponentMetadata adds component metadata overrides.
func SensorWithComponentMetadata[T any](name string, id string) types.Option[types.Sensor[T]] {
	return sensor.WithComponentMetadata[T](name, id)
}

// SensorWithLogger adds a logger to the Sensor.
func SensorWithLogger[T any](logger ...types.Logger) types.Option[types.Sensor[T]] {
	return sensor.WithLogger[T](logger...)
}

// SensorWithOnCancelFunc registers a callback for the OnCancel event.
func SensorWithOnCancelFunc[T any](callback ...func(c ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithOnCancelFunc[T](callback...)
}

// SensorWithOnCompleteFunc registers a callback for the OnComplete event.
func SensorWithOnCompleteFunc[T any](callback ...func(c ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithOnCompleteFunc[T](callback...)
}

// SensorWithOnErrorFunc registers a callback for the OnError event.
func SensorWithOnErrorFunc[T any](callback ...func(c ComponentMetadata, err error, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithOnErrorFunc[T](callback...)
}

// SensorWithOnElementProcessedFunc registers a callback for the OnElementProcessed event.
func SensorWithOnElementProcessedFunc[T any](callback ...func(c ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithOnElementProcessedFunc[T](callback...)
}

// SensorWithOnStartFunc registers a callback for the OnStart event.
func SensorWithOnStartFunc[T any](callback ...func(c ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithOnStartFunc[T](callback...)
}

// SensorWithOnStopFunc registers a callback for the OnTerminate event.
func SensorWithOnStopFunc[T any](callback ...func(c ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithOnStopFunc[T](callback...)
}

// SensorWithSurgeProtectorTripFunc creates an option to register a callback for the OnSurgeProtectorTrip event.
func SensorWithSurgeProtectorTripFunc[T any](callback ...func(c ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorTripFunc[T](callback...)
}

// SensorWithSurgeProtectorResetFunc creates an option to register a callback for the OnSurgeProtectorReset event.
func SensorWithSurgeProtectorResetFunc[T any](callback ...func(c ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorResetFunc[T](callback...)
}

// SensorWithSurgeProtectorBackupFailureFunc creates an option to register a callback for the OnBackupFailure event, including error handling.
func SensorWithSurgeProtectorBackupFailureFunc[T any](callback ...func(c ComponentMetadata, err error)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorBackupFailureFunc[T](callback...)
}

// SensorWithSurgeProtectorRateLimitExceededFunc creates an option to register a callback for the OnRateLimitExceeded event.
func SensorWithSurgeProtectorRateLimitExceededFunc[T any](callback ...func(c ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorRateLimitExceededFunc[T](callback...)
}

// SensorWithSurgeProtectorReleaseTokenFunc creates an option to register a callback for the OnReleaseToken event.
func SensorWithSurgeProtectorReleaseTokenFunc[T any](callback ...func(c ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorReleaseTokenFunc[T](callback...)
}

// WithSurgeProtectorBackupFailureFunc creates an option to register a callback for the OnBackupFailure event, including error handling.
func SensorWithSurgeProtectorBackupWireSubmissionFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorBackupWireSubmissionFunc[T](callback...)
}

// WithSurgeProtectorBackupFailureFunc creates an option to register a callback for the OnBackupFailure event, including error handling.
func SensorWithSurgeProtectorDroppedSubmissionFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorDroppedSubmissionFunc[T](callback...)
}

// SensorWithSurgeProtectorConnectResisterFunc creates an option to register a callback for the OnConnectResister event.
func SensorWithSurgeProtectorConnectResisterFunc[T any](callback ...func(c ComponentMetadata, r ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorConnectResisterFunc[T](callback...)
}

// SensorWithSurgeProtectorDetachedBackupsFunc creates an option to register a callback for the OnDetachedBackups event.
func SensorWithSurgeProtectorDetachedBackupsFunc[T any](callback ...func(c ComponentMetadata, bu ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithSurgeProtectorDetachedBackupsFunc[T](callback...)
}

// SensorWithSurgeProtectorQueueProcessedFunc creates an option to register a callback for the OnQueueProcessed event.
func SensorWithResisterDequeuedFunc[T any](callback ...func(c ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithResisterDequeuedFunc[T](callback...)
}

// SensorWithSurgeProtectorQueueProcessedFunc creates an option to register a callback for the OnQueueProcessed event.
func SensorWithResisterQueuedFunc[T any](callback ...func(c ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithResisterQueuedFunc[T](callback...)
}

// SensorWithSurgeProtectorQueueProcessedFunc creates an option to register a callback for the OnQueueProcessed event.
func SensorWithResisterRequeuedFunc[T any](callback ...func(c ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithResisterRequeuedFunc[T](callback...)
}

// SensorWithSurgeProtectorQueueEmptyFunc creates an option to register a callback for the OnQueueEmpty event.
func SensorWithResisterEmptyFunc[T any](callback ...func(c ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithResisterEmptyFunc[T](callback...)
}

// WithSurgeProtectorQueueEmptyFunc creates an option to register a callback for the OnQueueEmpty event.
func SensorWithMeter[T any](meter ...types.Meter[T]) types.Option[types.Sensor[T]] {
	return sensor.WithMeter[T](meter...)
}

// WithSurgeProtectorQueueEmptyFunc creates an option to register a callback for the OnQueueEmpty event.
func SensorWithOnRestartFunc[T any](callback ...func(r types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithOnRestartFunc[T](callback...)
}

// WithCircuitBreakerTripFunc creates an option to register a callback for the CircuitBreakerTrip event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the CircuitBreakerTrip event.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the CircuitBreakerTrip event.
func WithCircuitBreakerTripFunc[T any](callback ...func(c types.ComponentMetadata, time int64, nextReset int64)) types.Option[types.Sensor[T]] {
	return sensor.WithCircuitBreakerTripFunc[T](callback...)
}

// WithCircuitBreakerResetFunc creates an option to register a callback for the CircuitBreakerReset event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the CircuitBreakerReset event.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the CircuitBreakerReset event.
func WithCircuitBreakerResetFunc[T any](callback ...func(c types.ComponentMetadata, time int64)) types.Option[types.Sensor[T]] {
	return sensor.WithCircuitBreakerResetFunc[T](callback...)
}

// WithCircuitBreakerGroundWireSubmissionFunc creates an option to register a callback for the CircuitBreakerGroundWireSubmission event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the CircuitBreakerGroundWireSubmission event, each accepting an element of type T.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the CircuitBreakerGroundWireSubmission event.
func WithCircuitBreakerNeutralWireSubmissionFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithCircuitBreakerNeutralWireSubmissionFunc[T](callback...)
}

// WithCircuitBreakerRecordErrorFunc creates an option to register a callback for the CircuitBreakerRecordError event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the CircuitBreakerRecordError event.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the CircuitBreakerRecordError event.
func WithCircuitBreakerRecordErrorFunc[T any](callback ...func(c types.ComponentMetadata, time int64)) types.Option[types.Sensor[T]] {
	return sensor.WithCircuitBreakerRecordErrorFunc[T](callback...)
}

// WithCircuitBreakerAllowFunc creates an option to register a callback for the CircuitBreakerAllow event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the CircuitBreakerAllow event.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the CircuitBreakerAllow event.
func SensorWithWithCircuitBreakerAllowFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return sensor.WithCircuitBreakerAllowFunc[T](callback...)
}

// WithCircuitBreakerDropFunc creates an option to register a callback for the CircuitBreakerDrop event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the CircuitBreakerDrop event, each accepting an element of type T.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the CircuitBreakerDrop event.
func WithCircuitBreakerDropFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return sensor.WithCircuitBreakerDropFunc[T](callback...)
}

// NewSensor creates a new Sensor with specified options.
func NewSensor[T any](options ...types.Option[types.Sensor[T]]) types.Sensor[T] {
	return sensor.NewSensor[T](options...)
}
