package receivingrelay_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

type sample struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestUnwrapPayloadJSON(t *testing.T) {
	in := sample{ID: 7, Name: "delta"}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("json marshal error: %v", err)
	}

	wp := &relay.WrappedPayload{
		Payload: b,
		Metadata: &relay.MessageMetadata{
			ContentType: "application/json",
		},
	}

	var out sample
	if err := receivingrelay.UnwrapPayload(wp, "", &out); err != nil {
		t.Fatalf("UnwrapPayload error: %v", err)
	}
	if out != in {
		t.Fatalf("json decode mismatch: got %+v want %+v", out, in)
	}
}

func TestReceivingRelayReceive(t *testing.T) {
	ctx := context.Background()
	rr := receivingrelay.NewReceivingRelay[sample](ctx)
	impl, ok := rr.(*receivingrelay.ReceivingRelay[sample])
	if !ok {
		t.Fatalf("expected ReceivingRelay implementation")
	}

	in := sample{ID: 42, Name: "alpha"}
	wp, err := forwardrelay.WrapPayload(in, nil, nil, "")
	if err != nil {
		t.Fatalf("WrapPayload error: %v", err)
	}

	ack, err := rr.Receive(ctx, wp)
	if err != nil {
		t.Fatalf("Receive error: %v", err)
	}
	if ack == nil || !ack.Success {
		t.Fatalf("expected success ack")
	}

	select {
	case got := <-impl.GetDataChannel():
		if got != in {
			t.Fatalf("received mismatch: got %+v want %+v", got, in)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for receive")
	}
}
