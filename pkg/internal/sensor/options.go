package sensor

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithLogger registers loggers for the sensor.
func WithLogger[T any](logger ...types.Logger) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.ConnectLogger(logger...)
	}
}

// WithOnStartFunc registers start callbacks.
func WithOnStartFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnStart(callback...)
	}
}

// WithOnElementProcessedFunc registers processed callbacks.
func WithOnElementProcessedFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnElementProcessed(callback...)
	}
}

// WithOnErrorFunc registers error callbacks.
func WithOnErrorFunc[T any](callback ...func(types.ComponentMetadata, error, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnError(callback...)
	}
}

// WithOnCancelFunc registers cancel callbacks.
func WithOnCancelFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCancel(callback...)
	}
}

// WithOnCompleteFunc registers completion callbacks.
func WithOnCompleteFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnComplete(callback...)
	}
}

// WithOnStopFunc registers stop callbacks.
func WithOnStopFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnTerminate(callback...)
	}
}

// WithComponentMetadata sets the sensor name and id.
func WithComponentMetadata[T any](name string, id string) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.SetComponentMetadata(name, id)
	}
}

// WithOnHTTPClientRequestStartFunc registers HTTP request start callbacks.
func WithOnHTTPClientRequestStartFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientRequestStart(callback...)
	}
}

// WithOnHTTPClientResponseReceivedFunc registers HTTP response callbacks.
func WithOnHTTPClientResponseReceivedFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientResponseReceived(callback...)
	}
}

// WithOnHTTPClientErrorFunc registers HTTP error callbacks.
func WithOnHTTPClientErrorFunc[T any](callback ...func(types.ComponentMetadata, error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientError(callback...)
	}
}

// WithOnHTTPClientRequestCompleteFunc registers HTTP completion callbacks.
func WithOnHTTPClientRequestCompleteFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnHTTPClientRequestComplete(callback...)
	}
}

// WithSurgeProtectorTripFunc registers surge protector trip callbacks.
func WithSurgeProtectorTripFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorTrip(callback...)
	}
}

// WithSurgeProtectorResetFunc registers surge protector reset callbacks.
func WithSurgeProtectorResetFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorReset(callback...)
	}
}

// WithSurgeProtectorBackupFailureFunc registers surge protector backup failure callbacks.
func WithSurgeProtectorBackupFailureFunc[T any](callback ...func(types.ComponentMetadata, error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorBackupFailure(callback...)
	}
}

// WithSurgeProtectorBackupWireSubmissionFunc registers surge protector backup submission callbacks.
func WithSurgeProtectorBackupWireSubmissionFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorBackupWireSubmit(callback...)
	}
}

// WithSurgeProtectorDroppedSubmissionFunc registers surge protector drop callbacks.
func WithSurgeProtectorDroppedSubmissionFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorDrop(callback...)
	}
}

// WithSurgeProtectorSubmitFunc registers surge protector submission callbacks.
func WithSurgeProtectorSubmitFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorSubmit(callback...)
	}
}

// WithSurgeProtectorRateLimitExceededFunc registers rate limit callbacks.
func WithSurgeProtectorRateLimitExceededFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorRateLimitExceeded(callback...)
	}
}

// WithSurgeProtectorReleaseTokenFunc registers release token callbacks.
func WithSurgeProtectorReleaseTokenFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorReleaseToken(callback...)
	}
}

// WithSurgeProtectorConnectResisterFunc registers resister connection callbacks.
func WithSurgeProtectorConnectResisterFunc[T any](callback ...func(types.ComponentMetadata, types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorConnectResister(callback...)
	}
}

// WithSurgeProtectorDetachedBackupsFunc registers detached backups callbacks.
func WithSurgeProtectorDetachedBackupsFunc[T any](callback ...func(types.ComponentMetadata, types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSurgeProtectorDetachedBackups(callback...)
	}
}

// WithResisterDequeuedFunc registers resister dequeued callbacks.
func WithResisterDequeuedFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterDequeued(callback...)
	}
}

// WithResisterQueuedFunc registers resister queued callbacks.
func WithResisterQueuedFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterQueued(callback...)
	}
}

// WithResisterRequeuedFunc registers resister requeued callbacks.
func WithResisterRequeuedFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterRequeued(callback...)
	}
}

// WithResisterEmptyFunc registers resister empty callbacks.
func WithResisterEmptyFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnResisterEmpty(callback...)
	}
}

// WithOnSubmitFunc registers submit callbacks.
func WithOnSubmitFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnSubmit(callback...)
	}
}

// WithOnRestartFunc registers restart callbacks.
func WithOnRestartFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnRestart(callback...)
	}
}

// WithMeter registers meters with the sensor.
func WithMeter[T any](meter ...types.Meter[T]) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.ConnectMeter(meter...)
	}
}

