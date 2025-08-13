package types

import "time"

type Sensor[T any] interface {
	// ConnectLogger attaches one or more loggers to the Sensor. This is essential for recording significant
	// events and states within the sensoring process, helping in diagnostics and system monitoring.
	// Circuit breaker specific registration methods
	RegisterOnCircuitBreakerTrip(...func(ComponentMetadata, int64, int64))
	RegisterOnCircuitBreakerReset(...func(ComponentMetadata, int64))
	RegisterOnCircuitBreakerNeutralWireSubmission(...func(ComponentMetadata, T))
	RegisterOnCircuitBreakerRecordError(...func(ComponentMetadata, int64))
	RegisterOnCircuitBreakerAllow(...func(ComponentMetadata))
	RegisterOnCircuitBreakerDrop(...func(ComponentMetadata, T))

	// Circuit breaker specific invocation methods
	InvokeOnCircuitBreakerTrip(ComponentMetadata, int64, int64)
	InvokeOnCircuitBreakerReset(ComponentMetadata, int64)
	InvokeOnCircuitBreakerNeutralWireSubmission(ComponentMetadata, T)
	InvokeOnCircuitBreakerRecordError(ComponentMetadata, int64)
	InvokeOnCircuitBreakerAllow(ComponentMetadata)
	InvokeOnCircuitBreakerDrop(ComponentMetadata, T)

	RegisterOnRestart(callback ...func(c ComponentMetadata))
	InvokeOnRestart(c ComponentMetadata)
	ConnectLogger(...Logger)
	ConnectMeter(meter ...Meter[T])

	GetMeters() []Meter[T]

	// GetComponentMetadata retrieves metadata about the Sensor, including identifiers like name and ID,
	// useful for logging and monitoring purposes.
	GetComponentMetadata() ComponentMetadata

	// HTTP client specific hooks
	InvokeOnHTTPClientRequestStart(ComponentMetadata)
	InvokeOnHTTPClientResponseReceived(ComponentMetadata)
	InvokeOnHTTPClientError(cm ComponentMetadata, err error)
	InvokeOnHTTPClientRequestComplete(ComponentMetadata)

	// InvokeOnCancel triggers all registered callbacks associated with the cancellation of a process.
	// This function is called when a monitored process is cancelled, allowing for custom handling.
	InvokeOnCancel(cm ComponentMetadata, elem T)

	// InvokeOnComplete triggers all registered callbacks upon the completion of a process. This function
	// is called when a monitored process successfully completes, facilitating downstream activities or cleanup.
	InvokeOnComplete(ComponentMetadata)

	// InvokeOnElementProcessed triggers callbacks when an element has been processed. This is particularly
	// useful for real-time monitoring and processing feedback.
	InvokeOnElementProcessed(cm ComponentMetadata, elem T)

	RegisterOnInsulatorAttempt(...func(c ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration))
	RegisterOnInsulatorSuccess(...func(c ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration))
	RegisterOnInsulatorFailure(...func(c ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration))

	InvokeOnInsulatorAttempt(c ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)
	InvokeOnInsulatorSuccess(c ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)
	InvokeOnInsulatorFailure(c ComponentMetadata, currentElement T, originalElement T, currentErr error, originalErr error, currentAttempt int, maxThreshold int, interval time.Duration)

	// InvokeOnError triggers the registered error handling callbacks. This function is called when an error
	// occurs during the monitored processes, allowing for immediate and custom error handling strategies.
	InvokeOnError(cm ComponentMetadata, err error, elem T)

	// InvokeOnStart triggers all registered callbacks at the start of a process. This allows for initialization
	// actions or logging to be performed right as the process begins.
	InvokeOnStart(cm ComponentMetadata)

	// InvokeOnTerminate triggers all registered callbacks upon the termination of a process. This is useful
	// for performing cleanup or finalization tasks.
	InvokeOnStop(cm ComponentMetadata)

	// NotifyLoggers sends a formatted log message to all attached loggers at a specified log level. This method
	// supports dynamic and contextual logging based on the Sensor's observations and state changes.
	NotifyLoggers(level LogLevel, msg string, keysAndValues ...interface{})

	// RegisterOnCancel registers a callback to be called when a cancellation event occurs. This allows for
	// custom actions based on cancellation scenarios.
	RegisterOnCancel(...func(cm ComponentMetadata, elem T))

	// RegisterOnComplete registers a callback to be called upon the completion of a process. This enables
	// actions to be triggered automatically at the end of a monitored process.
	RegisterOnComplete(...func(cm ComponentMetadata))

	// RegisterOnElementProcessed registers a callback to be triggered when an element is processed.
	// This allows for real-time monitoring or post-processing actions specific to each element.
	RegisterOnElementProcessed(...func(cm ComponentMetadata, elem T))

	// RegisterOnError registers a callback to handle errors occurring during the processing of elements.
	// This facilitates immediate and custom error responses within the monitoring context.
	RegisterOnError(...func(cm ComponentMetadata, err error, elem T))

	// RegisterOnStart registers a callback to be executed at the start of a process. This can be used
	// to perform any preparatory actions that are necessary before processing begins.
	RegisterOnStart(...func(cm ComponentMetadata))
	RegisterOnSubmit(callback ...func(c ComponentMetadata, elem T))

	// InvokeOnTerminate triggers all registered callbacks upon the termination of a process. This is useful
	// for performing cleanup or finalization tasks.
	InvokeOnSubmit(cm ComponentMetadata, elem T)

	// RegisterOnTerminate registers a callback to be invoked when a process is terminated. This allows
	// for the execution of finalization or cleanup routines.
	RegisterOnTerminate(...func(cm ComponentMetadata))

	// Registration methods for HTTP client specific hooks
	RegisterOnHTTPClientRequestStart(...func(ComponentMetadata))
	RegisterOnHTTPClientResponseReceived(...func(ComponentMetadata))
	RegisterOnHTTPClientError(...func(ComponentMetadata, error))
	RegisterOnHTTPClientRequestComplete(...func(ComponentMetadata))

	// SetComponentMetadata sets the metadata for the Sensor, such as its name and ID. Adjusting these
	// properties can be crucial for identifying or reconfiguring the Sensor during runtime.
	SetComponentMetadata(name string, id string)

	// Surge protector specific hooks
	RegisterOnSurgeProtectorTrip(...func(ComponentMetadata))
	RegisterOnSurgeProtectorReset(...func(ComponentMetadata))
	RegisterOnSurgeProtectorBackupFailure(...func(ComponentMetadata, error))
	RegisterOnSurgeProtectorRateLimitExceeded(...func(c ComponentMetadata, elem T))
	RegisterOnSurgeProtectorBackupWireSubmit(...func(r ComponentMetadata, elem T))
	RegisterOnSurgeProtectorSubmit(...func(r ComponentMetadata, elem T))
	RegisterOnSurgeProtectorDrop(...func(r ComponentMetadata, elem T))
	RegisterOnResisterDequeued(...func(r ComponentMetadata, elem T))
	RegisterOnResisterQueued(...func(r ComponentMetadata, elem T))
	RegisterOnResisterRequeued(...func(r ComponentMetadata, elem T))
	RegisterOnResisterEmpty(...func(ComponentMetadata))
	RegisterOnSurgeProtectorReleaseToken(...func(ComponentMetadata))
	RegisterOnSurgeProtectorConnectResister(...func(ComponentMetadata, ComponentMetadata))
	RegisterOnSurgeProtectorDetachedBackups(...func(ComponentMetadata, ComponentMetadata))

	InvokeOnSurgeProtectorTrip(ComponentMetadata)
	InvokeOnSurgeProtectorReset(ComponentMetadata)
	InvokeOnSurgeProtectorBackupFailure(ComponentMetadata, error)
	InvokeOnSurgeProtectorRateLimitExceeded(c ComponentMetadata, elem T)
	InvokeOnSurgeProtectorSubmit(c ComponentMetadata, elem T)
	InvokeOnResisterDequeued(c ComponentMetadata, elem T)
	InvokeOnResisterQueued(c ComponentMetadata, elem T)
	InvokeOnResisterRequeued(c ComponentMetadata, elem T)
	InvokeOnResisterEmpty(ComponentMetadata)
	InvokeOnSurgeProtectorReleaseToken(ComponentMetadata)
	InvokeOnSurgeProtectorConnectResister(ComponentMetadata, ComponentMetadata)
	InvokeOnSurgeProtectorDetachedBackups(ComponentMetadata, ComponentMetadata)
	InvokeOnSurgeProtectorBackupWireSubmit(ComponentMetadata, T)
	InvokeOnSurgeProtectorDrop(ComponentMetadata, T)

	// S3 writer lifecycle
	RegisterOnS3WriterStart(...func(ComponentMetadata, string /*bucket*/, string /*prefixTpl*/, string /*format*/))
	RegisterOnS3WriterStop(...func(ComponentMetadata))

	InvokeOnS3WriterStart(ComponentMetadata, string, string, string)
	InvokeOnS3WriterStop(ComponentMetadata)

	// Key generation
	RegisterOnS3KeyRendered(...func(ComponentMetadata, string /*key*/))
	InvokeOnS3KeyRendered(ComponentMetadata, string)

	// PutObject
	RegisterOnS3PutAttempt(...func(ComponentMetadata, string /*bucket*/, string /*key*/, int /*bytes*/, string /*sseMode*/, string /*kmsKey*/))
	RegisterOnS3PutSuccess(...func(ComponentMetadata, string /*bucket*/, string /*key*/, int /*bytes*/, time.Duration /*dur*/))
	RegisterOnS3PutError(...func(ComponentMetadata, string /*bucket*/, string /*key*/, int /*bytes*/, error))

	InvokeOnS3PutAttempt(ComponentMetadata, string, string, int, string, string)
	InvokeOnS3PutSuccess(ComponentMetadata, string, string, int, time.Duration)
	InvokeOnS3PutError(ComponentMetadata, string, string, int, error)

	// Parquet rolling (before upload)
	RegisterOnS3ParquetRollFlush(...func(ComponentMetadata, int /*records*/, int /*bytes*/, string /*compression*/))
	InvokeOnS3ParquetRollFlush(ComponentMetadata, int, int, string)

	// Reader lifecycle + pages
	RegisterOnS3ReaderListStart(...func(ComponentMetadata, string /*bucket*/, string /*prefix*/))
	RegisterOnS3ReaderListPage(...func(ComponentMetadata, int /*objsInPage*/, bool /*isTruncated*/))
	RegisterOnS3ReaderObject(...func(ComponentMetadata, string /*key*/, int64 /*size*/))
	RegisterOnS3ReaderDecode(...func(ComponentMetadata, string /*key*/, int /*rows*/, string /*format*/))
	RegisterOnS3ReaderSpillToDisk(...func(ComponentMetadata, int64 /*threshold*/, int64 /*objectBytes*/))
	RegisterOnS3ReaderComplete(...func(ComponentMetadata, int /*objectsScanned*/, int /*rowsDecoded*/))

	InvokeOnS3ReaderListStart(ComponentMetadata, string, string)
	InvokeOnS3ReaderListPage(ComponentMetadata, int, bool)
	InvokeOnS3ReaderObject(ComponentMetadata, string, int64)
	InvokeOnS3ReaderDecode(ComponentMetadata, string, int, string)
	InvokeOnS3ReaderSpillToDisk(ComponentMetadata, int64, int64)
	InvokeOnS3ReaderComplete(ComponentMetadata, int, int)

	// (Optional) billing sampling: raw inputs to cost model
	RegisterOnS3BillingSample(...func(ComponentMetadata, string /*op: PUT|GET|LIST*/, int64 /*requestUnits*/, int64 /*bytes*/, string /*storageClass*/))
	InvokeOnS3BillingSample(ComponentMetadata, string, int64, int64, string)
}
