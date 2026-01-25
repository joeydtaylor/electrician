package wire_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/sensor"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/wire"
)

func TestWire_StartStop(t *testing.T) {
	ctx := context.Background()

	w := wire.NewWire[int](ctx, wire.WithConcurrencyControl[int](1024, 1))
	w.ConnectTransformer(func(v int) (int, error) { return v, nil })

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	if !w.IsStarted() {
		t.Fatalf("expected wire to be started")
	}

	if err := w.Stop(); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
	if w.IsStarted() {
		t.Fatalf("expected wire to be stopped")
	}

	select {
	case _, ok := <-w.GetOutputChannel():
		if ok {
			t.Fatalf("expected output channel to be closed after Stop()")
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("output channel did not close after Stop()")
	}
}

func TestWire_Submit_ErrorDoesNotStopPipeline(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	w := wire.NewWire[int](ctx, wire.WithConcurrencyControl[int](1024, 1))

	w.ConnectTransformer(func(v int) (int, error) {
		if v == 42 {
			return 0, fmt.Errorf("simulated error")
		}
		return v, nil
	})

	_ = w.Start(ctx)
	defer func() { _ = w.Stop() }()

	_ = w.Submit(ctx, 42)
	_ = w.Submit(ctx, 24)

	select {
	case got := <-w.GetOutputChannel():
		if got != 24 {
			t.Fatalf("expected 24, got %d", got)
		}
	case <-ctx.Done():
		t.Fatalf("timed out waiting for output")
	}
}

func TestWire_ConfigurationPanicsAfterStart(t *testing.T) {
	ctx := context.Background()

	w := wire.NewWire[int](ctx).(*wire.Wire[int])
	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = w.Stop() }()

	assertPanics(t, "ConnectCircuitBreaker", func() {
		w.ConnectCircuitBreaker(types.CircuitBreaker[int](nil))
	})
	assertPanics(t, "ConnectGenerator", func() {
		w.ConnectGenerator(types.Generator[int](nil))
	})
	assertPanics(t, "ConnectLogger", func() {
		w.ConnectLogger(types.Logger(nil))
	})
	assertPanics(t, "ConnectSensor", func() {
		w.ConnectSensor(types.Sensor[int](nil))
	})
	assertPanics(t, "ConnectSurgeProtector", func() {
		w.ConnectSurgeProtector(types.SurgeProtector[int](nil))
	})
	assertPanics(t, "ConnectTransformer", func() {
		w.ConnectTransformer(types.Transformer[int](nil))
	})
	assertPanics(t, "SetComponentMetadata", func() {
		w.SetComponentMetadata("wire", "wire-1")
	})
	assertPanics(t, "SetConcurrencyControl", func() {
		w.SetConcurrencyControl(1, 1)
	})
	assertPanics(t, "SetEncoder", func() {
		w.SetEncoder(types.Encoder[int](nil))
	})
	assertPanics(t, "SetInsulator", func() {
		w.SetInsulator(nil, 1, time.Millisecond)
	})
	assertPanics(t, "SetOutputBuffer", func() {
		w.SetOutputBuffer(bytes.Buffer{})
	})
	assertPanics(t, "SetErrorChannel", func() {
		w.SetErrorChannel(make(chan types.ElementError[int]))
	})
	assertPanics(t, "SetInputChannel", func() {
		w.SetInputChannel(make(chan int))
	})
	assertPanics(t, "SetOutputChannel", func() {
		w.SetOutputChannel(make(chan int))
	})
	assertPanics(t, "SetTransformerFactory", func() {
		w.SetTransformerFactory(func() types.Transformer[int] {
			return func(v int) (int, error) { return v, nil }
		})
	})
}