// WithCircuitBreakerTripFunc registers circuit breaker trip callbacks.
func WithCircuitBreakerTripFunc[T any](callback ...func(types.ComponentMetadata, int64, int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerTrip(callback...)
	}
}

// WithCircuitBreakerResetFunc registers circuit breaker reset callbacks.
func WithCircuitBreakerResetFunc[T any](callback ...func(types.ComponentMetadata, int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerReset(callback...)
	}
}

// WithCircuitBreakerNeutralWireSubmissionFunc registers circuit breaker diversion callbacks.
func WithCircuitBreakerNeutralWireSubmissionFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerNeutralWireSubmission(callback...)
	}
}

// WithCircuitBreakerRecordErrorFunc registers circuit breaker error callbacks.
func WithCircuitBreakerRecordErrorFunc[T any](callback ...func(types.ComponentMetadata, int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerRecordError(callback...)
	}
}

// WithCircuitBreakerAllowFunc registers circuit breaker allow callbacks.
func WithCircuitBreakerAllowFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerAllow(callback...)
	}
}

// WithCircuitBreakerDropFunc registers circuit breaker drop callbacks.
func WithCircuitBreakerDropFunc[T any](callback ...func(types.ComponentMetadata, T)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnCircuitBreakerDrop(callback...)
	}
}

// WithOnS3WriterStartFunc registers S3 writer start callbacks.
func WithOnS3WriterStartFunc[T any](callback ...func(types.ComponentMetadata, string, string, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3WriterStart(callback...)
	}
}

// WithOnS3WriterStopFunc registers S3 writer stop callbacks.
func WithOnS3WriterStopFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3WriterStop(callback...)
	}
}

// WithOnS3KeyRenderedFunc registers S3 key rendered callbacks.
func WithOnS3KeyRenderedFunc[T any](callback ...func(types.ComponentMetadata, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3KeyRendered(callback...)
	}
}

// WithOnS3PutAttemptFunc registers S3 put attempt callbacks.
func WithOnS3PutAttemptFunc[T any](callback ...func(types.ComponentMetadata, string, string, int, string, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3PutAttempt(callback...)
	}
}

// WithOnS3PutSuccessFunc registers S3 put success callbacks.
func WithOnS3PutSuccessFunc[T any](callback ...func(types.ComponentMetadata, string, string, int, time.Duration)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3PutSuccess(callback...)
	}
}

// WithOnS3PutErrorFunc registers S3 put error callbacks.
func WithOnS3PutErrorFunc[T any](callback ...func(types.ComponentMetadata, string, string, int, error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3PutError(callback...)
	}
}

// WithOnS3ParquetRollFlushFunc registers parquet roll flush callbacks.
func WithOnS3ParquetRollFlushFunc[T any](callback ...func(types.ComponentMetadata, int, int, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3ParquetRollFlush(callback...)
	}
}

// WithOnS3ReaderListStartFunc registers S3 list start callbacks.
func WithOnS3ReaderListStartFunc[T any](callback ...func(types.ComponentMetadata, string, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3ReaderListStart(callback...)
	}
}

// WithOnS3ReaderListPageFunc registers S3 list page callbacks.
func WithOnS3ReaderListPageFunc[T any](callback ...func(types.ComponentMetadata, int, bool)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3ReaderListPage(callback...)
	}
}

// WithOnS3ReaderObjectFunc registers S3 reader object callbacks.
func WithOnS3ReaderObjectFunc[T any](callback ...func(types.ComponentMetadata, string, int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3ReaderObject(callback...)
	}
}

// WithOnS3ReaderDecodeFunc registers S3 reader decode callbacks.
func WithOnS3ReaderDecodeFunc[T any](callback ...func(types.ComponentMetadata, string, int, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3ReaderDecode(callback...)
	}
}

// WithOnS3ReaderSpillToDiskFunc registers S3 spill callbacks.
func WithOnS3ReaderSpillToDiskFunc[T any](callback ...func(types.ComponentMetadata, int64, int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3ReaderSpillToDisk(callback...)
	}
}

// WithOnS3ReaderCompleteFunc registers S3 reader complete callbacks.
func WithOnS3ReaderCompleteFunc[T any](callback ...func(types.ComponentMetadata, int, int)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3ReaderComplete(callback...)
	}
}

// WithOnS3BillingSampleFunc registers S3 billing callbacks.
func WithOnS3BillingSampleFunc[T any](callback ...func(types.ComponentMetadata, string, int64, int64, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnS3BillingSample(callback...)
	}
}

// WithOnKafkaWriterStartFunc registers Kafka writer start callbacks.
func WithOnKafkaWriterStartFunc[T any](callback ...func(types.ComponentMetadata, string, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaWriterStart(callback...)
	}
}

