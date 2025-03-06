// Package sensor provides options for configuring Sensor components.
//
// This file defines various options that can be used to customize the behavior and settings of Sensor components
// within a data processing pipeline. These options allow users to add loggers, register callbacks for specific
// events such as OnStart or OnError, and configure custom metadata for Sensors.
package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// WithLogger creates an option to add a logger to a Sensor.
//
// Parameters:
//   - logger: One or more logger instances to be added to the Sensor for logging.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	connects the specified logger(s) to the Sensor.
func WithLogger[T any](logger ...types.Logger) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.ConnectLogger(logger...)
	}
}

// WithOnStartFunc creates an option to register a callback for the OnStart event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnStart event, each accepting no parasensors.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnStart event.
func WithOnStartFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnStart(callback...)
	}
}

// WithOnElementProcessedFunc creates an option to register a callback for the OnElementProcessed event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnElementProcessed event, each accepting an element of type T.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnElementProcessed event.
func WithOnElementProcessedFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnElementProcessed(callback...)
	}
}

// WithOnErrorFunc creates an option to register a callback for the OnError event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnError event, each accepting an error and an element of type T.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnError event.
func WithOnErrorFunc[T any](callback ...func(c types.ComponentMetadata, err error, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnError(callback...)
	}
}

// WithOnCancelFunc creates an option to register a callback for the OnCancel event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnCancel event, each accepting an element of type T.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnCancel event.
func WithOnCancelFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCancel(callback...)
	}
}

// WithOnCompleteFunc creates an option to register a callback for the OnComplete event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnComplete event, each accepting no parasensors.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnComplete event.
func WithOnCompleteFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnComplete(callback...)
	}
}

// WithOnStopFunc creates an option to register a callback for the OnTerminate event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnTerminate event, each accepting no parasensors.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnTerminate event.
func WithOnStopFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnTerminate(callback...)
	}
}

// WithComponentMetadata configures a Sensor component with custom metadata, including a name and an identifier.
// This function provides an option to set these metadata properties, which can be used for identification,
// logging, or other purposes where metadata is needed for a Sensor. It uses the SetComponentMetadata method
// internally to apply these settings. If the Sensor's configuration is frozen (indicating that the component
// has started operation and its configuration should no longer be changed), attempting to set metadata
// will result in a panic. This ensures the integrity of component configurations during runtime.
//
// Parameters:
//   - name: The name to set for the Sensor component, used for identification and logging.
//   - id: The unique identifier to set for the Sensor component, used for unique identification across systems.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]], which when called with a Sensor component,
//	sets the specified name and id in the component's metadata.
func WithComponentMetadata[T any](name string, id string) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.SetComponentMetadata(name, id)
	}
}

// WithOnHTTPClientRequestStartFunc creates an option to register a callback for the OnHTTPClientRequestStart event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnHTTPClientRequestStart event, each accepting no parameters.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnHTTPClientRequestStart event.
func WithOnHTTPClientRequestStartFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientRequestStart(callback...)
	}
}

// WithOnHTTPClientResponseReceivedFunc creates an option to register a callback for the OnHTTPClientResponseReceived event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnHTTPClientResponseReceived event, each accepting no parameters.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnHTTPClientResponseReceived event.
func WithOnHTTPClientResponseReceivedFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientResponseReceived(callback...)
	}
}

// WithOnHTTPClientErrorFunc creates an option to register a callback for the OnHTTPClientError event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnHTTPClientError event, each accepting an error parameter.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnHTTPClientError event.
func WithOnHTTPClientErrorFunc[T any](callback ...func(c types.ComponentMetadata, err error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientError(callback...)
	}
}

// WithOnHTTPClientRequestCompleteFunc creates an option to register a callback for the OnHTTPClientRequestComplete event.
//
// Parameters:
//   - callback: One or more callback functions to be registered for the OnHTTPClientRequestComplete event, each accepting no parameters.
//
// Returns:
//
//	A function conforming to types.Option[types.Sensor[T]] that, when called with a Sensor component,
//	registers the specified callback(s) for the OnHTTPClientRequestComplete event.
func WithOnHTTPClientRequestCompleteFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientRequestComplete(callback...)
	}
}

