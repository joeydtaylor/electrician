package forwardrelay_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type stubReceiver[T any] struct {
	ch      chan T
	started int32
	meta    types.ComponentMetadata
}

func newStubReceiver[T any]() *stubReceiver[T] {
	return &stubReceiver[T]{
		ch: make(chan T),
		meta: types.ComponentMetadata{
			ID:   "stub",
			Type: "STUB_RECEIVER",
		},
	}
}

func (s *stubReceiver[T]) ConnectCircuitBreaker(types.CircuitBreaker[T])        {}
func (s *stubReceiver[T]) ConnectLogger(...types.Logger)                        {}
func (s *stubReceiver[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {}
func (s *stubReceiver[T]) GetComponentMetadata() types.ComponentMetadata        { return s.meta }
func (s *stubReceiver[T]) SetComponentMetadata(name string, id string) {
	s.meta.Name = name
	s.meta.ID = id
}
func (s *stubReceiver[T]) GetOutputChannel() chan T { return s.ch }
func (s *stubReceiver[T]) IsStarted() bool          { return atomic.LoadInt32(&s.started) == 1 }
func (s *stubReceiver[T]) Start(context.Context) error {
	atomic.StoreInt32(&s.started, 1)
	return nil
}

func TestForwardRelayDefaults(t *testing.T) {
	ctx := context.Background()
	fr := forwardrelay.NewForwardRelay[string](ctx)
	impl, ok := fr.(*forwardrelay.ForwardRelay[string])
	if !ok {
		t.Fatalf("expected ForwardRelay implementation")
	}

	if !fr.GetAuthRequired() {
		t.Fatalf("expected auth required by default")
	}
	if got := fr.GetTargets(); len(got) != 0 {
		t.Fatalf("expected no targets, got %v", got)
	}
	if impl.GetPerformanceOptions() == nil {
		t.Fatalf("expected performance options to be set")
	}
}

func TestForwardRelayFreeze(t *testing.T) {
	ctx := context.Background()
	input := newStubReceiver[string]()

	fr := forwardrelay.NewForwardRelay[string](ctx, forwardrelay.WithInput[string](input))
	if err := fr.Start(ctx); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer fr.Stop()

	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected panic when mutating after Start")
		}
	}()

	fr.SetTargets("localhost:50051")
}

func TestWrapPayloadRoundTrip(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	perf := &relay.PerformanceOptions{UseCompression: true, CompressionAlgorithm: forwardrelay.COMPRESS_SNAPPY}
	sec := &relay.SecurityOptions{Enabled: true, Suite: forwardrelay.ENCRYPT_AES_GCM}

	type sample struct {
		ID   int
		Name string
	}
	in := sample{ID: 42, Name: "alpha"}

	wp, err := forwardrelay.WrapPayload(in, perf, sec, key)
	if err != nil {
		t.Fatalf("WrapPayload error: %v", err)
	}

	var out sample
	if err := receivingrelay.UnwrapPayload(wp, key, &out); err != nil {
		t.Fatalf("UnwrapPayload error: %v", err)
	}
	if out != in {
		t.Fatalf("round-trip mismatch: got %+v want %+v", out, in)
	}
}

func TestWrapPayloadRequiresEncryptionWhenKeyProvided(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	if _, err := forwardrelay.WrapPayload("hi", nil, nil, key); err == nil {
		t.Fatalf("expected error when key is provided without encryption enabled")
	}
}

func TestWrapPayloadRequiresKeyWhenEncryptionEnabled(t *testing.T) {
	sec := &relay.SecurityOptions{Enabled: true, Suite: forwardrelay.ENCRYPT_AES_GCM}
	if _, err := forwardrelay.WrapPayload("hi", nil, sec, ""); err == nil {
		t.Fatalf("expected error when encryption is enabled without a key")
	}
}
