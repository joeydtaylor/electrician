package meter

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

func (m *Meter[T]) resetRunningMetrics() {
	m.SetMetricCount(types.MetricComponentRunningCount, 0)
	m.SetMetricCount(types.MetricWireRunningCount, 0)
}

func (m *Meter[T]) initializeMetrics() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	metricTypes := map[string]types.MetricInfo{
		types.MetricCurrentCpuPercentage: {
			Name:          types.MetricCurrentCpuPercentage,
			DisplayAs:     "Current CPU",
			ThresholdType: types.Percentage,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCurrentRamPercentage: {
			Name:          types.MetricCurrentRamPercentage,
			DisplayAs:     "Current RAM",
			ThresholdType: types.Percentage,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricPeakGoRoutinesActive: {
			Name:          types.MetricPeakGoRoutinesActive,
			DisplayAs:     "Peak Go Routines Active",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCurrentGoRoutinesActive: {
			Name:          types.MetricCurrentGoRoutinesActive,
			DisplayAs:     "Current Go Routines Active",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricPeakProcessedPerSecond: {
			Name:          types.MetricPeakProcessedPerSecond,
			DisplayAs:     "Peak Processed Per Second",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricPeakTransformedPerSecond: {
			Name:          types.MetricPeakTransformedPerSecond,
			DisplayAs:     "Peak Transformed Per Second",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricPeakTransformationErrorsPerSecond: {
			Name:          types.MetricPeakTransformationErrorsPerSecond,
			DisplayAs:     "Maximum Error Per Second",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricProcessedPerSecond: {
			Name:          types.MetricProcessedPerSecond,
			DisplayAs:     "Processed Per Second",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTransformationErrorsPerSecond: {
			Name:          types.MetricTransformationErrorsPerSecond,
			DisplayAs:     "Transformation Errors Per Second",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTransformsPerSecond: {
			Name:          types.MetricTransformsPerSecond,
			DisplayAs:     "Transforms Per Second",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTotalPendingCount: {
			Name:          types.MetricTotalPendingCount,
			DisplayAs:     "Total Pending Items",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTotalComponentsUsedCount: {
			Name:          types.MetricTotalComponentsUsedCount,
			DisplayAs:     "Total Components Used",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricSurgeProtectorCurrentTripCount: {
			Name:          types.MetricSurgeProtectorCurrentTripCount,
			DisplayAs:     "Active Tripped Surge Protectors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTransformationErrorPercentage: {
			Name:          types.MetricTransformationErrorPercentage,
			DisplayAs:     "Total Transform Error Percentage",
			ThresholdType: types.Percentage,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTransformPercentage: {
			Name:          types.MetricTransformPercentage,
			DisplayAs:     "Total Transform Success Percentage",
			ThresholdType: types.Percentage,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricErrorPercentage: {
			Name:          types.MetricErrorPercentage,
			DisplayAs:     "Total Error Percentage",
			ThresholdType: types.Percentage,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTotalErrorCount: {
			Name:          types.MetricTotalErrorCount,
			DisplayAs:     "Total Error Count",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricProgressPercentage: {
			Name:          types.MetricProgressPercentage,
			DisplayAs:     "Progress",
			ThresholdType: types.Percentage,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTotalProcessedCount: {
			Name:          types.MetricTotalProcessedCount,
			DisplayAs:     "Total Items Processed",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTotalSubmittedCount: {
			Name:          types.MetricTotalSubmittedCount,
			DisplayAs:     "Total Items Submitted",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricTotalTransformedCount: {
			Name:          types.MetricTotalTransformedCount,
			DisplayAs:     "Total Items Transformed",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricMeterConnectedComponentCount: {
			Name:          types.MetricMeterConnectedComponentCount,
			DisplayAs:     "Meter Connected Components",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricComponentRunningCount: {
			Name:          types.MetricComponentRunningCount,
			DisplayAs:     "Active Components",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricLoggerConnectedComponentCount: {
			Name:          types.MetricLoggerConnectedComponentCount,
			DisplayAs:     "Logger Connected Components",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricComponentRestartCount: {
			Name:          types.MetricComponentRestartCount,
			DisplayAs:     "Component Restarts",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricComponentLifecycleErrorCount: {
			Name:          types.MetricComponentLifecycleErrorCount,
			DisplayAs:     "Component Lifecycle Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementTransformCount: {
			Name:          types.MetricWireElementTransformCount,
			DisplayAs:     "Wire Transform Operations",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementTransformErrorCount: {
			Name:          types.MetricWireElementTransformErrorCount,
			DisplayAs:     "Wire Transform Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireRunningCount: {
			Name:          types.MetricWireRunningCount,
			DisplayAs:     "Wires Currently Active",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementSubmitCount: {
			Name:          types.MetricWireElementSubmitCount,
			DisplayAs:     "Items Submitted via Wire",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementSubmitErrorCount: {
			Name:          types.MetricWireElementSubmitErrorCount,
			DisplayAs:     "Wire Submission Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementSubmitCancelCount: {
			Name:          types.MetricWireElementSubmitCancelCount,
			DisplayAs:     "Wire Submission Cancellations",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementCancelCount: {
			Name:          types.MetricWireElementCancelCount,
			DisplayAs:     "Wire Cancellations",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementErrorCount: {
			Name:          types.MetricWireElementErrorCount,
			DisplayAs:     "Wire Processing Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementRecoverAttemptCount: {
			Name:          types.MetricWireElementRecoverAttemptCount,
			DisplayAs:     "Wire Recovery Attempts",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementRecoverSuccessCount: {
			Name:          types.MetricWireElementRecoverSuccessCount,
			DisplayAs:     "Successful Wire Recoveries",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricWireElementRecoverFailureCount: {
			Name:          types.MetricWireElementRecoverFailureCount,
			DisplayAs:     "Failed Wire Recoveries",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerResetTimer: {
			Name:          types.MetricCircuitBreakerResetTimer,
			DisplayAs:     "Circuit Breaker Reset Timer",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerLastTripTime: {
			Name:          types.MetricCircuitBreakerLastTripTime,
			DisplayAs:     "Circuit Breaker Last Trip Time",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerLastErrorRecordTime: {
			Name:          types.MetricCircuitBreakerLastErrorRecordTime,
			DisplayAs:     "Circuit Breaker Last Recorded Error Time",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerLastResetTime: {
			Name:          types.MetricCircuitBreakerLastResetTime,
			DisplayAs:     "Circuit Breaker Last Reset Time",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerNextResetTime: {
			Name:          types.MetricCircuitBreakerNextResetTime,
			DisplayAs:     "Circuit Breaker Next Reset Time",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerCurrentTripCount: {
			Name:          types.MetricCircuitBreakerCurrentTripCount,
			DisplayAs:     "Current Tripped Circuit Breakers",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerTripCount: {
			Name:          types.MetricCircuitBreakerTripCount,
			DisplayAs:     "Total Circuit Breaker Trips",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerResetCount: {
			Name:          types.MetricCircuitBreakerResetCount,
			DisplayAs:     "Circuit Breaker Resets",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerNeutralWireSubmissionCount: {
			Name:          types.MetricCircuitBreakerNeutralWireSubmissionCount,
			DisplayAs:     "Elements Diverted by Circuit Breaker",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerRecordedErrorCount: {
			Name:          types.MetricCircuitBreakerRecordedErrorCount,
			DisplayAs:     "Errors Recorded by Circuit Breaker",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerDroppedElementCount: {
			Name:          types.MetricCircuitBreakerDroppedElementCount,
			DisplayAs:     "Elements Dropped by Circuit Breaker",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerNeutralWireConnectedCount: {
			Name:          types.MetricCircuitBreakerNeutralWireConnectedCount,
			DisplayAs:     "Ground Wires Running",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerNeutralWireFailureCount: {
			Name:          types.MetricCircuitBreakerNeutralWireFailureCount,
			DisplayAs:     "Ground Wire Failures",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricCircuitBreakerCount: {
			Name:          types.MetricCircuitBreakerCount,
			DisplayAs:     "Circuit Breaker Count",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPRequestMadeCount: {
			Name:          types.MetricHTTPRequestMadeCount,
			DisplayAs:     "HTTP Requests Made",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPRequestReceivedCount: {
			Name:          types.MetricHTTPRequestReceivedCount,
			DisplayAs:     "HTTP Requests Received",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPRequestCompletedCount: {
			Name:          types.MetricHTTPRequestCompletedCount,
			DisplayAs:     "HTTP Requests Completed",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPResponseCount: {
			Name:          types.MetricHTTPResponseCount,
			DisplayAs:     "HTTP Responses",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientErrorCount: {
			Name:          types.MetricHTTPClientErrorCount,
			DisplayAs:     "HTTP Client Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientRetryCount: {
			Name:          types.MetricHTTPClientRetryCount,
			DisplayAs:     "HTTP Client Retries",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPOAuth2ClientCount: {
			Name:          types.MetricHTTPOAuth2ClientCount,
			DisplayAs:     "OAuth2 Clients Active",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPResponseErrorCount: {
			Name:          types.MetricHTTPResponseErrorCount,
			DisplayAs:     "HTTP Response Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientWithTlsPinningCount: {
			Name:          types.MetricHTTPClientWithTlsPinningCount,
			DisplayAs:     "HTTP Clients with TLS Pinning",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientFetchCount: {
			Name:          types.MetricHTTPClientFetchCount,
			DisplayAs:     "HTTP Client Fetch Operations",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientFetchErrorCount: {
			Name:          types.MetricHTTPClientFetchErrorCount,
			DisplayAs:     "HTTP Client Fetch Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientFetchAndSubmitErrorCount: {
			Name:          types.MetricHTTPClientFetchAndSubmitErrorCount,
			DisplayAs:     "HTTP Fetch and Submit Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientWithActiveOAuthTokenCount: {
			Name:          types.MetricHTTPClientWithActiveOAuthTokenCount,
			DisplayAs:     "Active OAuth Tokens",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientOAuthTokenRequestCount: {
			Name:          types.MetricHTTPClientOAuthTokenRequestCount,
			DisplayAs:     "OAuth Token Requests",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientOAuthTokenErrorCount: {
			Name:          types.MetricHTTPClientOAuthTokenErrorCount,
			DisplayAs:     "OAuth Token Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientJsonUnmarshalErrorCount: {
			Name:          types.MetricHTTPClientJsonUnmarshalErrorCount,
			DisplayAs:     "JSON Unmarshalling Errors",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientSuccessfulFetchCount: {
			Name:          types.MetricHTTPClientSuccessfulFetchCount,
			DisplayAs:     "Successful HTTP Fetches",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricHTTPClientRetryExhaustedCount: {
			Name:          types.MetricHTTPClientRetryExhaustedCount,
			DisplayAs:     "HTTP Retries Exhausted",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricSurgeProtectorBackupWireSubmissionCount: {
			Name:          types.MetricSurgeProtectorBackupWireSubmissionCount,
			DisplayAs:     "Surge Protector Backup Wire Submissions",
			ThresholdType: types.Absolute,
			Total:         0,
			Count:         new(uint64),
			Timestamp:     time.Now().Unix(),
		},
		types.MetricSurgeProtectorBackupWireSubmissionPercentage: {
			Name:          types.MetricSurgeProtectorBackupWireSubmissionPercentage,
			DisplayAs:     "Surge Protector Backup Wire Percentage",
			ThresholdType: types.Percentage,
			Total:         0,
			Count:         new(uint64),
		},
	}

	for name, info := range metricTypes {
		if _, exists := m.metrics[name]; !exists {
			m.metrics[name] = &types.MetricInfo{
				Name:          info.Name,
				DisplayAs:     info.DisplayAs,
				ThresholdType: info.ThresholdType,
				Count:         new(uint64),

				Timestamp: time.Now().Unix(),
			}
		}
	}

	metricDisplayNames := map[string]string{
		types.MetricHTTPClientFetchErrorCount:                    "HTTP Client Fetch Errors",
		types.MetricHTTPClientFetchAndSubmitErrorCount:           "HTTP Fetch and Submit Errors",
		types.MetricHTTPClientWithActiveOAuthTokenCount:          "Active OAuth Tokens",
		types.MetricHTTPClientOAuthTokenRequestCount:             "OAuth Token Requests",
		types.MetricHTTPClientOAuthTokenErrorCount:               "OAuth Token Errors",
		types.MetricHTTPClientJsonUnmarshalErrorCount:            "JSON Unmarshalling Errors",
		types.MetricHTTPClientSuccessfulFetchCount:               "Successful HTTP Fetches",
		types.MetricHTTPClientRetryExhaustedCount:                "HTTP Retries Exhausted",
		types.MetricSurgeProtectorBackupWireSubmissionCount:      "Surge Protector Backup Wire Submissions",
		types.MetricSurgeProtectorBackupWireSubmissionPercentage: "Surge Protector Backup Wire Percentage",
		types.MetricSurgeProtectorAttachedCount:                  "Surge Protectors Attached",
		types.MetricSurgeTripCount:                               "Surge Protector Trips",
		types.MetricSurgeCount:                                   "Surge Protector Count",
		types.MetricSurgeResetCount:                              "Surge Protector Resets",
		types.MetricSurgeBackupFailureCount:                      "Surge Protector Backup Failures",
		types.MetricSurgeRateLimitExceedCount:                    "Surge Rate Limit Exceedances",
		types.MetricSurgeBlackoutTripCount:                       "Surge Blackout Trips",
		types.MetricSurgeBlackoutResetCount:                      "Surge Blackout Resets",
		types.MetricResisterElementRequeued:                      "Resister Elements Requeued",
		types.MetricResisterElementQueuedCount:                   "Resister Total Queued Count",
		types.MetricResisterElementCurrentlyQueuedCount:          "Resister Currently Queued",
		types.MetricResisterElementDequeued:                      "Resister Elements Dequeued",
		types.MetricReceivingRelayListeningCount:                 "Relays Currently Listening",
		types.MetricReceivingRelayRunningCount:                   "Relays Currently Running",
		types.MetricReceivingRelayReceivedCount:                  "Relays Received Count",
		types.MetricReceivingRelayUnwrappedPayloadCount:          "Relays Unwrapped Payloads",
		types.MetricReceivingRelayRelayedCount:                   "Relays Relayed Messages",
		types.MetricReceivingRelayUnwrappedPayloadErrorCount:     "Relays Payload Unwrap Errors",
		types.MetricReceivingRelayStreamReceiveCount:             "Relays Stream Received",
		types.MetricReceivingRelayTargetCount:                    "Relay Targets Count",
		types.MetricReceivingRelayErrorCount:                     "Relay Errors",
		types.MetricReceivingRelayNoTLSCount:                     "Relays without TLS",
		types.MetricProcessDuration:                              "Process Duration",
		types.MetricGeneratorRunningCount:                        "Generators Currently Active",
		types.MetricGeneratorSubmitCount:                         "Generator Submissions",
		types.MetricGeneratorSubmitErrorCount:                    "Generator Submission Errors",
		types.MetricForwardRelaySubmitErrorCount:                 "Forward Relay Submission Errors",
		types.MetricForwardRelaySubmitCount:                      "Forward Relay Submissions",
		types.MetricForwardRelayRelayedCount:                     "Forward Relay Relayed Messages",
		types.MetricForwardRelayWithPerformanceOptionCount:       "Forward Relays with Performance Options",
		types.MetricForwardRelayWithTlsConfigCount:               "Forward Relays with TLS Configuration",
		types.MetricForwardRelayRunningCount:                     "Forward Relays Running",
		types.MetricForwardRelayErrorCount:                       "Forward Relay Errors",
		types.MetricForwardRelayPayloadCompressionCount:          "Payload Compressions",
		types.MetricForwardRelayPayloadCompressionErrorCount:     "Payload Compression Errors",
		types.MetricForwardRelayWrappedPayloadCount:              "Wrapped Payloads",
		types.MetricForwardRelayInputCount:                       "Forward Relay Inputs",
		types.MetricConduitRunningCount:                          "Conduits Running",
	}

	// Initialize all metric counters and their display names
	for name, displayName := range metricDisplayNames {
		if _, exists := m.metrics[name]; !exists {
			m.metrics[name] = &types.MetricInfo{
				Count:     new(uint64),
				DisplayAs: displayName,
				Name:      name,
			}
		}
	}
	// Initialize all metric counters here
	metricNames := []string{
		types.MetricTransformsPerSecond,
		types.MetricTransformationErrorsPerSecond,
		types.MetricTotalErrorCount,
		types.MetricTransformPercentage,
		types.MetricErrorPercentage,
		types.MetricProgressPercentage,
		types.MetricTotalProcessedCount,
		types.MetricTotalSubmittedCount,
		types.MetricTotalTransformedCount,
		types.MetricMeterConnectedComponentCount,
		types.MetricComponentRunningCount,
		types.MetricMeterConnectedComponentCount,
		types.MetricLoggerConnectedComponentCount,
		types.MetricComponentRestartCount,
		types.MetricComponentLifecycleErrorCount,
		types.MetricWireElementTransformCount,
		types.MetricWireElementTransformErrorCount,
		types.MetricWireRunningCount,
		types.MetricWireElementSubmitCount,
		types.MetricWireElementSubmitErrorCount,
		types.MetricWireElementSubmitCancelCount,
		types.MetricWireElementCancelCount,
		types.MetricWireElementErrorCount,
		types.MetricWireElementRecoverAttemptCount,
		types.MetricWireElementRecoverSuccessCount,
		types.MetricWireElementRecoverFailureCount,
		types.MetricCircuitBreakerResetTimer,
		types.MetricCircuitBreakerLastErrorRecordTime,
		types.MetricCircuitBreakerLastResetTime,
		types.MetricCircuitBreakerNextResetTime,
		types.MetricCircuitBreakerCurrentTripCount,
		types.MetricCircuitBreakerTripCount,
		types.MetricCircuitBreakerResetCount,
		types.MetricCircuitBreakerNeutralWireSubmissionCount,
		types.MetricCircuitBreakerRecordedErrorCount,
		types.MetricCircuitBreakerDroppedElementCount,
		types.MetricCircuitBreakerNeutralWireConnectedCount,
		types.MetricCircuitBreakerNeutralWireFailureCount,
		types.MetricCircuitBreakerCount,
		types.MetricTransformationErrorPercentage,
		types.MetricHTTPRequestMadeCount,
		types.MetricHTTPRequestReceivedCount,
		types.MetricHTTPRequestCompletedCount,
		types.MetricHTTPResponseCount,
		types.MetricHTTPClientErrorCount,
		types.MetricHTTPClientRetryCount,
		types.MetricHTTPOAuth2ClientCount,
		types.MetricHTTPResponseErrorCount,
		types.MetricHTTPClientWithTlsPinningCount,
		types.MetricHTTPClientFetchCount,
		types.MetricHTTPClientFetchErrorCount,
		types.MetricHTTPClientFetchAndSubmitErrorCount,
		types.MetricHTTPClientWithActiveOAuthTokenCount,
		types.MetricHTTPClientOAuthTokenRequestCount,
		types.MetricHTTPClientFailCount,
		types.MetricHTTPClientOAuthTokenErrorCount,
		types.MetricHTTPClientJsonUnmarshalErrorCount,
		types.MetricHTTPClientSuccessfulFetchCount,
		types.MetricHTTPClientRetryExhaustedCount,
		types.MetricSurgeProtectorBackupWireSubmissionCount,
		types.MetricSurgeProtectorBackupWireSubmissionPercentage,
		types.MetricSurgeProtectorAttachedCount,
		types.MetricSurgeCount,
		types.MetricSurgeProtectorCurrentTripCount,
		types.MetricSurgeTripCount,
		types.MetricSurgeResetCount,
		types.MetricSurgeBackupFailureCount,
		types.MetricSurgeRateLimitExceedCount,
		types.MetricSurgeBlackoutTripCount,
		types.MetricSurgeBlackoutResetCount,
		types.MetricResisterElementRequeued,
		types.MetricResisterElementQueuedCount,
		types.MetricResisterElementCurrentlyQueuedCount,
		types.MetricResisterElementDequeued,
		types.MetricReceivingRelayListeningCount,
		types.MetricReceivingRelayRunningCount,
		types.MetricReceivingRelayReceivedCount,
		types.MetricReceivingRelayUnwrappedPayloadCount,
		types.MetricReceivingRelayRelayedCount,
		types.MetricReceivingRelayUnwrappedPayloadErrorCount,
		types.MetricReceivingRelayStreamReceiveCount,
		types.MetricReceivingRelayTargetCount,
		types.MetricReceivingRelayErrorCount,
		types.MetricReceivingRelayNoTLSCount,
		types.MetricProcessDuration,
		types.MetricGeneratorRunningCount,
		types.MetricGeneratorSubmitCount,
		types.MetricGeneratorSubmitErrorCount,
		types.MetricForwardRelaySubmitErrorCount,
		types.MetricForwardRelaySubmitCount,
		types.MetricForwardRelayRelayedCount,
		types.MetricForwardRelayWithPerformanceOptionCount,
		types.MetricForwardRelayWithTlsConfigCount,
		types.MetricForwardRelayRunningCount,
		types.MetricForwardRelayErrorCount,
		types.MetricForwardRelayPayloadCompressionCount,
		types.MetricForwardRelayPayloadCompressionErrorCount,
		types.MetricForwardRelayWrappedPayloadCount,
		types.MetricForwardRelayInputCount,
		types.MetricConduitRunningCount,
	}
	for _, name := range metricNames {
		m.metricNames = append(m.metricNames, name)
		m.counts[name] = new(uint64)
	}
}

func (m *Meter[T]) updateDisplay() {
	currentTime := time.Now()
	elapsedTime := currentTime.Sub(m.startTime).Seconds() // Get elapsed time in seconds

	// Get the counts for metrics
	totalSubmitted := float64(m.GetMetricCount(types.MetricTotalSubmittedCount))
	totalTransformed := float64(m.GetMetricCount(types.MetricTotalTransformedCount))
	// If you actually track total errors in "MetricTotalErrorCount"
	// ensure that's the right name. Adjust as needed if you store errors differently.
	totalErrors := float64(m.GetMetricCount(types.MetricTotalErrorCount))

	// Calculate per second metrics
	transformsPerSecond := 0.0
	errorsPerSecond := 0.0
	if elapsedTime > 0 {
		transformsPerSecond = totalTransformed / elapsedTime
		errorsPerSecond = totalErrors / elapsedTime

		// Update the "transforms_per_second" and "transformation_errors_per_second" metrics
		m.SetMetricCount(types.MetricTransformsPerSecond, uint64(transformsPerSecond))
		m.SetMetricCount(types.MetricTransformationErrorsPerSecond, uint64(errorsPerSecond))
	}

	// ------------------------------------------------------------------------
	// NEW CODE: Ensure "peak_processed_per_second" is updated
	// "processedPerSecond" = transformsPerSecond + errorsPerSecond
	// ------------------------------------------------------------------------
	processedPerSecond := transformsPerSecond + errorsPerSecond

	oldPeak := float64(m.GetMetricCount(types.MetricPeakProcessedPerSecond))
	if processedPerSecond > oldPeak {
		m.SetMetricCount(types.MetricPeakProcessedPerSecond, uint64(processedPerSecond))
	}
	// ------------------------------------------------------------------------

	// Calculate percentages if we have a known totalSubmitted
	if totalSubmitted > 0 {
		// Measure CPU usage
		cpuPercentages, _ := cpu.Percent(time.Millisecond*500, false)
		memStats, _ := mem.VirtualMemory()

		transformPercentage := (totalTransformed / totalSubmitted) * 100
		errorPercentage := (totalErrors / totalSubmitted) * 100
		backupWirePercentage := (float64(m.GetMetricCount(types.MetricSurgeProtectorBackupWireSubmissionCount)) / totalSubmitted) * 100

		m.SetMetricPercentage(types.MetricTransformPercentage, transformPercentage)
		m.SetMetricPercentage(types.MetricErrorPercentage, errorPercentage)

		// Current CPU usage and Peak CPU usage
		m.SetMetricPercentage(types.MetricCurrentCpuPercentage, cpuPercentages[0])
		m.SetMetricPeakPercentage(types.MetricCurrentCpuPercentage, cpuPercentages[0])

		// Current RAM usage and Peak RAM usage
		m.SetMetricPercentage(types.MetricCurrentRamPercentage, memStats.UsedPercent)
		m.SetMetricPeakPercentage(types.MetricCurrentRamPercentage, memStats.UsedPercent)

		// Go Routines
		m.SetMetricCount(types.MetricCurrentGoRoutinesActive, uint64(runtime.NumGoroutine()))
		m.SetMetricPeak(types.MetricPeakGoRoutinesActive, uint64(runtime.NumGoroutine()))

		// Also update the "peak" metrics for throughput if desired:
		m.SetMetricPeak(types.MetricPeakTransformedPerSecond, uint64(transformsPerSecond))
		m.SetMetricPeak(types.MetricPeakTransformationErrorsPerSecond, uint64(errorsPerSecond))

		m.SetMetricPercentage(types.MetricSurgeProtectorBackupWireSubmissionPercentage, backupWirePercentage)
	}

	// Clear previous output and move cursor up
	fmt.Printf("\033[2J\033[H") // Clear the screen and move cursor to the top

	// Display start/elapsed times
	fmt.Printf("Start Time: %v, Elapsed Time: %s\n",
		m.startTime.Format("01-02-2006 15:04:05"),
		time.Duration(elapsedTime)*time.Second,
	)

	// Print Go routines & CPU/RAM usage (current & peak)
	fmt.Printf("%s: %d, %s: %d, %s: %.2f%%, Peak: %.2f%%, %s: %.2f%%, Peak: %.2f%%\n",
		m.GetMetricDisplayName(types.MetricCurrentGoRoutinesActive),
		m.GetMetricCount(types.MetricCurrentGoRoutinesActive),
		m.GetMetricDisplayName(types.MetricPeakGoRoutinesActive),
		m.GetMetricCount(types.MetricPeakGoRoutinesActive),
		m.GetMetricDisplayName(types.MetricCurrentCpuPercentage),
		m.GetMetricPercentage(types.MetricCurrentCpuPercentage),
		m.GetMetricPeakPercentage(types.MetricCurrentCpuPercentage),
		m.GetMetricDisplayName(types.MetricCurrentRamPercentage),
		m.GetMetricPercentage(types.MetricCurrentRamPercentage),
		m.GetMetricPeakPercentage(types.MetricCurrentRamPercentage),
	)

	// If you have a totalItems value, show % processed
	if m.totalItems != 0 {
		fmt.Printf("Total Expected Elements: %d, Progress: %d%%\n",
			m.totalItems,
			int(float64(m.GetMetricCount(types.MetricTotalProcessedCount))/float64(m.totalItems)*100),
		)
	}

	// Show total submitted, current throughput, and peak throughput
	fmt.Printf("%s: %d, %s: %d, %s: %d\n",
		m.GetMetricDisplayName(types.MetricTotalSubmittedCount),
		m.GetMetricCount(types.MetricTotalSubmittedCount),
		m.GetMetricDisplayName(types.MetricProcessedPerSecond),
		(m.GetMetricCount(types.MetricTransformsPerSecond) + m.GetMetricCount(types.MetricTransformationErrorsPerSecond)),
		m.GetMetricDisplayName(types.MetricPeakProcessedPerSecond),
		m.GetMetricCount(types.MetricPeakProcessedPerSecond),
	)

	// Print transform stats
	fmt.Printf("%s: %d (%.2f%%), %s: %d, %s: %d\n",
		"Transforms",
		m.GetMetricCount(types.MetricTotalTransformedCount),
		m.GetMetricPercentage(types.MetricTransformPercentage),
		m.GetMetricDisplayName(types.MetricTransformsPerSecond),
		m.GetMetricCount(types.MetricTransformsPerSecond),
		m.GetMetricDisplayName(types.MetricPeakTransformedPerSecond),
		m.GetMetricCount(types.MetricPeakTransformedPerSecond),
	)

	// Print error stats
	fmt.Printf("%s: %d (%.2f%%), %s: %d, %s: %d\n",
		"Transform Errors",
		m.GetMetricCount(types.MetricTotalErrorCount), // or whatever your real error metric is
		m.GetMetricPercentage(types.MetricErrorPercentage),
		m.GetMetricDisplayName(types.MetricTransformationErrorsPerSecond),
		m.GetMetricCount(types.MetricTransformationErrorsPerSecond),
		m.GetMetricDisplayName(types.MetricPeakTransformationErrorsPerSecond),
		m.GetMetricCount(types.MetricPeakTransformationErrorsPerSecond),
	)

	// Print any other non-zero or monitored metrics
	for _, name := range m.GetMetricNames() {
		switch name {
		case types.MetricTransformsPerSecond,
			types.MetricCurrentGoRoutinesActive,
			types.MetricTransformationErrorsPerSecond,
			types.MetricTotalProcessedCount,
			types.MetricTotalSubmittedCount,
			types.MetricTransformPercentage,
			types.MetricTransformationErrorPercentage,
			types.MetricTotalTransformedCount,
			types.MetricPeakProcessedPerSecond,
			types.MetricPeakTransformedPerSecond,
			types.MetricPeakTransformationErrorsPerSecond:
			continue
		case types.MetricSurgeProtectorBackupWireSubmissionCount:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d (%.2f%%)\n",
					m.metrics[name].DisplayAs,
					count,
					m.GetMetricPercentage(types.MetricSurgeProtectorBackupWireSubmissionPercentage),
				)
			}
		default:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d\n", m.metrics[name].DisplayAs, count)
			}
		}
	}

	for _, metricInfo := range m.monitoredMetrics {
		fmt.Printf("%s: %d\n",
			metricInfo.DisplayAs,
			m.GetMetricCount(metricInfo.Name))
	}
}

func (m *Meter[T]) printFinalProgress() {
	m.endTime = time.Now()                              // Capture the end time when final progress is printed
	elapsedTime := m.endTime.Sub(m.startTime).Seconds() // Get elapsed time in seconds

	totalSubmitted := m.GetMetricCount(types.MetricTotalSubmittedCount)
	totalErrors := m.GetMetricCount(types.MetricTransformationErrorPercentage)
	totalTransformed := m.GetMetricCount(types.MetricTotalTransformedCount)
	totalPending := totalSubmitted - (totalErrors + totalTransformed) // Example additional metric

	var pendingPercentage float64
	if totalSubmitted > 0 {
		pendingPercentage = float64(totalPending) / float64(totalSubmitted) * 100
	}

	m.resetRunningMetrics()
	// Clear previous output and move cursor up
	fmt.Printf("\033[2J\033[H") // Clear the screen and move cursor to the top
	fmt.Printf("Start Time: %v, Elapsed Time: %s\n", m.startTime.Format("01-02-2006 15:04:05"), time.Duration(elapsedTime)*time.Second)
	fmt.Printf("%s: %d, %s: %d, %s: %.2f%%, %s: %.2f%%, %s: %.2f%%, %s: %.2f%%\n",
		"Last Recorded Active Go Routines",
		m.GetMetricCount(types.MetricCurrentGoRoutinesActive),
		"Peak",
		m.GetMetricCount(types.MetricPeakGoRoutinesActive),
		"Last Recorded CPU Percentage",
		m.GetMetricPercentage(types.MetricCurrentCpuPercentage),
		"Peak",
		m.GetMetricPeakPercentage(types.MetricCurrentCpuPercentage),
		"Last Recorded RAM Percentage",
		m.GetMetricPercentage(types.MetricCurrentRamPercentage),
		"Peak",
		m.GetMetricPeakPercentage(types.MetricCurrentRamPercentage),
	) // Truncate to remove microsecond precision if not needed
	if m.totalItems != 0 {
		fmt.Printf("Total Expected Elements: %d, Progress: %d%%, Pending: %d (%.2f%%)\n", m.totalItems, int(float64(m.GetMetricCount(types.MetricTotalProcessedCount))/float64(m.totalItems)*100), totalPending, pendingPercentage)
	}
	fmt.Printf("%s: %d, Last Recorded %s: %d, Peak %s: %d\n", m.GetMetricDisplayName(types.MetricTotalSubmittedCount), m.GetMetricCount(types.MetricTotalSubmittedCount), m.GetMetricDisplayName(types.MetricProcessedPerSecond), (m.GetMetricCount(types.MetricTransformsPerSecond) + m.GetMetricCount(types.MetricTransformationErrorsPerSecond)), m.GetMetricDisplayName(types.MetricPeakProcessedPerSecond), m.GetMetricCount(types.MetricPeakProcessedPerSecond))

	fmt.Printf("%s: %d (%.2f%%), Last Recorded %s: %d, Peak %s: %d\n", "Transform Success", m.GetMetricCount(types.MetricTotalTransformedCount), m.GetMetricPercentage(types.MetricTransformPercentage), m.GetMetricDisplayName(types.MetricTransformsPerSecond), m.GetMetricCount(types.MetricTransformsPerSecond), m.GetMetricDisplayName(types.MetricPeakTransformedPerSecond), m.GetMetricCount(types.MetricPeakTransformedPerSecond))
	fmt.Printf("%s: %d (%.2f%%), Last Recorded %s: %d, Peak %s: %d\n", "Transform Errors", m.GetMetricCount(types.MetricTransformationErrorPercentage), m.GetMetricPercentage(types.MetricErrorPercentage), m.GetMetricDisplayName(types.MetricTransformationErrorsPerSecond), m.GetMetricCount(types.MetricTransformationErrorsPerSecond), m.GetMetricDisplayName(types.MetricPeakTransformationErrorsPerSecond), m.GetMetricCount(types.MetricPeakTransformationErrorsPerSecond))
	for _, name := range m.GetMetricNames() {
		// Skip the metrics already displayed above
		switch name {
		case types.MetricTotalProcessedCount, types.MetricTotalSubmittedCount, types.MetricTransformPercentage, types.MetricTransformationErrorPercentage:
			continue
		case types.MetricSurgeProtectorBackupWireSubmissionCount:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d (%.2f%%)\n", m.metrics[name].DisplayAs, count, m.GetMetricPercentage(types.MetricSurgeProtectorBackupWireSubmissionPercentage))
			}
		default:
			count := m.GetMetricCount(name)
			if count > 0 {
				fmt.Printf("%s: %d\n", m.metrics[name].DisplayAs, count)
			}
		}
	}

	for _, metricInfo := range m.monitoredMetrics {
		fmt.Printf("%s: %d\n", metricInfo.DisplayAs, m.GetMetricCount(metricInfo.Name))
	}
	for _, metricInfo := range m.monitoredMetrics {
		fmt.Printf("%s: %d\n", metricInfo.DisplayAs, m.GetMetricCount(metricInfo.Name))
	}
	fmt.Println() // Add a blank line at the end for spacing
}

func (m *Meter[T]) monitorIdleTime(ctx context.Context) {
	defer m.idleTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			return // Exit if context is done
		case <-m.dataChan:
			// Reset the timer whenever there's activity
			if !m.idleTimer.Stop() {
				<-m.idleTimer.C
			}
			m.idleTimer.Reset(m.idleTimeout)
		}
	}
}
