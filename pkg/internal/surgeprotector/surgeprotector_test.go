package surgeprotector_test

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/resister"
	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/surgeprotector"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func TestSurgeProtector_TryTakeRateLimit(t *testing.T) {
	ctx := context.Background()
	sp := surgeprotector.NewSurgeProtector[int](ctx)

	sp.SetRateLimit(1, 50*time.Millisecond, 0)

	if !sp.TryTake() {
		t.Fatalf("expected first token to be available")
	}
	if sp.TryTake() {
		t.Fatalf("expected no tokens left")
	}

	time.Sleep(60 * time.Millisecond)
	if !sp.TryTake() {
		t.Fatalf("expected token after refill")
	}
}

func TestSurgeProtector_SubmitNilElement(t *testing.T) {
	ctx := context.Background()
	sp := surgeprotector.NewSurgeProtector[int](ctx)

	if err := sp.Submit(ctx, nil); err != nil {
		t.Fatalf("expected nil error for nil element, got %v", err)
	}
}

func TestSurgeProtector_SubmitToManagedComponents(t *testing.T) {
	ctx := context.Background()

	component := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}
	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectComponent(component)

	elem := types.NewElementFast(10)
	if err := sp.Submit(ctx, elem); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	if got := component.SubmitCount(); got != 1 {
		t.Fatalf("expected 1 submission, got %d", got)
	}
	if got := component.Last(); got != 10 {
		t.Fatalf("expected last submission to be 10, got %d", got)
	}
}

func TestSurgeProtector_BackupStartsWhenStopped(t *testing.T) {
	ctx := context.Background()

	backup := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.AttachBackup(backup)

	if err := sp.Submit(ctx, types.NewElementFast(3)); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	if got := backup.StartCount(); got != 1 {
		t.Fatalf("expected backup start count 1, got %d", got)
	}
	if got := backup.SubmitCount(); got != 1 {
		t.Fatalf("expected backup submit count 1, got %d", got)
	}
}

func TestSurgeProtector_SubmitWithBackups(t *testing.T) {
	ctx := context.Background()

	backup := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}
	primary := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.AttachBackup(backup)
	sp.ConnectComponent(primary)

	elem := types.NewElementFast(7)
	if err := sp.Submit(ctx, elem); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	if got := backup.SubmitCount(); got != 1 {
		t.Fatalf("expected backup to receive 1 submission, got %d", got)
	}
	if got := primary.SubmitCount(); got != 0 {
		t.Fatalf("expected primary to receive 0 submissions, got %d", got)
	}
}

func TestSurgeProtector_ReleaseTokenRestoresCapacity(t *testing.T) {
	ctx := context.Background()

	var releases int32
	s := sensor.NewSensor[int](
		sensor.WithSurgeProtectorReleaseTokenFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&releases, 1)
		}),
	)

	sp := surgeprotector.NewSurgeProtector[int](ctx).(*surgeprotector.SurgeProtector[int])
	sp.ConnectSensor(s)
	sp.SetRateLimit(1, time.Minute, 0)

	if !sp.TryTake() {
		t.Fatalf("expected first token to be available")
	}
	if sp.TryTake() {
		t.Fatalf("expected no tokens left")
	}

	sp.ReleaseToken()
	if got := atomic.LoadInt32(&releases); got != 1 {
		t.Fatalf("expected 1 release notification, got %d", got)
	}
	if !sp.TryTake() {
		t.Fatalf("expected token after release")
	}
}

func TestSurgeProtector_GetTimeUntilNextRefill(t *testing.T) {
	ctx := context.Background()
	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.SetRateLimit(1, 40*time.Millisecond, 0)

	if !sp.TryTake() {
		t.Fatalf("expected token to be available")
	}

	if got := sp.GetTimeUntilNextRefill(); got <= 0 {
		t.Fatalf("expected positive refill duration, got %v", got)
	}

	time.Sleep(60 * time.Millisecond)
	if got := sp.GetTimeUntilNextRefill(); got != 0 {
		t.Fatalf("expected refill duration to be 0 after interval, got %v", got)
	}
}