// WithOnKafkaWriterStopFunc registers Kafka writer stop callbacks.
func WithOnKafkaWriterStopFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaWriterStop(callback...)
	}
}

// WithOnKafkaProduceAttemptFunc registers Kafka produce attempt callbacks.
func WithOnKafkaProduceAttemptFunc[T any](callback ...func(types.ComponentMetadata, string, int, int, int)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaProduceAttempt(callback...)
	}
}

// WithOnKafkaProduceSuccessFunc registers Kafka produce success callbacks.
func WithOnKafkaProduceSuccessFunc[T any](callback ...func(types.ComponentMetadata, string, int, int64, time.Duration)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaProduceSuccess(callback...)
	}
}

// WithOnKafkaProduceErrorFunc registers Kafka produce error callbacks.
func WithOnKafkaProduceErrorFunc[T any](callback ...func(types.ComponentMetadata, string, int, error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaProduceError(callback...)
	}
}

// WithOnKafkaBatchFlushFunc registers Kafka batch flush callbacks.
func WithOnKafkaBatchFlushFunc[T any](callback ...func(types.ComponentMetadata, string, int, int, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaBatchFlush(callback...)
	}
}

// WithOnKafkaKeyRenderedFunc registers Kafka key rendered callbacks.
func WithOnKafkaKeyRenderedFunc[T any](callback ...func(types.ComponentMetadata, []byte)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaKeyRendered(callback...)
	}
}

// WithOnKafkaHeadersRenderedFunc registers Kafka headers rendered callbacks.
func WithOnKafkaHeadersRenderedFunc[T any](callback ...func(types.ComponentMetadata, []struct{ Key, Value string })) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaHeadersRendered(callback...)
	}
}

// WithOnKafkaConsumerStartFunc registers Kafka consumer start callbacks.
func WithOnKafkaConsumerStartFunc[T any](callback ...func(types.ComponentMetadata, string, []string, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaConsumerStart(callback...)
	}
}

// WithOnKafkaConsumerStopFunc registers Kafka consumer stop callbacks.
func WithOnKafkaConsumerStopFunc[T any](callback ...func(types.ComponentMetadata)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaConsumerStop(callback...)
	}
}

// WithOnKafkaPartitionAssignedFunc registers Kafka partition assigned callbacks.
func WithOnKafkaPartitionAssignedFunc[T any](callback ...func(types.ComponentMetadata, string, int, int64, int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaPartitionAssigned(callback...)
	}
}

// WithOnKafkaPartitionRevokedFunc registers Kafka partition revoked callbacks.
func WithOnKafkaPartitionRevokedFunc[T any](callback ...func(types.ComponentMetadata, string, int)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaPartitionRevoked(callback...)
	}
}

// WithOnKafkaMessageFunc registers Kafka message callbacks.
func WithOnKafkaMessageFunc[T any](callback ...func(types.ComponentMetadata, string, int, int64, int, int)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaMessage(callback...)
	}
}

// WithOnKafkaDecodeFunc registers Kafka decode callbacks.
func WithOnKafkaDecodeFunc[T any](callback ...func(types.ComponentMetadata, string, int, string)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaDecode(callback...)
	}
}

// WithOnKafkaCommitSuccessFunc registers Kafka commit success callbacks.
func WithOnKafkaCommitSuccessFunc[T any](callback ...func(types.ComponentMetadata, string, map[string]int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaCommitSuccess(callback...)
	}
}

// WithOnKafkaCommitErrorFunc registers Kafka commit error callbacks.
func WithOnKafkaCommitErrorFunc[T any](callback ...func(types.ComponentMetadata, string, error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaCommitError(callback...)
	}
}

// WithOnKafkaDLQProduceAttemptFunc registers Kafka DLQ produce attempt callbacks.
func WithOnKafkaDLQProduceAttemptFunc[T any](callback ...func(types.ComponentMetadata, string, int, int, int)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaDLQProduceAttempt(callback...)
	}
}

// WithOnKafkaDLQProduceSuccessFunc registers Kafka DLQ produce success callbacks.
func WithOnKafkaDLQProduceSuccessFunc[T any](callback ...func(types.ComponentMetadata, string, int, int64, time.Duration)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaDLQProduceSuccess(callback...)
	}
}

// WithOnKafkaDLQProduceErrorFunc registers Kafka DLQ produce error callbacks.
func WithOnKafkaDLQProduceErrorFunc[T any](callback ...func(types.ComponentMetadata, string, int, error)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaDLQProduceError(callback...)
	}
}

// WithOnKafkaBillingSampleFunc registers Kafka billing callbacks.
func WithOnKafkaBillingSampleFunc[T any](callback ...func(types.ComponentMetadata, string, int64, int64)) types.Option[types.Sensor[T]] {
	return func(m types.Sensor[T]) {
		m.RegisterOnKafkaBillingSample(callback...)
	}
}
