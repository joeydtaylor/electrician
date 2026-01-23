package wire_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

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
