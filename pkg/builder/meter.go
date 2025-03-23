package builder

import (
	"context"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/meter"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// MetricName is a type alias for metric names used in the Meter.
type MetricName string

type MetricInfo = types.MetricInfo

// Here we re-export the constants from the meter package
const (
	MetricTransformPercentage                          MetricName = MetricName(types.MetricTransformPercentage)
	MetricErrorPercentage                              MetricName = MetricName(types.MetricErrorPercentage)
	MetricTransformsPerSecond                          MetricName = MetricName(types.MetricTransformsPerSecond)
	MetricTransformationErrorsPerSecond                MetricName = MetricName(types.MetricTransformationErrorsPerSecond)
	MetricTotalErrorCount                              MetricName = MetricName(types.MetricTotalErrorCount)
	MetricProgressPercentage                           MetricName = MetricName(types.MetricProgressPercentage)
	MetricTotalPendingCount                            MetricName = MetricName(types.MetricTotalPendingCount)
	MetricTotalProcessedCount                          MetricName = MetricName(types.MetricTotalProcessedCount)
	MetricTotalSubmittedCount                          MetricName = MetricName(types.MetricTotalSubmittedCount)
	MetricTotalTransformedCount                        MetricName = MetricName(types.MetricTotalTransformedCount)
	MetricComponentRunningCount                        MetricName = MetricName(types.MetricComponentRunningCount)
	MetricMeterConnectedComponentCount                 MetricName = MetricName(types.MetricMeterConnectedComponentCount)
	MetricLoggerConnectedComponentCount                MetricName = MetricName(types.MetricLoggerConnectedComponentCount)
	MetricComponentRestartCount                        MetricName = MetricName(types.MetricComponentRestartCount)
	MetricComponentLifecycleErrorCount                 MetricName = MetricName(types.MetricComponentLifecycleErrorCount)
	MetricWireRunningCount                             MetricName = MetricName(types.MetricWireRunningCount)
	MetricWireElementTransformCount                    MetricName = MetricName(types.MetricWireElementTransformCount)
	MetricWireElementTransformErrorCount               MetricName = MetricName(types.MetricWireElementTransformErrorCount)
	MetricWireElementSubmitCount                       MetricName = MetricName(types.MetricWireElementSubmitCount)
	MetricWireElementSubmitErrorCount                  MetricName = MetricName(types.MetricWireElementSubmitErrorCount)
	MetricWireElementSubmitCancelCount                 MetricName = MetricName(types.MetricWireElementSubmitCancelCount)
	MetricWireElementCancelCount                       MetricName = MetricName(types.MetricWireElementCancelCount)
	MetricWireElementErrorCount                        MetricName = MetricName(types.MetricWireElementErrorCount)
	MetricWireElementRecoverAttemptCount               MetricName = MetricName(types.MetricWireElementRecoverAttemptCount)
	MetricWireElementRecoverSuccessCount               MetricName = MetricName(types.MetricWireElementRecoverSuccessCount)
	MetricWireElementRecoverFailureCount               MetricName = MetricName(types.MetricWireElementRecoverFailureCount)
	MetricCircuitBreakerResetTimer                     MetricName = MetricName(types.MetricCircuitBreakerResetTimer)
	MetricCircuitBreakerLastErrorRecordTime            MetricName = MetricName(types.MetricCircuitBreakerLastErrorRecordTime)
	MetricCircuitBreakerLastTripTime                   MetricName = MetricName(types.MetricCircuitBreakerLastTripTime)
	MetricCircuitBreakerLastResetTime                  MetricName = MetricName(types.MetricCircuitBreakerLastResetTime)
	MetricCircuitBreakerNextResetTime                  MetricName = MetricName(types.MetricCircuitBreakerNextResetTime)
	MetricCircuitBreakerCurrentTripCount               MetricName = MetricName(types.MetricCircuitBreakerCurrentTripCount)
	MetricCircuitBreakerTripCount                      MetricName = MetricName(types.MetricCircuitBreakerTripCount)
	MetricCircuitBreakerResetCount                     MetricName = MetricName(types.MetricCircuitBreakerResetCount)
	MetricCircuitBreakerNeutralWireSubmissionCount     MetricName = MetricName(types.MetricCircuitBreakerNeutralWireSubmissionCount)
	MetricCircuitBreakerRecordedErrorCount             MetricName = MetricName(types.MetricCircuitBreakerRecordedErrorCount)
	MetricCircuitBreakerDroppedElementCount            MetricName = MetricName(types.MetricCircuitBreakerDroppedElementCount)
	MetricCircuitBreakerNeutralWireConnectedCount      MetricName = MetricName(types.MetricCircuitBreakerNeutralWireConnectedCount)
	MetricCircuitBreakerNeutralWireFailureCount        MetricName = MetricName(types.MetricCircuitBreakerNeutralWireFailureCount)
	MetricCircuitBreakerCount                          MetricName = MetricName(types.MetricCircuitBreakerCount)
	MetricTransformationErrorPercentage                MetricName = MetricName(types.MetricTransformationErrorPercentage)
	MetricHTTPRequestMadeCount                         MetricName = MetricName(types.MetricHTTPRequestMadeCount)
	MetricHTTPRequestReceivedCount                     MetricName = MetricName(types.MetricHTTPRequestReceivedCount)
	MetricHTTPRequestCompletedCount                    MetricName = MetricName(types.MetricHTTPRequestCompletedCount)
	MetricHTTPResponseCount                            MetricName = MetricName(types.MetricHTTPResponseCount)
	MetricHTTPClientErrorCount                         MetricName = MetricName(types.MetricHTTPClientErrorCount)
	MetricHTTPClientRetryCount                         MetricName = MetricName(types.MetricHTTPClientRetryCount)
	MetricHTTPOAuth2ClientCount                        MetricName = MetricName(types.MetricHTTPOAuth2ClientCount)
	MetricHTTPResponseErrorCount                       MetricName = MetricName(types.MetricHTTPResponseErrorCount)
	MetricHTTPClientWithTlsPinningCount                MetricName = MetricName(types.MetricHTTPClientWithTlsPinningCount)
	MetricHTTPClientFetchCount                         MetricName = MetricName(types.MetricHTTPClientFetchCount)
	MetricHTTPClientFetchErrorCount                    MetricName = MetricName(types.MetricHTTPClientFetchErrorCount)
	MetricHTTPClientFetchAndSubmitErrorCount           MetricName = MetricName(types.MetricHTTPClientFetchAndSubmitErrorCount)
	MetricHTTPClientWithActiveOAuthTokenCount          MetricName = MetricName(types.MetricHTTPClientWithActiveOAuthTokenCount)
	MetricHTTPClientOAuthTokenRequestCount             MetricName = MetricName(types.MetricHTTPClientOAuthTokenRequestCount)
	MetricHTTPClientFailCount                          MetricName = MetricName(types.MetricHTTPClientFailCount)
	MetricHTTPClientOAuthTokenErrorCount               MetricName = MetricName(types.MetricHTTPClientOAuthTokenErrorCount)
	MetricHTTPClientJsonUnmarshalErrorCount            MetricName = MetricName(types.MetricHTTPClientJsonUnmarshalErrorCount)
	MetricHTTPClientSuccessfulFetchCount               MetricName = MetricName(types.MetricHTTPClientSuccessfulFetchCount)
	MetricHTTPClientRetryExhaustedCount                MetricName = MetricName(types.MetricHTTPClientRetryExhaustedCount)
	MetricSurgeProtectorAttachedCount                  MetricName = MetricName(types.MetricSurgeProtectorAttachedCount)
	MetricSurgeProtectorBackupWireSubmissionCount      MetricName = MetricName(types.MetricSurgeProtectorBackupWireSubmissionCount)
	MetricSurgeProtectorBackupWireSubmissionPercentage MetricName = MetricName(types.MetricSurgeProtectorBackupWireSubmissionPercentage)
	MetricSurgeProtectorDropCount                      MetricName = MetricName(types.MetricSurgeProtectorDropCount)
	MetricSurgeCount                                   MetricName = MetricName(types.MetricSurgeCount)
	MetricSurgeTripCount                               MetricName = MetricName(types.MetricSurgeTripCount)
	MetricSurgeResetCount                              MetricName = MetricName(types.MetricSurgeResetCount)
	MetricSurgeBackupFailureCount                      MetricName = MetricName(types.MetricSurgeBackupFailureCount)
	MetricSurgeRateLimitExceedCount                    MetricName = MetricName(types.MetricSurgeRateLimitExceedCount)
	MetricSurgeBlackoutTripCount                       MetricName = MetricName(types.MetricSurgeBlackoutTripCount)
	MetricSurgeBlackoutResetCount                      MetricName = MetricName(types.MetricSurgeBlackoutResetCount)
	MetricResisterElementRequeued                      MetricName = MetricName(types.MetricResisterElementRequeued)
	MetricResisterElementQueuedCount                   MetricName = MetricName(types.MetricResisterElementQueuedCount)
	MetricResisterElementCurrentlyQueuedCount          MetricName = MetricName(types.MetricResisterElementCurrentlyQueuedCount)
	MetricResisterElementDequeued                      MetricName = MetricName(types.MetricResisterElementDequeued)
	MetricReceivingRelayListeningCount                 MetricName = MetricName(types.MetricReceivingRelayListeningCount)
	MetricReceivingRelayRunningCount                   MetricName = MetricName(types.MetricReceivingRelayRunningCount)
	MetricReceivingRelayReceivedCount                  MetricName = MetricName(types.MetricReceivingRelayReceivedCount)
	MetricReceivingRelayUnwrappedPayloadCount          MetricName = MetricName(types.MetricReceivingRelayUnwrappedPayloadCount)
	MetricReceivingRelayRelayedCount                   MetricName = MetricName(types.MetricReceivingRelayRelayedCount)
	MetricReceivingRelayUnwrappedPayloadErrorCount     MetricName = MetricName(types.MetricReceivingRelayUnwrappedPayloadErrorCount)
	MetricReceivingRelayStreamReceiveCount             MetricName = MetricName(types.MetricReceivingRelayStreamReceiveCount)
	MetricReceivingRelayTargetCount                    MetricName = MetricName(types.MetricReceivingRelayTargetCount)
	MetricReceivingRelayErrorCount                     MetricName = MetricName(types.MetricReceivingRelayErrorCount)
	MetricReceivingRelayNoTLSCount                     MetricName = MetricName(types.MetricReceivingRelayNoTLSCount)
	MetricProcessDuration                              MetricName = MetricName(types.MetricProcessDuration)
	MetricGeneratorRunningCount                        MetricName = MetricName(types.MetricGeneratorRunningCount)
	MetricGeneratorSubmitCount                         MetricName = MetricName(types.MetricGeneratorSubmitCount)
	MetricGeneratorSubmitErrorCount                    MetricName = MetricName(types.MetricGeneratorSubmitErrorCount)
	MetricForwardRelaySubmitErrorCount                 MetricName = MetricName(types.MetricForwardRelaySubmitErrorCount)
	MetricForwardRelaySubmitCount                      MetricName = MetricName(types.MetricForwardRelaySubmitCount)
	MetricForwardRelayRelayedCount                     MetricName = MetricName(types.MetricForwardRelayRelayedCount)
	MetricForwardRelayWithPerformanceOptionCount       MetricName = MetricName(types.MetricForwardRelayWithPerformanceOptionCount)
	MetricForwardRelayWithTlsConfigCount               MetricName = MetricName(types.MetricForwardRelayWithTlsConfigCount)
	MetricForwardRelayRunningCount                     MetricName = MetricName(types.MetricForwardRelayRunningCount)
	MetricForwardRelayErrorCount                       MetricName = MetricName(types.MetricForwardRelayErrorCount)
	MetricForwardRelayPayloadCompressionCount          MetricName = MetricName(types.MetricForwardRelayPayloadCompressionCount)
	MetricForwardRelayPayloadCompressionErrorCount     MetricName = MetricName(types.MetricForwardRelayPayloadCompressionErrorCount)
	MetricForwardRelayWrappedPayloadCount              MetricName = MetricName(types.MetricForwardRelayWrappedPayloadCount)
	MetricForwardRelayInputCount                       MetricName = MetricName(types.MetricForwardRelayInputCount)
)

func NewMeter[T any](ctx context.Context, options ...types.Option[types.Meter[T]]) types.Meter[T] {
	return meter.NewMeter[T](ctx, options...)
}

// WithMetricTotal sets an initial total for a specific metric.
func MeterWithIdleTimeout[T any](to time.Duration) types.Option[types.Meter[T]] {
	return meter.WithIdleTimeout[T](to)
}

// Use MetricName for metric name parameters
func MeterWithMetricTotal[T any](metricName MetricName, total uint64) types.Option[types.Meter[T]] {
	return meter.WithMetricTotal[T](string(metricName), total)
}

func MeterWithInitialMetricCount[T any](metricName MetricName, count uint64) types.Option[types.Meter[T]] {
	return meter.WithInitialMetricCount[T](string(metricName), count)
}

func MeterWithErrorThreshold[T any](threshold float64) types.Option[types.Meter[T]] {
	return meter.WithErrorThreshold[T](threshold)
}

func MeterWithTimerStart[T any](metricName MetricName, startTime time.Time) types.Option[types.Meter[T]] {
	return meter.WithTimerStart[T](string(metricName), startTime)
}

func MeterWithTotalItems[T any](total uint64) types.Option[types.Meter[T]] {
	return meter.WithTotalItems[T](total)
}

func MeterWithComponentMetadata[T any](name string, id string) types.Option[types.Meter[T]] {
	return meter.WithComponentMetadata[T](name, id)
}

func MeterWithLogger[T any](loggers ...types.Logger) types.Option[types.Meter[T]] {
	return meter.WithLogger[T](loggers...)
}

// WithDynamicMetric adds a new metric to be monitored, with optional initial values.
func MeterWithDynamicMetric[T any](metricName MetricName, total uint64, initialCount uint64, threshold float64) types.Option[types.Meter[T]] {
	return meter.WithDynamicMetric[T](string(metricName), total, initialCount, threshold)
}

// WithUpdateFrequency sets the frequency at which the Meter updates its readings.
func MeterWithUpdateFrequency[T any](duration time.Duration) types.Option[types.Meter[T]] {
	return meter.WithUpdateFrequency[T](duration)
}

// WithCancellationHook adds a custom function to be called upon cancellation.
func MeterWithCancellationHook[T any](hook func()) types.Option[types.Meter[T]] {
	return meter.WithCancellationHook[T](hook)
}
