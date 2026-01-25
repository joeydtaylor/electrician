package sensor_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestSensor_WireCallbacks(t *testing.T) {
	s := sensor.NewSensor[int]()

	var startCount int32
	var submitCount int32
	var processedCount int32
	var errorCount int32
	var cancelCount int32
	var completeCount int32
	var stopCount int32
	var restartCount int32

	s.RegisterOnStart(func(types.ComponentMetadata) { atomic.AddInt32(&startCount, 1) })
	s.RegisterOnSubmit(func(types.ComponentMetadata, int) { atomic.AddInt32(&submitCount, 1) })
	s.RegisterOnElementProcessed(func(types.ComponentMetadata, int) { atomic.AddInt32(&processedCount, 1) })
	s.RegisterOnError(func(types.ComponentMetadata, error, int) { atomic.AddInt32(&errorCount, 1) })
	s.RegisterOnCancel(func(types.ComponentMetadata, int) { atomic.AddInt32(&cancelCount, 1) })
	s.RegisterOnComplete(func(types.ComponentMetadata) { atomic.AddInt32(&completeCount, 1) })
	s.RegisterOnTerminate(func(types.ComponentMetadata) { atomic.AddInt32(&stopCount, 1) })
	s.RegisterOnRestart(func(types.ComponentMetadata) { atomic.AddInt32(&restartCount, 1) })

	meta := types.ComponentMetadata{Type: "WIRE"}
	s.InvokeOnStart(meta)
	s.InvokeOnSubmit(meta, 1)
	s.InvokeOnElementProcessed(meta, 1)
	s.InvokeOnError(meta, errSentinel{}, 1)
	s.InvokeOnCancel(meta, 1)
	s.InvokeOnComplete(meta)
	s.InvokeOnStop(meta)
	s.InvokeOnRestart(meta)

	if startCount != 1 {
		t.Fatalf("expected 1 start callback, got %d", startCount)
	}
	if submitCount != 1 {
		t.Fatalf("expected 1 submit callback, got %d", submitCount)
	}
	if processedCount != 1 {
		t.Fatalf("expected 1 processed callback, got %d", processedCount)
	}
	if errorCount != 1 {
		t.Fatalf("expected 1 error callback, got %d", errorCount)
	}
	if cancelCount != 1 {
		t.Fatalf("expected 1 cancel callback, got %d", cancelCount)
	}
	if completeCount != 1 {
		t.Fatalf("expected 1 complete callback, got %d", completeCount)
	}
	if stopCount != 1 {
		t.Fatalf("expected 1 stop callback, got %d", stopCount)
	}
	if restartCount != 1 {
		t.Fatalf("expected 1 restart callback, got %d", restartCount)
	}
}

func TestSensor_CircuitBreakerCallbacks(t *testing.T) {
	s := sensor.NewSensor[int]()

	var tripCount int32
	var resetCount int32
	var allowCount int32
	var dropCount int32

	s.RegisterOnCircuitBreakerTrip(func(types.ComponentMetadata, int64, int64) { atomic.AddInt32(&tripCount, 1) })
	s.RegisterOnCircuitBreakerReset(func(types.ComponentMetadata, int64) { atomic.AddInt32(&resetCount, 1) })
	s.RegisterOnCircuitBreakerAllow(func(types.ComponentMetadata) { atomic.AddInt32(&allowCount, 1) })
	s.RegisterOnCircuitBreakerDrop(func(types.ComponentMetadata, int) { atomic.AddInt32(&dropCount, 1) })

	meta := types.ComponentMetadata{Type: "CIRCUIT_BREAKER"}
	s.InvokeOnCircuitBreakerTrip(meta, 10, 20)
	s.InvokeOnCircuitBreakerReset(meta, 30)
	s.InvokeOnCircuitBreakerAllow(meta)
	s.InvokeOnCircuitBreakerDrop(meta, 1)

	if tripCount != 1 {
		t.Fatalf("expected 1 trip callback, got %d", tripCount)
	}
	if resetCount != 1 {
		t.Fatalf("expected 1 reset callback, got %d", resetCount)
	}
	if allowCount != 1 {
		t.Fatalf("expected 1 allow callback, got %d", allowCount)
	}
	if dropCount != 1 {
		t.Fatalf("expected 1 drop callback, got %d", dropCount)
	}
}

func TestSensor_NotifyLoggers(t *testing.T) {
	s := sensor.NewSensor[int]()

	logger := &countingLogger{level: types.InfoLevel}
	s.ConnectLogger(logger)

	s.NotifyLoggers(types.DebugLevel, "debug")
	s.NotifyLoggers(types.InfoLevel, "info")

	if got := atomic.LoadInt32(&logger.debug); got != 0 {
		t.Fatalf("expected 0 debug logs, got %d", got)
	}
	if got := atomic.LoadInt32(&logger.info); got != 1 {
		t.Fatalf("expected 1 info log, got %d", got)
	}
}