func TestSurgeProtector_SubmitEnqueuesWhenRateLimited(t *testing.T) {
	ctx := context.Background()

	queue := resister.NewResister[int](ctx)
	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectResister(queue)
	sp.SetRateLimit(0, time.Second, 0)

	elem := types.NewElementFast(1)
	if err := sp.Submit(ctx, elem); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	if got := sp.GetResisterQueue(); got != 1 {
		t.Fatalf("expected queue length 1, got %d", got)
	}
}

func TestSurgeProtector_ResisterEnqueueDequeueErrors(t *testing.T) {
	ctx := context.Background()
	sp := surgeprotector.NewSurgeProtector[int](ctx)

	if err := sp.Enqueue(types.NewElementFast(1)); err == nil {
		t.Fatalf("expected enqueue error without resister")
	}
	if _, err := sp.Dequeue(); err == nil {
		t.Fatalf("expected dequeue error without resister")
	}

	queue := resister.NewResister[int](ctx)
	sp.ConnectResister(queue)

	if _, err := sp.Dequeue(); err == nil {
		t.Fatalf("expected dequeue error on empty queue")
	}

	elem := types.NewElementFast(9)
	if err := sp.Enqueue(elem); err != nil {
		t.Fatalf("enqueue error: %v", err)
	}
	got, err := sp.Dequeue()
	if err != nil {
		t.Fatalf("dequeue error: %v", err)
	}
	if got.Data != 9 {
		t.Fatalf("expected dequeued element 9, got %d", got.Data)
	}
}

func TestSurgeProtector_ConnectResisterNilDisables(t *testing.T) {
	ctx := context.Background()
	queue := resister.NewResister[int](ctx)

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectResister(queue)
	if !sp.IsResisterConnected() {
		t.Fatalf("expected resister to be connected")
	}

	sp.ConnectResister(nil)
	if sp.IsResisterConnected() {
		t.Fatalf("expected resister to be disconnected")
	}
	if got := sp.GetResisterQueue(); got != 0 {
		t.Fatalf("expected resister queue length 0, got %d", got)
	}
}

func TestSurgeProtector_DrainQueueBeforeEnqueue(t *testing.T) {
	ctx := context.Background()

	queue := resister.NewResister[int](ctx)
	component := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}

	elem := types.NewElementFast(5)
	_ = queue.Push(elem)

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectResister(queue)
	sp.ConnectComponent(component)
	sp.SetRateLimit(1, time.Minute, 0)

	current := types.NewElementFast(9)
	if err := sp.Submit(ctx, current); err != nil {
		t.Fatalf("submit error: %v", err)
	}

	if got := component.SubmitCount(); got != 1 {
		t.Fatalf("expected one drained submission, got %d", got)
	}
	if got := component.Last(); got != 5 {
		t.Fatalf("expected drained element 5, got %d", got)
	}
	if got := sp.GetResisterQueue(); got != 1 {
		t.Fatalf("expected current element enqueued, got %d", got)
	}
}

func TestSurgeProtector_DrainQueueFailureRequeues(t *testing.T) {
	ctx := context.Background()

	queue := resister.NewResister[int](ctx)
	elem := types.NewElementFast(2)
	_ = queue.Push(elem)

	component := &recordingSubmitter[int]{
		metadata:  types.ComponentMetadata{Type: "WIRE"},
		submitErr: errors.New("fail"),
	}

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectResister(queue)
	sp.ConnectComponent(component)
	sp.SetRateLimit(1, time.Minute, 0)

	if err := sp.Submit(ctx, types.NewElementFast(8)); err == nil {
		t.Fatalf("expected submit error from component")
	}

	if got := sp.GetResisterQueue(); got != 1 {
		t.Fatalf("expected queue length 1 after requeue, got %d", got)
	}
	requeued, err := sp.Dequeue()
	if err != nil {
		t.Fatalf("dequeue error: %v", err)
	}
	if requeued.RetryCount != 1 {
		t.Fatalf("expected retry count 1, got %d", requeued.RetryCount)
	}
}