func TestWire_AllocBudget_ScratchBytes(t *testing.T) {
	// Enforce a per-item allocation budget for the scratch-bytes path.
	items := 1_000_000
	if testing.Short() {
		items = 200_000
	}

	const (
		queue       = 65_536
		workers     = 12
		payloadSize = 256
		rounds      = 1
	)

	ctx := context.Background()

	payload := make([]byte, payloadSize)
	for i := range payload {
		payload[i] = byte(i)
	}

	bufLen := payloadSize + 8 + 32

	w := wire.NewWire[Event](ctx,
		wire.WithConcurrencyControl[Event](queue, workers),
		wire.WithScratchBytes[Event](bufLen, func(buf []byte, e Event) (Event, error) {
			copy(buf[:payloadSize], e.Payload)
			binary.LittleEndian.PutUint64(buf[payloadSize:payloadSize+8], e.Seq)

			var prev [32]byte
			for i := 0; i < rounds; i++ {
				copy(buf[payloadSize+8:], prev[:])
				prev = sha256.Sum256(buf)
			}

			e.Hash = prev
			return e, nil
		}),
	)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	out := w.GetOutputChannel()
	done := make(chan struct{})
	go func() {
		for i := 0; i < items; i++ {
			<-out
		}
		close(done)
	}()

	runtime.GC()
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)

	for i := 0; i < items; i++ {
		_ = w.Submit(ctx, Event{
			Seq:        uint64(i),
			Payload:    payload,
			IsNegative: (i%10 == 0),
		})
	}

	<-done
	_ = w.Stop()

	runtime.GC()
	var m1 runtime.MemStats
	runtime.ReadMemStats(&m1)

	deltaAlloc := m1.TotalAlloc - m0.TotalAlloc
	bytesPerItem := float64(deltaAlloc) / float64(items)

	const maxBytesPerItem = 8.0
	if bytesPerItem > maxBytesPerItem {
		t.Fatalf("alloc regression: %.2f bytes/item (delta=%d bytes over %d items)", bytesPerItem, deltaAlloc, items)
	}

	_ = atomic.LoadUint64(new(uint64))
}

func TestWire_LoadAsJSONArray(t *testing.T) {
	ctx := context.Background()
	w := wire.NewWire[int](ctx,
		wire.WithConcurrencyControl[int](16, 1),
		wire.WithTransformer[int](func(v int) (int, error) { return v * 2, nil }),
	)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	_ = w.Submit(ctx, 1)
	_ = w.Submit(ctx, 2)

	waitFor(t, 500*time.Millisecond, func() bool {
		return len(w.GetOutputChannel()) >= 2
	}, "expected output to be ready")

	data, err := w.LoadAsJSONArray()
	if err != nil {
		t.Fatalf("LoadAsJSONArray error: %v", err)
	}

	var got []int
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("json unmarshal error: %v", err)
	}

	expected := []int{2, 4}
	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected %v, got %v", expected, got)
	}
}

func TestWire_LoadCopiesOutputBuffer(t *testing.T) {
	ctx := context.Background()
	w := wire.NewWire[int](ctx,
		wire.WithConcurrencyControl[int](16, 1),
		wire.WithTransformer[int](func(v int) (int, error) { return v, nil }),
		wire.WithEncoder[int](testEncoder{}),
	)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	_ = w.Submit(ctx, 1)
	_ = w.Submit(ctx, 2)

	waitFor(t, 500*time.Millisecond, func() bool {
		return len(w.GetOutputChannel()) >= 2
	}, "expected output to be ready")

	buf := w.Load()
	if buf == nil {
		t.Fatal("expected non-nil buffer")
	}

	got := buf.String()
	if got != "1,2," {
		t.Fatalf("expected encoded buffer \"1,2,\", got %q", got)
	}
}

func TestWire_EncoderErrorReportsToSensor(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	s := sensor.NewSensor[int]()
	var errorsSeen int32
	s.RegisterOnError(func(cm types.ComponentMetadata, err error, elem int) {
		atomic.AddInt32(&errorsSeen, 1)
	})

	w := wire.NewWire[int](ctx,
		wire.WithConcurrencyControl[int](16, 1),
		wire.WithTransformer[int](func(v int) (int, error) { return v, nil }),
		wire.WithEncoder[int](testEncoder{err: errors.New("encode error")}),
		wire.WithSensor[int](s),
	)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = w.Stop() }()

	_ = w.Submit(ctx, 1)

	waitFor(t, 500*time.Millisecond, func() bool {
		return atomic.LoadInt32(&errorsSeen) > 0
	}, "expected error callback")
}

