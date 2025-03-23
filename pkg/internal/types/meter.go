package types

import (
	"context"
	"time"
)

const (
	MetricCurrentCpuPercentage                         = "current_cpu_percentage"
	MetricCurrentRamPercentage                         = "current_ram_percentage"
	MetricPeakGoRoutinesActive                         = "peak_go_routines_active"
	MetricCurrentGoRoutinesActive                      = "current_go_routines_active"
	MetricPeakProcessedPerSecond                       = "max_processed_per_second"
	MetricPeakTransformedPerSecond                     = "max_transformed_per_second"
	MetricPeakTransformationErrorsPerSecond            = "max_transform_errors_per_second"
	MetricProcessedPerSecond                           = "processed_per_second"
	MetricTransformsPerSecond                          = "transforms_per_second"
	MetricTransformationErrorsPerSecond                = "errors_per_second"
	MetricTotalComponentsUsedCount                     = "total_components_used_count"
	MetricTransformPercentage                          = "transform_percentage"
	MetricErrorPercentage                              = "error_percentage"
	MetricTotalErrorCount                              = "total_error_count"
	MetricProgressPercentage                           = "progress_percentage"
	MetricTotalPendingCount                            = "element_pending_total_count"
	MetricTotalSubmittedCount                          = "element_submitted_total_count"
	MetricTotalProcessedCount                          = "element_processed_total_count"
	MetricTotalTransformedCount                        = "element_transformed_total_count"
	MetricComponentRunningCount                        = "component_running_count"
	MetricMeterConnectedComponentCount                 = "meter_connected_component_count"
	MetricLoggerConnectedComponentCount                = "logger_connected_component_count"
	MetricComponentRestartCount                        = "component_restart_count"
	MetricComponentLifecycleErrorCount                 = "component_lifecycle_error_count"
	MetricWireElementTransformCount                    = "element_transform_count"
	MetricWireElementTransformErrorCount               = "element_transform_error_count"
	MetricWireRunningCount                             = "wires_running_count"
	MetricWireElementSubmitCount                       = "element_submit_count"
	MetricWireElementSubmitErrorCount                  = "element_submit_error_count"
	MetricWireElementSubmitCancelCount                 = "element_submit_cancel_count"
	MetricWireElementCancelCount                       = "element_cancel_count"
	MetricWireElementErrorCount                        = "element_error_count"
	MetricWireElementRecoverAttemptCount               = "element_retry_count"
	MetricWireElementRecoverSuccessCount               = "element_recover_success_count"
	MetricWireElementRecoverFailureCount               = "element_recover_failure_count"
	MetricCircuitBreakerResetTimer                     = "circuit_breaker_reset_timer"
	MetricCircuitBreakerLastErrorRecordTime            = "circuit_breaker_last_error_record_time"
	MetricCircuitBreakerLastTripTime                   = "circuit_breaker_last_trip_time"
	MetricCircuitBreakerLastResetTime                  = "circuit_breaker_last_reset_time"
	MetricCircuitBreakerNextResetTime                  = "circuit_breaker_next_reset_time"
	MetricCircuitBreakerCurrentTripCount               = "circuit_breaker_current_trip_count"
	MetricCircuitBreakerTripCount                      = "circuit_breaker_trip_count"
	MetricCircuitBreakerResetCount                     = "circuit_breaker_reset_count"
	MetricCircuitBreakerNeutralWireSubmissionCount     = "circuit_breaker_diverted_element_count"
	MetricCircuitBreakerRecordedErrorCount             = "circuit_breaker_recorded_error_count"
	MetricCircuitBreakerDroppedElementCount            = "circuit_breaker_dropped_element_count"
	MetricCircuitBreakerNeutralWireConnectedCount      = "circuit_breaker_neutral_wire_running_count"
	MetricCircuitBreakerNeutralWireFailureCount        = "circuit_breaker_neutral_wire_failure_count"
	MetricCircuitBreakerCount                          = "circuit_breaker_count"
	MetricTransformationErrorPercentage                = "error_count"
	MetricHTTPRequestMadeCount                         = "http_request_made_count"
	MetricHTTPRequestReceivedCount                     = "http_request_received_count"
	MetricHTTPRequestCompletedCount                    = "http_request_completed_count"
	MetricHTTPResponseCount                            = "http_response_count"
	MetricHTTPClientErrorCount                         = "http_client_error_count"
	MetricHTTPClientRetryCount                         = "http_client_retry_count"
	MetricHTTPOAuth2ClientCount                        = "http_oauth2_client_count"
	MetricHTTPResponseErrorCount                       = "http_response_error_count"
	MetricHTTPClientWithTlsPinningCount                = "http_client_with_tls_pinning_count"
	MetricHTTPClientFetchCount                         = "http_client_fetch_count"
	MetricHTTPClientFetchErrorCount                    = "http_client_fetch_error_count"
	MetricHTTPClientFetchAndSubmitErrorCount           = "http_client_fetch_and_submit_error_count"
	MetricHTTPClientWithActiveOAuthTokenCount          = "http_client_with_active_oauth_token_count"
	MetricHTTPClientOAuthTokenRequestCount             = "http_client_oauth_token_request_count"
	MetricHTTPClientFailCount                          = "http_client_oauth_token_request_count"
	MetricHTTPClientOAuthTokenErrorCount               = "http_client_oauth_token_error_count"
	MetricHTTPClientJsonUnmarshalErrorCount            = "http_client_json_unmarshal_error_count"
	MetricHTTPClientSuccessfulFetchCount               = "http_client_fetch_successful_count"
	MetricHTTPClientRetryExhaustedCount                = "http_client_fetch_retry_exhausted_count"
	MetricSurgeProtectorAttachedCount                  = "surge_attached_count"
	MetricSurgeProtectorCurrentTripCount               = "surge_current_trip_count"
	MetricSurgeProtectorBackupWireSubmissionCount      = "surge_backup_wire_submission_count"
	MetricSurgeProtectorBackupWireSubmissionPercentage = "surge_backup_wire_submission_percentage"
	MetricSurgeProtectorDropCount                      = "surge_drop_count"
	MetricSurgeCount                                   = "surge_count"
	MetricSurgeTripCount                               = "surge_trip_count"
	MetricSurgeResetCount                              = "surge_reset_count"
	MetricSurgeBackupFailureCount                      = "surge_backup_failure_count"
	MetricSurgeRateLimitExceedCount                    = "surge_rate_limit_exceed_count"
	MetricSurgeBlackoutTripCount                       = "surge_blackout_trip_count"
	MetricSurgeBlackoutResetCount                      = "surge_blackout_reset_count"
	MetricResisterElementRequeued                      = "resister_element_requeued"
	MetricResisterConnectedCount                       = "resister_connected_count"
	MetricResisterClearedCount                         = "resister_cleared_count"
	MetricResisterElementQueuedCount                   = "resister_element_queued_count"
	MetricResisterElementCurrentlyQueuedCount          = "resister_elements_currently_queued_count"
	MetricResisterElementDequeued                      = "resister_element_dequeued"
	MetricReceivingRelayListeningCount                 = "receiving_relay_listening_count"
	MetricReceivingRelayRunningCount                   = "receiving_relay_running_count"
	MetricReceivingRelayReceivedCount                  = "receiving_relay_received_count"
	MetricReceivingRelayUnwrappedPayloadCount          = "receiving_relay_unwrapped_payload_count"
	MetricReceivingRelayRelayedCount                   = "receiving_relay_relayed_count"
	MetricReceivingRelayUnwrappedPayloadErrorCount     = "receiving_relay_unwrapped_payload_error_count"
	MetricReceivingRelayStreamReceiveCount             = "receiving_relay_stream_receive_count"
	MetricReceivingRelayTargetCount                    = "receiving_relays_as_targets_count"
	MetricReceivingRelayErrorCount                     = "receiving_relay_error_count"
	MetricReceivingRelayNoTLSCount                     = "receiving_relay_no_tls_count"
	MetricProcessDuration                              = "process_duration"
	MetricGeneratorRunningCount                        = "generator_running_count"
	MetricGeneratorSubmitCount                         = "generator_submit_count"
	MetricGeneratorSubmitErrorCount                    = "generator_submit_error_count"
	MetricForwardRelaySubmitErrorCount                 = "forward_relay_submit_error_count"
	MetricForwardRelaySubmitCount                      = "forward_relay_submit_count"
	MetricForwardRelayRelayedCount                     = "forward_relay_relayed_count"
	MetricForwardRelayWithPerformanceOptionCount       = "forward_relay_with_performance_option_count"
	MetricForwardRelayWithTlsConfigCount               = "forward_relay_with_tls_config_count"
	MetricForwardRelayRunningCount                     = "forward_relay_running_count"
	MetricForwardRelayErrorCount                       = "forward_relay_error_count"
	MetricForwardRelayPayloadCompressionCount          = "forward_relay_payload_compression_count"
	MetricForwardRelayPayloadCompressionErrorCount     = "forward_relay_payload_compression_error_count"
	MetricForwardRelayWrappedPayloadCount              = "forward_relay_wrapped_payload_count"
	MetricForwardRelayInputCount                       = "forward_relay_input_count"
	MetricConduitRunningCount                          = "conduit_running_count"
)