// WithSurgeProtectorTripFunc creates an option to register a callback for the OnSurgeProtectorTrip event.
func WithSurgeProtectorTripFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorTrip(callback...)
	}
}

// WithSurgeProtectorResetFunc creates an option to register a callback for the OnSurgeProtectorReset event.
func WithSurgeProtectorResetFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorReset(callback...)
	}
}

// WithSurgeProtectorBackupFailureFunc creates an option to register a callback for the OnBackupFailure event, including error handling.
func WithSurgeProtectorBackupFailureFunc[T any](callback ...func(c types.ComponentMetadata, err error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorBackupFailure(callback...)
	}
}

// WithSurgeProtectorBackupFailureFunc creates an option to register a callback for the OnBackupFailure event, including error handling.
func WithSurgeProtectorBackupWireSubmissionFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorBackupWireSubmit(callback...)
	}
}

// WithSurgeProtectorBackupFailureFunc creates an option to register a callback for the OnBackupFailure event, including error handling.
func WithSurgeProtectorDroppedSubmissionFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorDrop(callback...)
	}
}

// WithSurgeProtectorRateLimitExceededFunc creates an option to register a callback for the OnRateLimitExceeded event.
func WithSurgeProtectorRateLimitExceededFunc[T any](callback ...func(c types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorRateLimitExceeded(callback...)
	}
}

// WithSurgeProtectorReleaseTokenFunc creates an option to register a callback for the OnReleaseToken event.
func WithSurgeProtectorReleaseTokenFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorReleaseToken(callback...)
	}
}

// WithSurgeProtectorConnectResisterFunc creates an option to register a callback for the OnConnectResister event.
func WithSurgeProtectorConnectResisterFunc[T any](callback ...func(c types.ComponentMetadata, r types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorConnectResister(callback...)
	}
}

// WithSurgeProtectorDetachedBackupsFunc creates an option to register a callback for the OnDetachedBackups event.
func WithSurgeProtectorDetachedBackupsFunc[T any](callback ...func(c types.ComponentMetadata, bu types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorDetachedBackups(callback...)
	}
}

// WithSurgeProtectorQueueProcessedFunc creates an option to register a callback for the OnQueueProcessed event.
func WithResisterDequeuedFunc[T any](callback ...func(r types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterDequeued(callback...)
	}
}

// WithSurgeProtectorQueueProcessedFunc creates an option to register a callback for the OnQueueProcessed event.
func WithResisterQueuedFunc[T any](callback ...func(r types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterQueued(callback...)
	}
}

// WithSurgeProtectorQueueProcessedFunc creates an option to register a callback for the OnQueueProcessed event.
func WithResisterRequeuedFunc[T any](callback ...func(r types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterRequeued(callback...)
	}
}

// WithSurgeProtectorQueueEmptyFunc creates an option to register a callback for the OnQueueEmpty event.
func WithResisterEmptyFunc[T any](callback ...func(r types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterEmpty(callback...)
	}
}

// WithSurgeProtectorQueueEmptyFunc creates an option to register a callback for the OnQueueEmpty event.
func WithOnSubmitFunc[T any](callback ...func(r types.ComponentMetadata, elem T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSubmit(callback...)
	}
}

// WithSurgeProtectorQueueEmptyFunc creates an option to register a callback for the OnQueueEmpty event.
func WithOnRestartFunc[T any](callback ...func(r types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnRestart(callback...)
	}
}

// WithSurgeProtectorQueueEmptyFunc creates an option to register a callback for the OnQueueEmpty event.
func WithMeter[T any](meter ...types.Meter[T]) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.ConnectMeter(meter...)
	}
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
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerTrip(callback...)
	}
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
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerReset(callback...)
	}
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
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerNeutralWireSubmission(callback...)
	}
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
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerRecordError(callback...)
	}
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
func WithCircuitBreakerAllowFunc[T any](callback ...func(c types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerAllow(callback...)
	}
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
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerDrop(callback...)
	}
}
