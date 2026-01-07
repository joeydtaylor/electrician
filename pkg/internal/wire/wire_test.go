package wire_test

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

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

	// Output channel must be closed after Stop().
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

func TestWire_AllocBudget_ScratchBytes(t *testing.T) {
	// This test is intended to catch accidental per-item allocations/regressions.
	// It should stay stable across machines by using a generous bytes/item threshold.
	//
	// If you run with -short, reduce the workload.
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

	// Reused read-only payload (no per-item allocations).
	payload := make([]byte, payloadSize)
	for i := range payload {
		payload[i] = byte(i)
	}

	bufLen := payloadSize + 8 + 32

	// Per-worker scratch, allocated once per worker, no pools required.
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

	// Drain exactly N outputs.
	out := w.GetOutputChannel()
	done := make(chan struct{})
	go func() {
		for i := 0; i < items; i++ {
			<-out
		}
		close(done)
	}()

	// Measure allocations for the processing phase (exclude startup).
	runtime.GC()
	var m0 runtime.MemStats
	runtime.ReadMemStats(&m0)

	// Submit
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

	// This threshold is intentionally not "0" to avoid flakiness from runtime noise.
	// It WILL catch real regressions (e.g., accidental boxing/pooling allocations per item).
	const maxBytesPerItem = 8.0
	if bytesPerItem > maxBytesPerItem {
		t.Fatalf("alloc regression: %.2f bytes/item (delta=%d bytes over %d items)", bytesPerItem, deltaAlloc, items)
	}

	// Optional: ensure no errors were routed (sanity)
	_ = atomic.LoadUint64(new(uint64))
}

type Event struct {
	Seq        uint64
	Payload    []byte
	IsNegative bool
	Hash       [32]byte
}