func TestSurgeProtector_TripResetAndRestart(t *testing.T) {
	ctx := context.Background()

	var trips int32
	var resets int32
	s := sensor.NewSensor[int](
		sensor.WithSurgeProtectorTripFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&trips, 1)
		}),
		sensor.WithSurgeProtectorResetFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&resets, 1)
		}),
	)

	component := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}
	gen := &recordingGenerator[int]{metadata: types.ComponentMetadata{Type: "GENERATOR"}}
	component.ConnectGenerator(gen)

	sp := surgeprotector.NewSurgeProtector[int](ctx, surgeprotector.WithSensor[int](s))
	sp.ConnectComponent(component)

	sp.Trip()
	if !sp.IsTripped() {
		t.Fatalf("expected surge protector to be tripped")
	}

	sp.Reset()
	if sp.IsTripped() {
		t.Fatalf("expected surge protector to be reset")
	}

	if got := atomic.LoadInt32(&trips); got != 1 {
		t.Fatalf("expected 1 trip notification, got %d", got)
	}
	if got := atomic.LoadInt32(&resets); got != 1 {
		t.Fatalf("expected 1 reset notification, got %d", got)
	}

	if got := component.StartCount(); got != 1 {
		t.Fatalf("expected component to be started on reset, got %d", got)
	}
	if got := gen.StartCount(); got != 1 {
		t.Fatalf("expected generator to be started on reset, got %d", got)
	}
}

func TestSurgeProtector_ResetRestartsConduit(t *testing.T) {
	ctx := context.Background()

	conduitComponent := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "CONDUIT"}}

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectComponent(conduitComponent)

	sp.Trip()
	sp.Reset()

	if got := conduitComponent.RestartCount(); got != 1 {
		t.Fatalf("expected conduit restart count 1, got %d", got)
	}
	if got := conduitComponent.StartCount(); got != 0 {
		t.Fatalf("expected conduit start count 0, got %d", got)
	}
}

func TestSurgeProtector_ResetNotifiesRestart(t *testing.T) {
	ctx := context.Background()

	var restarts int32
	s := sensor.NewSensor[int](
		sensor.WithOnRestartFunc[int](func(types.ComponentMetadata) {
			atomic.AddInt32(&restarts, 1)
		}),
	)

	component := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}
	gen := &recordingGenerator[int]{metadata: types.ComponentMetadata{Type: "GENERATOR"}}
	component.ConnectGenerator(gen)

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectSensor(s)
	sp.ConnectComponent(component)

	sp.Trip()
	sp.Reset()

	if got := atomic.LoadInt32(&restarts); got != 2 {
		t.Fatalf("expected 2 restart callbacks, got %d", got)
	}
}

func TestSurgeProtector_ConnectResisterNotifiesSensor(t *testing.T) {
	ctx := context.Background()

	var connected int32
	s := sensor.NewSensor[int](
		sensor.WithSurgeProtectorConnectResisterFunc[int](func(types.ComponentMetadata, types.ComponentMetadata) {
			atomic.AddInt32(&connected, 1)
		}),
	)

	queue := resister.NewResister[int](ctx)
	sp := surgeprotector.NewSurgeProtector[int](ctx, surgeprotector.WithSensor[int](s))
	sp.ConnectResister(queue)

	if got := atomic.LoadInt32(&connected); got != 1 {
		t.Fatalf("expected 1 connect resister notification, got %d", got)
	}
}

func TestSurgeProtector_DetachBackupsNotifiesSensor(t *testing.T) {
	ctx := context.Background()

	var detached int32
	s := sensor.NewSensor[int](
		sensor.WithSurgeProtectorDetachedBackupsFunc[int](func(types.ComponentMetadata, types.ComponentMetadata) {
			atomic.AddInt32(&detached, 1)
		}),
	)

	backup1 := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}
	backup2 := &recordingSubmitter[int]{metadata: types.ComponentMetadata{Type: "WIRE"}}

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectSensor(s)
	sp.AttachBackup(backup1, backup2)

	backups := sp.GetBackupSystems()
	if len(backups) != 2 {
		t.Fatalf("expected 2 backups, got %d", len(backups))
	}
	backups[0] = nil
	if len(sp.GetBackupSystems()) != 2 {
		t.Fatalf("expected backup list copy to remain intact")
	}

	sp.DetachBackups()
	if got := atomic.LoadInt32(&detached); got != 2 {
		t.Fatalf("expected 2 detached notifications, got %d", got)
	}
	if len(sp.GetBackupSystems()) != 0 {
		t.Fatalf("expected backups to be detached")
	}
}