func TestSensor_MeterDecorators(t *testing.T) {
	s := sensor.NewSensor[int]()
	meter := newStubMeter[int]()
	s.ConnectMeter(meter)

	meta := types.ComponentMetadata{Type: "WIRE"}
	s.InvokeOnStart(meta)
	if got := meter.GetMetricCount(types.MetricComponentRunningCount); got != 1 {
		t.Fatalf("expected running count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricWireRunningCount); got != 1 {
		t.Fatalf("expected wire running count 1, got %d", got)
	}

	s.InvokeOnStop(meta)
	if got := meter.GetMetricCount(types.MetricComponentRunningCount); got != 0 {
		t.Fatalf("expected running count 0, got %d", got)
	}

	s.InvokeOnCircuitBreakerTrip(types.ComponentMetadata{Type: "CIRCUIT_BREAKER"}, 100, 200)
	if meter.GetMetricCount(types.MetricCircuitBreakerTripCount) == 0 {
		t.Fatalf("expected circuit breaker trip count to increment")
	}
	if meter.GetMetricTimestamp(types.MetricCircuitBreakerLastTripTime) != 100 {
		t.Fatalf("expected last trip timestamp to be recorded")
	}
}

func TestSensor_MetadataAndOptions(t *testing.T) {
	var startCount int32
	var httpStartCount int32
	var s3StartCount int32
	var kafkaStartCount int32
	var resisterQueuedCount int32
	var allowCount int32

	logger := &countingLogger{level: types.InfoLevel}
	meter := newStubMeter[int]()

	s := sensor.NewSensor[int](
		sensor.WithComponentMetadata[int]("sensor-name", "sensor-id"),
		sensor.WithOnStartFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&startCount, 1)
		}),
		sensor.WithOnHTTPClientRequestStartFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&httpStartCount, 1)
		}),
		sensor.WithOnS3WriterStartFunc[int](func(types.ComponentMetadata, string, string, string) {
			atomic.AddInt32(&s3StartCount, 1)
		}),
		sensor.WithOnKafkaWriterStartFunc[int](func(types.ComponentMetadata, string, string) {
			atomic.AddInt32(&kafkaStartCount, 1)
		}),
		sensor.WithResisterQueuedFunc[int](func(types.ComponentMetadata, int) {
			atomic.AddInt32(&resisterQueuedCount, 1)
		}),
		sensor.WithCircuitBreakerAllowFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&allowCount, 1)
		}),
		sensor.WithLogger[int](logger),
		sensor.WithMeter[int](meter),
	)

	meta := s.GetComponentMetadata()
	if meta.Name != "sensor-name" || meta.ID != "sensor-id" || meta.Type != "SENSOR" {
		t.Fatalf("unexpected metadata: %+v", meta)
	}

	s.InvokeOnStart(meta)
	s.InvokeOnHTTPClientRequestStart(meta)
	s.InvokeOnS3WriterStart(meta, "bucket", "prefix", "json")
	s.InvokeOnKafkaWriterStart(meta, "topic", "json")
	s.InvokeOnResisterQueued(meta, 1)
	s.InvokeOnCircuitBreakerAllow(meta)

	if startCount != 1 {
		t.Fatalf("expected start callback to be called once, got %d", startCount)
	}
	if httpStartCount != 1 {
		t.Fatalf("expected http start callback to be called once, got %d", httpStartCount)
	}
	if s3StartCount != 1 {
		t.Fatalf("expected s3 start callback to be called once, got %d", s3StartCount)
	}
	if kafkaStartCount != 1 {
		t.Fatalf("expected kafka start callback to be called once, got %d", kafkaStartCount)
	}
	if resisterQueuedCount != 1 {
		t.Fatalf("expected resister queued callback to be called once, got %d", resisterQueuedCount)
	}
	if allowCount != 1 {
		t.Fatalf("expected circuit breaker allow callback to be called once, got %d", allowCount)
	}

	s.NotifyLoggers(types.InfoLevel, "hello")
	if got := atomic.LoadInt32(&logger.info); got != 1 {
		t.Fatalf("expected info log to be called once, got %d", got)
	}

	if got := meter.GetMetricCount(types.MetricComponentRunningCount); got != 1 {
		t.Fatalf("expected component running count to increment, got %d", got)
	}
}

func TestSensor_GetMetersSnapshot(t *testing.T) {
	s := sensor.NewSensor[int]()
	m1 := newStubMeter[int]()
	m2 := newStubMeter[int]()

	s.ConnectMeter(nil, m1, nil, m2)
	meters := s.GetMeters()
	if len(meters) != 2 {
		t.Fatalf("expected 2 meters, got %d", len(meters))
	}

	meters[0] = nil
	meters2 := s.GetMeters()
	if meters2[0] == nil {
		t.Fatalf("expected GetMeters to return a copy")
	}
}

func TestSensor_NotifyLoggers_LevelChecker(t *testing.T) {
	s := sensor.NewSensor[int]()
	logger := &levelCheckLogger{
		enabled: map[types.LogLevel]bool{
			types.WarnLevel: true,
		},
	}
	logger.level = types.DebugLevel

	s.ConnectLogger(logger)

	s.NotifyLoggers(types.InfoLevel, "info")
	s.NotifyLoggers(types.WarnLevel, "warn")

	if got := atomic.LoadInt32(&logger.info); got != 0 {
		t.Fatalf("expected info logs to be skipped, got %d", got)
	}
	if got := atomic.LoadInt32(&logger.warn); got != 1 {
		t.Fatalf("expected warn logs to be called once, got %d", got)
	}
}