func TestWire_RestartResetsChannels(t *testing.T) {
	ctx := context.Background()
	w := wire.NewWire[int](ctx,
		wire.WithConcurrencyControl[int](16, 1),
		wire.WithTransformer[int](func(v int) (int, error) { return v + 1, nil }),
	)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	_ = w.Submit(ctx, 1)
	if got := readOutput(t, w.GetOutputChannel(), 500*time.Millisecond); got != 2 {
		t.Fatalf("expected 2, got %d", got)
	}

	oldOutput := w.GetOutputChannel()
	if err := w.Restart(ctx); err != nil {
		t.Fatalf("Restart() error: %v", err)
	}

	newOutput := w.GetOutputChannel()
	if oldOutput == newOutput {
		t.Fatalf("expected output channel to be replaced on restart")
	}

	_ = w.Submit(ctx, 2)
	if got := readOutput(t, newOutput, 500*time.Millisecond); got != 3 {
		t.Fatalf("expected 3 after restart, got %d", got)
	}

	_ = w.Stop()
}

func TestWire_TransformerFactory(t *testing.T) {
	ctx := context.Background()
	w := wire.NewWire[int](ctx,
		wire.WithConcurrencyControl[int](16, 1),
		wire.WithTransformerFactory[int](func() types.Transformer[int] {
			return func(v int) (int, error) { return v * 3, nil }
		}),
	)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = w.Stop() }()

	_ = w.Submit(ctx, 2)
	if got := readOutput(t, w.GetOutputChannel(), 500*time.Millisecond); got != 6 {
		t.Fatalf("expected 6, got %d", got)
	}
}

func TestWire_TransformerFactoryNilPanics(t *testing.T) {
	ctx := context.Background()
	w := wire.NewWire[int](ctx,
		wire.WithTransformerFactory[int](func() types.Transformer[int] { return nil }),
	)

	assertPanics(t, "Start", func() {
		_ = w.Start(ctx)
	})
}

func TestWire_InsulatorRetriesSuccess(t *testing.T) {
	ctx := context.Background()

	var attempts int32
	s := sensor.NewSensor[int]()
	var attemptCalls int32
	var successCalls int32
	s.RegisterOnInsulatorAttempt(func(types.ComponentMetadata, int, int, error, error, int, int, time.Duration) {
		atomic.AddInt32(&attemptCalls, 1)
	})
	s.RegisterOnInsulatorSuccess(func(types.ComponentMetadata, int, int, error, error, int, int, time.Duration) {
		atomic.AddInt32(&successCalls, 1)
	})

	insulator := func(ctx context.Context, elem int, err error) (int, error) {
		if atomic.AddInt32(&attempts, 1) == 1 {
			return elem, errors.New("retry")
		}
		return elem + 10, nil
	}

	w := wire.NewWire[int](ctx,
		wire.WithConcurrencyControl[int](16, 1),
		wire.WithTransformer[int](func(v int) (int, error) { return v, errors.New("fail") }),
		wire.WithInsulator[int](insulator, 2, 0),
		wire.WithSensor[int](s),
	)

	if err := w.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer func() { _ = w.Stop() }()

	_ = w.Submit(ctx, 1)
	if got := readOutput(t, w.GetOutputChannel(), 500*time.Millisecond); got != 11 {
		t.Fatalf("expected recovered element 11, got %d", got)
	}

	if atomic.LoadInt32(&attemptCalls) == 0 {
		t.Fatalf("expected insulator attempt callback")
	}
	if atomic.LoadInt32(&successCalls) == 0 {
		t.Fatalf("expected insulator success callback")
	}
}