func TestSurgeProtector_CreationIncrementsMeter(t *testing.T) {
	ctx := context.Background()

	meter := newStubMeter[int]()
	s := sensor.NewSensor[int]()
	s.ConnectMeter(meter)

	sp := surgeprotector.NewSurgeProtector[int](ctx)
	sp.ConnectSensor(s)

	if got := meter.GetMetricCount(types.MetricSurgeCount); got != 1 {
		t.Fatalf("expected surge count to increment, got %d", got)
	}
}

func TestSurgeProtector_BackupFailure(t *testing.T) {
	ctx := context.Background()

	var failures int32
	s := sensor.NewSensor[int](
		sensor.WithSurgeProtectorBackupFailureFunc[int](func(types.ComponentMetadata, error) {
			atomic.AddInt32(&failures, 1)
		}),
	)

	backup := &recordingSubmitter[int]{
		metadata:  types.ComponentMetadata{Type: "WIRE"},
		submitErr: errors.New("boom"),
	}

	sp := surgeprotector.NewSurgeProtector[int](ctx, surgeprotector.WithSensor[int](s))
	sp.AttachBackup(backup)

	err := sp.Submit(ctx, types.NewElementFast(1))
	if err == nil {
		t.Fatalf("expected error from backup")
	}
	if got := atomic.LoadInt32(&failures); got != 1 {
		t.Fatalf("expected 1 backup failure notification, got %d", got)
	}
}

type recordingSubmitter[T any] struct {
	metadata   types.ComponentMetadata
	submitErr  error
	started    int32
	startCount int32
	restartCnt int32
	last       T
	submits    int32
	lock       sync.Mutex
	generators []types.Generator[T]
}

func (r *recordingSubmitter[T]) Submit(ctx context.Context, elem T) error {
	_ = ctx
	if r.submitErr != nil {
		return r.submitErr
	}
	r.lock.Lock()
	r.last = elem
	r.lock.Unlock()
	atomic.AddInt32(&r.submits, 1)
	return nil
}

func (r *recordingSubmitter[T]) ConnectGenerator(gens ...types.Generator[T]) {
	r.generators = append(r.generators, gens...)
}

func (r *recordingSubmitter[T]) GetGenerators() []types.Generator[T] { return r.generators }

func (r *recordingSubmitter[T]) ConnectLogger(...types.Logger) {}

func (r *recordingSubmitter[T]) Restart(ctx context.Context) error {
	_ = ctx
	atomic.AddInt32(&r.restartCnt, 1)
	return nil
}