func TestSensor_WireMetrics(t *testing.T) {
	meter := newStubMeter[int]()
	s := sensor.NewSensor[int](sensor.WithMeter[int](meter))
	meta := types.ComponentMetadata{Type: "WIRE"}

	s.InvokeOnStart(meta)
	s.InvokeOnRestart(meta)

	if got := meter.GetMetricCount(types.MetricComponentRestartCount); got != 1 {
		t.Fatalf("expected restart count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricComponentRunningCount); got != 0 {
		t.Fatalf("expected component running count 0 after restart, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricWireRunningCount); got != 0 {
		t.Fatalf("expected wire running count 0 after restart, got %d", got)
	}

	s.InvokeOnSubmit(meta, 1)
	s.InvokeOnElementProcessed(meta, 1)
	s.InvokeOnError(meta, errSentinel{}, 1)
	s.InvokeOnCancel(meta, 1)

	if got := meter.GetMetricCount(types.MetricTotalSubmittedCount); got != 1 {
		t.Fatalf("expected total submitted count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricTotalProcessedCount); got != 2 {
		t.Fatalf("expected total processed count 2, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricTotalTransformedCount); got != 1 {
		t.Fatalf("expected total transformed count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricTotalErrorCount); got != 1 {
		t.Fatalf("expected total error count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricTransformationErrorPercentage); got != 1 {
		t.Fatalf("expected transformation error percentage count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricWireElementSubmitCancelCount); got != 1 {
		t.Fatalf("expected submit cancel count 1, got %d", got)
	}
}

func TestSensor_InsulatorCallbacks(t *testing.T) {
	s := sensor.NewSensor[int]()
	var attempts int32
	var successes int32
	var failures int32

	s.RegisterOnInsulatorAttempt(func(types.ComponentMetadata, int, int, error, error, int, int, time.Duration) {
		atomic.AddInt32(&attempts, 1)
	})
	s.RegisterOnInsulatorSuccess(func(types.ComponentMetadata, int, int, error, error, int, int, time.Duration) {
		atomic.AddInt32(&successes, 1)
	})
	s.RegisterOnInsulatorFailure(func(types.ComponentMetadata, int, int, error, error, int, int, time.Duration) {
		atomic.AddInt32(&failures, 1)
	})

	meta := types.ComponentMetadata{Type: "WIRE"}
	s.InvokeOnInsulatorAttempt(meta, 1, 1, errSentinel{}, errSentinel{}, 1, 3, time.Millisecond)
	s.InvokeOnInsulatorSuccess(meta, 2, 1, nil, errSentinel{}, 2, 3, time.Millisecond)
	s.InvokeOnInsulatorFailure(meta, 1, 1, errSentinel{}, errSentinel{}, 3, 3, time.Millisecond)

	if attempts != 1 {
		t.Fatalf("expected 1 attempt callback, got %d", attempts)
	}
	if successes != 1 {
		t.Fatalf("expected 1 success callback, got %d", successes)
	}
	if failures != 1 {
		t.Fatalf("expected 1 failure callback, got %d", failures)
	}
}

func TestSensor_HTTPClientCallbacksAndMetrics(t *testing.T) {
	meter := newStubMeter[int]()
	s := sensor.NewSensor[int](sensor.WithMeter[int](meter))

	var startCount int32
	var respCount int32
	var errCount int32
	var completeCount int32

	s.RegisterOnHTTPClientRequestStart(func(types.ComponentMetadata) {
		atomic.AddInt32(&startCount, 1)
	})
	s.RegisterOnHTTPClientResponseReceived(func(types.ComponentMetadata) {
		atomic.AddInt32(&respCount, 1)
	})
	s.RegisterOnHTTPClientError(func(types.ComponentMetadata, error) {
		atomic.AddInt32(&errCount, 1)
	})
	s.RegisterOnHTTPClientRequestComplete(func(types.ComponentMetadata) {
		atomic.AddInt32(&completeCount, 1)
	})

	meta := types.ComponentMetadata{Type: "HTTP_CLIENT"}
	s.InvokeOnHTTPClientRequestStart(meta)
	s.InvokeOnHTTPClientResponseReceived(meta)
	s.InvokeOnHTTPClientError(meta, errSentinel{})
	s.InvokeOnHTTPClientRequestComplete(meta)

	if startCount != 1 || respCount != 1 || errCount != 1 || completeCount != 1 {
		t.Fatalf("unexpected http client callback counts: start=%d resp=%d err=%d complete=%d", startCount, respCount, errCount, completeCount)
	}
	if got := meter.GetMetricCount(types.MetricHTTPRequestMadeCount); got != 1 {
		t.Fatalf("expected request made count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricHTTPResponseCount); got != 1 {
		t.Fatalf("expected response count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricHTTPRequestCompletedCount); got != 1 {
		t.Fatalf("expected request complete count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricHTTPClientErrorCount); got != 1 {
		t.Fatalf("expected http client error count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricTotalErrorCount); got != 1 {
		t.Fatalf("expected total error count 1, got %d", got)
	}
}

func TestSensor_CircuitBreakerCallbacksAndMetrics(t *testing.T) {
	meter := newStubMeter[int]()
	s := sensor.NewSensor[int](sensor.WithMeter[int](meter))

	var tripCount int32
	var resetCount int32
	var allowCount int32
	var dropCount int32
	var recordCount int32
	var neutralCount int32

	s.RegisterOnCircuitBreakerTrip(func(types.ComponentMetadata, int64, int64) {
		atomic.AddInt32(&tripCount, 1)
	})
	s.RegisterOnCircuitBreakerReset(func(types.ComponentMetadata, int64) {
		atomic.AddInt32(&resetCount, 1)
	})
	s.RegisterOnCircuitBreakerAllow(func(types.ComponentMetadata) {
		atomic.AddInt32(&allowCount, 1)
	})
	s.RegisterOnCircuitBreakerDrop(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&dropCount, 1)
	})
	s.RegisterOnCircuitBreakerRecordError(func(types.ComponentMetadata, int64) {
		atomic.AddInt32(&recordCount, 1)
	})
	s.RegisterOnCircuitBreakerNeutralWireSubmission(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&neutralCount, 1)
	})

	meta := types.ComponentMetadata{Type: "CIRCUIT_BREAKER"}
	s.InvokeOnCircuitBreakerTrip(meta, 10, 20)
	s.InvokeOnCircuitBreakerReset(meta, 30)
	s.InvokeOnCircuitBreakerAllow(meta)
	s.InvokeOnCircuitBreakerDrop(meta, 1)
	s.InvokeOnCircuitBreakerRecordError(meta, 40)
	s.InvokeOnCircuitBreakerNeutralWireSubmission(meta, 2)

	if tripCount != 1 || resetCount != 1 || allowCount != 1 || dropCount != 1 || recordCount != 1 || neutralCount != 1 {
		t.Fatalf("unexpected circuit breaker callback counts")
	}
	if got := meter.GetMetricCount(types.MetricCircuitBreakerTripCount); got != 1 {
		t.Fatalf("expected trip count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricCircuitBreakerResetCount); got != 1 {
		t.Fatalf("expected reset count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricCircuitBreakerCurrentTripCount); got != 0 {
		t.Fatalf("expected current trip count 0 after reset, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricCircuitBreakerNeutralWireSubmissionCount); got != 1 {
		t.Fatalf("expected neutral wire submission count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricCircuitBreakerDroppedElementCount); got != 1 {
		t.Fatalf("expected drop count 1, got %d", got)
	}
	if got := meter.GetMetricTimestamp(types.MetricCircuitBreakerLastTripTime); got != 10 {
		t.Fatalf("expected last trip timestamp 10, got %d", got)
	}
	if got := meter.GetMetricTimestamp(types.MetricCircuitBreakerNextResetTime); got != 20 {
		t.Fatalf("expected next reset timestamp 20, got %d", got)
	}
	if got := meter.GetMetricTimestamp(types.MetricCircuitBreakerLastResetTime); got != 30 {
		t.Fatalf("expected last reset timestamp 30, got %d", got)
	}
}

func TestSensor_SurgeProtectorCallbacksAndMetrics(t *testing.T) {
	meter := newStubMeter[int]()
	s := sensor.NewSensor[int](sensor.WithMeter[int](meter))

	var tripCount int32
	var resetCount int32
	var backupFailCount int32
	var rateLimitCount int32
	var backupSubmitCount int32
	var submitCount int32
	var dropCount int32
	var releaseCount int32
	var connectCount int32
	var detachCount int32

	s.RegisterOnSurgeProtectorTrip(func(types.ComponentMetadata) {
		atomic.AddInt32(&tripCount, 1)
	})
	s.RegisterOnSurgeProtectorReset(func(types.ComponentMetadata) {
		atomic.AddInt32(&resetCount, 1)
	})
	s.RegisterOnSurgeProtectorBackupFailure(func(types.ComponentMetadata, error) {
		atomic.AddInt32(&backupFailCount, 1)
	})
	s.RegisterOnSurgeProtectorRateLimitExceeded(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&rateLimitCount, 1)
	})
	s.RegisterOnSurgeProtectorBackupWireSubmit(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&backupSubmitCount, 1)
	})
	s.RegisterOnSurgeProtectorSubmit(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&submitCount, 1)
	})
	s.RegisterOnSurgeProtectorDrop(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&dropCount, 1)
	})
	s.RegisterOnSurgeProtectorReleaseToken(func(types.ComponentMetadata) {
		atomic.AddInt32(&releaseCount, 1)
	})
	s.RegisterOnSurgeProtectorConnectResister(func(types.ComponentMetadata, types.ComponentMetadata) {
		atomic.AddInt32(&connectCount, 1)
	})
	s.RegisterOnSurgeProtectorDetachedBackups(func(types.ComponentMetadata, types.ComponentMetadata) {
		atomic.AddInt32(&detachCount, 1)
	})

	meta := types.ComponentMetadata{Type: "SURGE_PROTECTOR"}
	resisterMeta := types.ComponentMetadata{Type: "RESISTER"}
	backupMeta := types.ComponentMetadata{Type: "WIRE"}

	s.InvokeOnSurgeProtectorTrip(meta)
	s.InvokeOnSurgeProtectorReset(meta)
	s.InvokeOnSurgeProtectorBackupFailure(meta, errSentinel{})
	s.InvokeOnSurgeProtectorRateLimitExceeded(meta, 1)
	s.InvokeOnSurgeProtectorBackupWireSubmit(meta, 1)
	s.InvokeOnSurgeProtectorSubmit(meta, 1)
	s.InvokeOnSurgeProtectorDrop(meta, 1)
	s.InvokeOnSurgeProtectorReleaseToken(meta)
	s.InvokeOnSurgeProtectorConnectResister(meta, resisterMeta)
	s.InvokeOnSurgeProtectorDetachedBackups(meta, backupMeta)

	if tripCount != 1 || resetCount != 1 || backupFailCount != 1 || rateLimitCount != 1 || backupSubmitCount != 1 || submitCount != 1 || dropCount != 1 || releaseCount != 1 || connectCount != 1 || detachCount != 1 {
		t.Fatalf("unexpected surge protector callback counts")
	}
	if got := meter.GetMetricCount(types.MetricSurgeTripCount); got != 1 {
		t.Fatalf("expected trip count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricSurgeResetCount); got != 1 {
		t.Fatalf("expected reset count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricSurgeProtectorCurrentTripCount); got != 0 {
		t.Fatalf("expected current trip count 0 after reset, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricSurgeBackupFailureCount); got != 1 {
		t.Fatalf("expected backup failure count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricSurgeRateLimitExceedCount); got != 1 {
		t.Fatalf("expected rate limit exceed count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricSurgeProtectorBackupWireSubmissionCount); got != 1 {
		t.Fatalf("expected backup submission count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricSurgeProtectorDropCount); got != 1 {
		t.Fatalf("expected drop count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricSurgeProtectorAttachedCount); got != 1 {
		t.Fatalf("expected submit count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricResisterConnectedCount); got != 1 {
		t.Fatalf("expected resister connected count 1, got %d", got)
	}
}

func TestSensor_ResisterCallbacksAndMetrics(t *testing.T) {
	meter := newStubMeter[int]()
	s := sensor.NewSensor[int](sensor.WithMeter[int](meter))

	var dequeued int32
	var queued int32
	var requeued int32
	var empty int32

	s.RegisterOnResisterDequeued(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&dequeued, 1)
	})
	s.RegisterOnResisterQueued(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&queued, 1)
	})
	s.RegisterOnResisterRequeued(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&requeued, 1)
	})
	s.RegisterOnResisterEmpty(func(types.ComponentMetadata) {
		atomic.AddInt32(&empty, 1)
	})

	meta := types.ComponentMetadata{Type: "RESISTER"}
	s.InvokeOnResisterQueued(meta, 1)
	s.InvokeOnResisterDequeued(meta, 1)
	s.InvokeOnResisterRequeued(meta, 1)
	s.InvokeOnResisterEmpty(meta)

	if dequeued != 1 || queued != 1 || requeued != 1 || empty != 1 {
		t.Fatalf("unexpected resister callback counts")
	}
	if got := meter.GetMetricCount(types.MetricResisterElementQueuedCount); got != 1 {
		t.Fatalf("expected queued count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricResisterElementDequeued); got != 1 {
		t.Fatalf("expected dequeued count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricResisterElementRequeued); got != 1 {
		t.Fatalf("expected requeued count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricResisterClearedCount); got != 1 {
		t.Fatalf("expected cleared count 1, got %d", got)
	}
	if got := meter.GetMetricCount(types.MetricResisterElementCurrentlyQueuedCount); got != 0 {
		t.Fatalf("expected currently queued count 0 after dequeue, got %d", got)
	}
}

func TestSensor_S3Callbacks(t *testing.T) {
	s := sensor.NewSensor[int]()
	var writerStart int32
	var writerStop int32
	var keyRendered int32
	var putAttempt int32
	var putSuccess int32
	var putError int32
	var rollFlush int32
	var listStart int32
	var listPage int32
	var objectSeen int32
	var decode int32
	var spill int32
	var complete int32
	var billing int32

	s.RegisterOnS3WriterStart(func(types.ComponentMetadata, string, string, string) {
		atomic.AddInt32(&writerStart, 1)
	})
	s.RegisterOnS3WriterStop(func(types.ComponentMetadata) {
		atomic.AddInt32(&writerStop, 1)
	})
	s.RegisterOnS3KeyRendered(func(types.ComponentMetadata, string) {
		atomic.AddInt32(&keyRendered, 1)
	})
	s.RegisterOnS3PutAttempt(func(types.ComponentMetadata, string, string, int, string, string) {
		atomic.AddInt32(&putAttempt, 1)
	})
	s.RegisterOnS3PutSuccess(func(types.ComponentMetadata, string, string, int, time.Duration) {
		atomic.AddInt32(&putSuccess, 1)
	})
	s.RegisterOnS3PutError(func(types.ComponentMetadata, string, string, int, error) {
		atomic.AddInt32(&putError, 1)
	})
	s.RegisterOnS3ParquetRollFlush(func(types.ComponentMetadata, int, int, string) {
		atomic.AddInt32(&rollFlush, 1)
	})
	s.RegisterOnS3ReaderListStart(func(types.ComponentMetadata, string, string) {
		atomic.AddInt32(&listStart, 1)
	})
	s.RegisterOnS3ReaderListPage(func(types.ComponentMetadata, int, bool) {
		atomic.AddInt32(&listPage, 1)
	})
	s.RegisterOnS3ReaderObject(func(types.ComponentMetadata, string, int64) {
		atomic.AddInt32(&objectSeen, 1)
	})
	s.RegisterOnS3ReaderDecode(func(types.ComponentMetadata, string, int, string) {
		atomic.AddInt32(&decode, 1)
	})
	s.RegisterOnS3ReaderSpillToDisk(func(types.ComponentMetadata, int64, int64) {
		atomic.AddInt32(&spill, 1)
	})
	s.RegisterOnS3ReaderComplete(func(types.ComponentMetadata, int, int) {
		atomic.AddInt32(&complete, 1)
	})
	s.RegisterOnS3BillingSample(func(types.ComponentMetadata, string, int64, int64, string) {
		atomic.AddInt32(&billing, 1)
	})

	meta := types.ComponentMetadata{Type: "S3"}
	s.InvokeOnS3WriterStart(meta, "bucket", "prefix", "parquet")
	s.InvokeOnS3WriterStop(meta)
	s.InvokeOnS3KeyRendered(meta, "key")
	s.InvokeOnS3PutAttempt(meta, "bucket", "key", 100, "sse", "kms")
	s.InvokeOnS3PutSuccess(meta, "bucket", "key", 100, time.Millisecond)
	s.InvokeOnS3PutError(meta, "bucket", "key", 100, errSentinel{})
	s.InvokeOnS3ParquetRollFlush(meta, 10, 1000, "snappy")
	s.InvokeOnS3ReaderListStart(meta, "bucket", "prefix")
	s.InvokeOnS3ReaderListPage(meta, 2, false)
	s.InvokeOnS3ReaderObject(meta, "key", 42)
	s.InvokeOnS3ReaderDecode(meta, "key", 3, "parquet")
	s.InvokeOnS3ReaderSpillToDisk(meta, 100, 200)
	s.InvokeOnS3ReaderComplete(meta, 2, 3)
	s.InvokeOnS3BillingSample(meta, "GET", 1, 2048, "STANDARD")

	if writerStart != 1 || writerStop != 1 || keyRendered != 1 || putAttempt != 1 || putSuccess != 1 || putError != 1 || rollFlush != 1 || listStart != 1 || listPage != 1 || objectSeen != 1 || decode != 1 || spill != 1 || complete != 1 || billing != 1 {
		t.Fatalf("unexpected s3 callback counts")
	}
}

func TestSensor_KafkaCallbacks(t *testing.T) {
	s := sensor.NewSensor[int]()
	var writerStart int32
	var writerStop int32
	var produceAttempt int32
	var produceSuccess int32
	var produceError int32
	var batchFlush int32
	var keyRendered int32
	var headersRendered int32
	var consumerStart int32
	var consumerStop int32
	var partitionAssigned int32
	var partitionRevoked int32
	var message int32
	var decode int32
	var commitSuccess int32
	var commitError int32
	var dlqAttempt int32
	var dlqSuccess int32
	var dlqError int32
	var billing int32

	s.RegisterOnKafkaWriterStart(func(types.ComponentMetadata, string, string) {
		atomic.AddInt32(&writerStart, 1)
	})
	s.RegisterOnKafkaWriterStop(func(types.ComponentMetadata) {
		atomic.AddInt32(&writerStop, 1)
	})
	s.RegisterOnKafkaProduceAttempt(func(types.ComponentMetadata, string, int, int, int) {
		atomic.AddInt32(&produceAttempt, 1)
	})
	s.RegisterOnKafkaProduceSuccess(func(types.ComponentMetadata, string, int, int64, time.Duration) {
		atomic.AddInt32(&produceSuccess, 1)
	})
	s.RegisterOnKafkaProduceError(func(types.ComponentMetadata, string, int, error) {
		atomic.AddInt32(&produceError, 1)
	})
	s.RegisterOnKafkaBatchFlush(func(types.ComponentMetadata, string, int, int, string) {
		atomic.AddInt32(&batchFlush, 1)
	})
	s.RegisterOnKafkaKeyRendered(func(types.ComponentMetadata, []byte) {
		atomic.AddInt32(&keyRendered, 1)
	})
	s.RegisterOnKafkaHeadersRendered(func(types.ComponentMetadata, []struct{ Key, Value string }) {
		atomic.AddInt32(&headersRendered, 1)
	})
	s.RegisterOnKafkaConsumerStart(func(types.ComponentMetadata, string, []string, string) {
		atomic.AddInt32(&consumerStart, 1)
	})
	s.RegisterOnKafkaConsumerStop(func(types.ComponentMetadata) {
		atomic.AddInt32(&consumerStop, 1)
	})
	s.RegisterOnKafkaPartitionAssigned(func(types.ComponentMetadata, string, int, int64, int64) {
		atomic.AddInt32(&partitionAssigned, 1)
	})
	s.RegisterOnKafkaPartitionRevoked(func(types.ComponentMetadata, string, int) {
		atomic.AddInt32(&partitionRevoked, 1)
	})
	s.RegisterOnKafkaMessage(func(types.ComponentMetadata, string, int, int64, int, int) {
		atomic.AddInt32(&message, 1)
	})
	s.RegisterOnKafkaDecode(func(types.ComponentMetadata, string, int, string) {
		atomic.AddInt32(&decode, 1)
	})
	s.RegisterOnKafkaCommitSuccess(func(types.ComponentMetadata, string, map[string]int64) {
		atomic.AddInt32(&commitSuccess, 1)
	})
	s.RegisterOnKafkaCommitError(func(types.ComponentMetadata, string, error) {
		atomic.AddInt32(&commitError, 1)
	})
	s.RegisterOnKafkaDLQProduceAttempt(func(types.ComponentMetadata, string, int, int, int) {
		atomic.AddInt32(&dlqAttempt, 1)
	})
	s.RegisterOnKafkaDLQProduceSuccess(func(types.ComponentMetadata, string, int, int64, time.Duration) {
		atomic.AddInt32(&dlqSuccess, 1)
	})
	s.RegisterOnKafkaDLQProduceError(func(types.ComponentMetadata, string, int, error) {
		atomic.AddInt32(&dlqError, 1)
	})
	s.RegisterOnKafkaBillingSample(func(types.ComponentMetadata, string, int64, int64) {
		atomic.AddInt32(&billing, 1)
	})

	meta := types.ComponentMetadata{Type: "KAFKA"}
	s.InvokeOnKafkaWriterStart(meta, "topic", "json")
	s.InvokeOnKafkaWriterStop(meta)
	s.InvokeOnKafkaProduceAttempt(meta, "topic", 0, 2, 4)
	s.InvokeOnKafkaProduceSuccess(meta, "topic", 0, 12, time.Millisecond)
	s.InvokeOnKafkaProduceError(meta, "topic", 0, errSentinel{})
	s.InvokeOnKafkaBatchFlush(meta, "topic", 2, 42, "snappy")
	s.InvokeOnKafkaKeyRendered(meta, []byte("key"))
	s.InvokeOnKafkaHeadersRendered(meta, []struct{ Key, Value string }{{Key: "k", Value: "v"}})
	s.InvokeOnKafkaConsumerStart(meta, "group", []string{"topic"}, "latest")
	s.InvokeOnKafkaConsumerStop(meta)
	s.InvokeOnKafkaPartitionAssigned(meta, "topic", 1, 10, 20)
	s.InvokeOnKafkaPartitionRevoked(meta, "topic", 1)
	s.InvokeOnKafkaMessage(meta, "topic", 1, 20, 1, 2)
	s.InvokeOnKafkaDecode(meta, "topic", 3, "ndjson")
	s.InvokeOnKafkaCommitSuccess(meta, "group", map[string]int64{"topic": 12})
	s.InvokeOnKafkaCommitError(meta, "group", errSentinel{})
	s.InvokeOnKafkaDLQProduceAttempt(meta, "dlq", 1, 1, 2)
	s.InvokeOnKafkaDLQProduceSuccess(meta, "dlq", 1, 2, time.Millisecond)
	s.InvokeOnKafkaDLQProduceError(meta, "dlq", 1, errSentinel{})
	s.InvokeOnKafkaBillingSample(meta, "produce", 1, 2048)

	if writerStart != 1 || writerStop != 1 || produceAttempt != 1 || produceSuccess != 1 || produceError != 1 || batchFlush != 1 || keyRendered != 1 || headersRendered != 1 || consumerStart != 1 || consumerStop != 1 || partitionAssigned != 1 || partitionRevoked != 1 || message != 1 || decode != 1 || commitSuccess != 1 || commitError != 1 || dlqAttempt != 1 || dlqSuccess != 1 || dlqError != 1 || billing != 1 {
		t.Fatalf("unexpected kafka callback counts")
	}
}

type errSentinel struct{}

func (errSentinel) Error() string { return "sentinel" }

type levelCheckLogger struct {
	countingLogger
	enabled map[types.LogLevel]bool
}

func (l *levelCheckLogger) IsLevelEnabled(level types.LogLevel) bool {
	return l.enabled[level]
}

type countingLogger struct {
	level  types.LogLevel
	debug  int32
	info   int32
	warn   int32
	err    int32
	dpanic int32
	panic  int32
	fatal  int32
}

func (c *countingLogger) GetLevel() types.LogLevel { return c.level }

func (c *countingLogger) SetLevel(level types.LogLevel) { c.level = level }

func (c *countingLogger) Debug(string, ...interface{})  { atomic.AddInt32(&c.debug, 1) }
func (c *countingLogger) Info(string, ...interface{})   { atomic.AddInt32(&c.info, 1) }
func (c *countingLogger) Warn(string, ...interface{})   { atomic.AddInt32(&c.warn, 1) }
func (c *countingLogger) Error(string, ...interface{})  { atomic.AddInt32(&c.err, 1) }
func (c *countingLogger) DPanic(string, ...interface{}) { atomic.AddInt32(&c.dpanic, 1) }
func (c *countingLogger) Panic(string, ...interface{})  { atomic.AddInt32(&c.panic, 1) }
func (c *countingLogger) Fatal(string, ...interface{})  { atomic.AddInt32(&c.fatal, 1) }

func (c *countingLogger) Flush() error { return nil }

func (c *countingLogger) AddSink(string, types.SinkConfig) error { return nil }

func (c *countingLogger) RemoveSink(string) error { return nil }

func (c *countingLogger) ListSinks() ([]string, error) { return nil, nil }

type stubMeter[T any] struct {
	counts     map[string]uint64
	timestamps map[string]int64
}

func newStubMeter[T any]() *stubMeter[T] {
	return &stubMeter[T]{
		counts:     make(map[string]uint64),
		timestamps: make(map[string]int64),
	}
}

func (s *stubMeter[T]) GetMetricDisplayName(metricName string) string { return metricName }

func (s *stubMeter[T]) GetMetricPercentage(metricName string) float64 { return 0 }

func (s *stubMeter[T]) SetMetricTimestamp(metricName string, time int64) {
	s.timestamps[metricName] = time
}

func (s *stubMeter[T]) GetMetricTimestamp(metricName string) int64 { return s.timestamps[metricName] }

func (s *stubMeter[T]) SetMetricPercentage(name string, percentage float64) {}

func (s *stubMeter[T]) GetMetricNames() []string { return nil }

func (s *stubMeter[T]) ReportData() {}

func (s *stubMeter[T]) SetIdleTimeout(to time.Duration) {}

func (s *stubMeter[T]) SetContextCancelHook(hook func()) {}

func (s *stubMeter[T]) SetDynamicMetric(metricName string, total uint64, initialCount uint64, threshold float64) {
}

func (s *stubMeter[T]) GetDynamicMetricInfo(metricName string) (*types.MetricInfo, bool) {
	return nil, false
}

func (s *stubMeter[T]) SetMetricPeak(metricName string, count uint64) {}

func (s *stubMeter[T]) SetDynamicMetricTotal(metricName string, total uint64) {}

func (s *stubMeter[T]) GetTicker() *time.Ticker { return nil }

func (s *stubMeter[T]) SetTicker(ticker *time.Ticker) {}

func (s *stubMeter[T]) GetOriginalContext() context.Context { return nil }

func (s *stubMeter[T]) GetOriginalContextCancel() context.CancelFunc { return nil }

func (s *stubMeter[T]) Monitor() {}

func (s *stubMeter[T]) AddTotalItems(additionalTotal uint64) {}

func (s *stubMeter[T]) CheckMetrics() bool { return true }

func (s *stubMeter[T]) PauseProcessing() {}

func (s *stubMeter[T]) ResumeProcessing() {}

func (s *stubMeter[T]) GetMetricCount(metricName string) uint64 { return s.counts[metricName] }

func (s *stubMeter[T]) SetMetricCount(metricName string, count uint64) { s.counts[metricName] = count }

func (s *stubMeter[T]) GetMetricTotal(metricName string) uint64 { return 0 }

func (s *stubMeter[T]) SetMetricTotal(metricName string, total uint64) {}

func (s *stubMeter[T]) IncrementCount(metricName string) { s.counts[metricName]++ }

func (s *stubMeter[T]) DecrementCount(metricName string) {
	if s.counts[metricName] > 0 {
		s.counts[metricName]--
	}
}

func (s *stubMeter[T]) IsTimerRunning(metricName string) bool { return false }

func (s *stubMeter[T]) GetTimerStartTime(metricName string) (time.Time, bool) {
	return time.Time{}, false
}

func (s *stubMeter[T]) SetTimerStartTime(metricName string, startTime time.Time) {}

func (s *stubMeter[T]) SetTotal(metricName string, total uint64) {}

func (s *stubMeter[T]) StartTimer(metricName string) {}

func (s *stubMeter[T]) StopTimer(metricName string) time.Duration { return 0 }

func (s *stubMeter[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{}
}

func (s *stubMeter[T]) ConnectLogger(...types.Logger) {}

func (s *stubMeter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {}

func (s *stubMeter[T]) SetComponentMetadata(name string, id string) {}

func (s *stubMeter[T]) ResetMetrics() {}

var _ types.Meter[int] = (*stubMeter[int])(nil)
