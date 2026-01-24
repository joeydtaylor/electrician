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