func (r *recordingSubmitter[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}

func (r *recordingSubmitter[T]) GetComponentMetadata() types.ComponentMetadata { return r.metadata }

func (r *recordingSubmitter[T]) SetComponentMetadata(name string, id string) {
	r.metadata = types.ComponentMetadata{Name: name, ID: id, Type: r.metadata.Type}
}

func (r *recordingSubmitter[T]) Stop() error { return nil }

func (r *recordingSubmitter[T]) IsStarted() bool { return atomic.LoadInt32(&r.started) == 1 }

func (r *recordingSubmitter[T]) Start(ctx context.Context) error {
	_ = ctx
	atomic.StoreInt32(&r.started, 1)
	atomic.AddInt32(&r.startCount, 1)
	return nil
}

func (r *recordingSubmitter[T]) StartCount() int32 { return atomic.LoadInt32(&r.startCount) }

func (r *recordingSubmitter[T]) RestartCount() int32 { return atomic.LoadInt32(&r.restartCnt) }

func (r *recordingSubmitter[T]) SubmitCount() int32 { return atomic.LoadInt32(&r.submits) }

func (r *recordingSubmitter[T]) Last() T {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.last
}

type recordingGenerator[T any] struct {
	metadata   types.ComponentMetadata
	started    int32
	startCount int32
}

func (r *recordingGenerator[T]) ConnectCircuitBreaker(types.CircuitBreaker[T]) {}

func (r *recordingGenerator[T]) Start(ctx context.Context) error {
	_ = ctx
	atomic.StoreInt32(&r.started, 1)
	atomic.AddInt32(&r.startCount, 1)
	return nil
}

func (r *recordingGenerator[T]) Stop() error { return nil }

func (r *recordingGenerator[T]) IsStarted() bool { return atomic.LoadInt32(&r.started) == 1 }

func (r *recordingGenerator[T]) Restart(ctx context.Context) error {
	_ = ctx
	return nil
}

func (r *recordingGenerator[T]) ConnectPlug(...types.Plug[T]) {}

func (r *recordingGenerator[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}

func (r *recordingGenerator[T]) ConnectLogger(...types.Logger) {}

func (r *recordingGenerator[T]) ConnectToComponent(...types.Submitter[T]) {}

func (r *recordingGenerator[T]) GetComponentMetadata() types.ComponentMetadata { return r.metadata }

func (r *recordingGenerator[T]) SetComponentMetadata(name string, id string) {
	r.metadata = types.ComponentMetadata{Name: name, ID: id, Type: r.metadata.Type}
}

func (r *recordingGenerator[T]) ConnectSensor(...types.Sensor[T]) {}

func (r *recordingGenerator[T]) StartCount() int32 { return atomic.LoadInt32(&r.startCount) }

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
func (s *stubMeter[T]) GetMetricTimestamp(metricName string) int64          { return s.timestamps[metricName] }
func (s *stubMeter[T]) SetMetricPercentage(name string, percentage float64) {}
func (s *stubMeter[T]) GetMetricNames() []string                            { return nil }
func (s *stubMeter[T]) ReportData()                                         {}
func (s *stubMeter[T]) SetIdleTimeout(to time.Duration)                     {}
func (s *stubMeter[T]) SetContextCancelHook(hook func())                    {}
func (s *stubMeter[T]) SetDynamicMetric(metricName string, total uint64, initialCount uint64, threshold float64) {
}
func (s *stubMeter[T]) GetDynamicMetricInfo(metricName string) (*types.MetricInfo, bool) {
	return nil, false
}
func (s *stubMeter[T]) SetMetricPeak(metricName string, count uint64) {}
func (s *stubMeter[T]) SetDynamicMetricTotal(metricName string, total uint64) {
}
func (s *stubMeter[T]) GetTicker() *time.Ticker                        { return nil }
func (s *stubMeter[T]) SetTicker(ticker *time.Ticker)                  {}
func (s *stubMeter[T]) GetOriginalContext() context.Context            { return nil }
func (s *stubMeter[T]) GetOriginalContextCancel() context.CancelFunc   { return nil }
func (s *stubMeter[T]) Monitor()                                       {}
func (s *stubMeter[T]) AddTotalItems(additionalTotal uint64)           {}
func (s *stubMeter[T]) CheckMetrics() bool                             { return true }
func (s *stubMeter[T]) PauseProcessing()                               {}
func (s *stubMeter[T]) ResumeProcessing()                              {}
func (s *stubMeter[T]) GetMetricCount(metricName string) uint64        { return s.counts[metricName] }
func (s *stubMeter[T]) SetMetricCount(metricName string, count uint64) { s.counts[metricName] = count }
func (s *stubMeter[T]) GetMetricTotal(metricName string) uint64        { return 0 }
func (s *stubMeter[T]) SetMetricTotal(metricName string, total uint64) {}
func (s *stubMeter[T]) IncrementCount(metricName string)               { s.counts[metricName]++ }
func (s *stubMeter[T]) DecrementCount(metricName string)               { s.counts[metricName]-- }
func (s *stubMeter[T]) IsTimerRunning(metricName string) bool          { return false }
func (s *stubMeter[T]) GetTimerStartTime(metricName string) (time.Time, bool) {
	return time.Time{}, false
}
func (s *stubMeter[T]) SetTimerStartTime(metricName string, startTime time.Time) {}
func (s *stubMeter[T]) SetTotal(metricName string, total uint64)                 {}
func (s *stubMeter[T]) StartTimer(metricName string)                             {}
func (s *stubMeter[T]) StopTimer(metricName string) time.Duration                { return 0 }
func (s *stubMeter[T]) GetComponentMetadata() types.ComponentMetadata {
	return types.ComponentMetadata{}
}
func (s *stubMeter[T]) ConnectLogger(...types.Logger) {}
func (s *stubMeter[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {
}
func (s *stubMeter[T]) SetComponentMetadata(name string, id string) {}
func (s *stubMeter[T]) ResetMetrics()                               {}