func TestWire_CircuitBreakerTripDrops(t *testing.T) {
	ctx := context.Background()

	var drops int32
	s := sensor.NewSensor[int]()
	s.RegisterOnCircuitBreakerDrop(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&drops, 1)
	})

	cb := newStubCircuitBreaker[int](false, nil)

	w := wire.NewWire[int](ctx,
		wire.WithCircuitBreaker[int](cb),
		wire.WithSensor[int](s),
	)

	if err := w.Submit(ctx, 7); err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	waitFor(t, 200*time.Millisecond, func() bool {
		return atomic.LoadInt32(&drops) > 0
	}, "expected circuit breaker drop notification")
}

func TestWire_CircuitBreakerTripSubmitsToNeutralWire(t *testing.T) {
	ctx := context.Background()

	neutral := wire.NewWire[int](ctx, wire.WithConcurrencyControl[int](1, 1))
	cb := newStubCircuitBreaker[int](false, []types.Wire[int]{neutral})

	var submissions int32
	s := sensor.NewSensor[int]()
	s.RegisterOnCircuitBreakerNeutralWireSubmission(func(types.ComponentMetadata, int) {
		atomic.AddInt32(&submissions, 1)
	})

	w := wire.NewWire[int](ctx,
		wire.WithCircuitBreaker[int](cb),
		wire.WithSensor[int](s),
	)

	if err := w.Submit(ctx, 42); err != nil {
		t.Fatalf("Submit() error: %v", err)
	}

	select {
	case got := <-neutral.GetInputChannel():
		if got != 42 {
			t.Fatalf("expected neutral wire to receive 42, got %d", got)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("neutral wire did not receive submission")
	}

	waitFor(t, 200*time.Millisecond, func() bool {
		return atomic.LoadInt32(&submissions) > 0
	}, "expected neutral wire submission notification")
}

func assertPanics(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("expected panic: %s", name)
		}
	}()
	fn()
}

type Event struct {
	Seq        uint64
	Payload    []byte
	IsNegative bool
	Hash       [32]byte
}

type testEncoder struct {
	err error
}

func (e testEncoder) Encode(w io.Writer, v int) error {
	if _, err := fmt.Fprintf(w, "%d,", v); err != nil {
		return err
	}
	return e.err
}

type stubCircuitBreaker[T any] struct {
	allow     bool
	neutral   []types.Wire[T]
	metadata  types.ComponentMetadata
	resetOnce sync.Once
	resetCh   chan struct{}
}

func newStubCircuitBreaker[T any](allow bool, neutral []types.Wire[T]) *stubCircuitBreaker[T] {
	return &stubCircuitBreaker[T]{
		allow:   allow,
		neutral: neutral,
		resetCh: make(chan struct{}),
	}
}

func (s *stubCircuitBreaker[T]) ConnectSensor(...types.Sensor[T]) {}
func (s *stubCircuitBreaker[T]) Allow() bool                      { return s.allow }
func (s *stubCircuitBreaker[T]) SetDebouncePeriod(int)            {}
func (s *stubCircuitBreaker[T]) ConnectNeutralWire(wires ...types.Wire[T]) {
	s.neutral = append(s.neutral, wires...)
}
func (s *stubCircuitBreaker[T]) ConnectLogger(types.Logger) {}
func (s *stubCircuitBreaker[T]) GetComponentMetadata() types.ComponentMetadata {
	return s.metadata
}
func (s *stubCircuitBreaker[T]) GetNeutralWires() []types.Wire[T] { return s.neutral }
func (s *stubCircuitBreaker[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {
}
func (s *stubCircuitBreaker[T]) NotifyOnReset() <-chan struct{} { return s.resetCh }
func (s *stubCircuitBreaker[T]) RecordError()                   {}
func (s *stubCircuitBreaker[T]) Reset() {
	s.resetOnce.Do(func() {
		close(s.resetCh)
	})
}
func (s *stubCircuitBreaker[T]) SetComponentMetadata(name string, id string) {
	s.metadata.Name = name
	s.metadata.ID = id
}
func (s *stubCircuitBreaker[T]) Trip() { s.allow = false }

func waitFor(t *testing.T, timeout time.Duration, cond func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting: %s", msg)
}

func readOutput(t *testing.T, ch <-chan int, timeout time.Duration) int {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(timeout):
		t.Fatalf("timeout waiting for output")
	}
	return 0
}