// ThresholdType defines whether a threshold is an absolute number or a percentage.
type ThresholdType int

const (
	Absolute   ThresholdType = iota // Absolute value threshold
	Percentage                      // Percentage value threshold
)

type MetricInfo struct {
	Count          *uint64
	Total          uint64
	Percentage     float64
	Threshold      float64
	ThresholdType  ThresholdType // Add this field to indicate the type of threshold
	PeakPercentage float64
	Alerted        bool
	DisplayAs      string
	Name           string
	Timestamp      int64
	Monitored      bool // Indicates if the metric should be actively monitored
}

type Meter[T any] interface {
	GetMetricDisplayName(metricName string) string
	GetMetricPercentage(metricName string) float64
	SetMetricTimestamp(metricName string, time int64)
	SetMetricPercentage(name string, percentage float64)
	GetMetricNames() []string
	ReportData()
	SetIdleTimeout(to time.Duration)
	SetContextCancelHook(hook func())
	SetDynamicMetric(metricName string, total uint64, initialCount uint64, threshold float64)
	GetDynamicMetricInfo(metricName string) (*MetricInfo, bool)
	SetMetricPeak(metricName string, count uint64)
	SetDynamicMetricTotal(metricName string, total uint64)
	GetTicker() *time.Ticker
	SetTicker(ticker *time.Ticker)
	GetOriginalContext() context.Context
	GetOriginalContextCancel() context.CancelFunc
	Monitor()
	AddTotalItems(additionalTotal uint64)
	CheckMetrics() bool
	PauseProcessing()
	ResumeProcessing()
	GetMetricCount(metricName string) uint64
	SetMetricCount(metricName string, count uint64)
	GetMetricTotal(metricName string) uint64
	SetMetricTotal(metricName string, total uint64)
	IncrementCount(metricName string)
	DecrementCount(metricName string)
	IsTimerRunning(metricName string) bool
	GetTimerStartTime(metricName string) (time.Time, bool)
	SetTimerStartTime(metricName string, startTime time.Time)
	SetTotal(metricName string, total uint64)
	StartTimer(metricName string)
	StopTimer(metricName string) time.Duration
	GetComponentMetadata() ComponentMetadata
	ConnectLogger(...Logger)
	NotifyLoggers(level LogLevel, format string, args ...interface{})
	SetComponentMetadata(name string, id string)
	ResetMetrics()
}
