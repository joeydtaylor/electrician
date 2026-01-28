package receivingrelay

import (
	"context"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"google.golang.org/grpc/metadata"
)

func waitForWrappedPayload(t *testing.T, ch <-chan *relay.WrappedPayload) *relay.WrappedPayload {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(200 * time.Millisecond):
		t.Fatalf("timeout waiting for payload")
	}
	return nil
}

func TestStreamReceivePreservesTraceIDFromPayload(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("trace-id", "ctx-trace"))
	rr := NewReceivingRelay[*relay.WrappedPayload](ctx).(*ReceivingRelay[*relay.WrappedPayload])
	rr.SetPassthrough(true)

	stream := newBidiStream(
		ctx,
		&relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Open{Open: &relay.StreamOpen{
			StreamId: "s-trace",
			Defaults: &relay.MessageMetadata{TraceId: "default-trace"},
			AckMode:  relay.AckMode_ACK_PER_MESSAGE,
		}}},
		&relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Payload{Payload: &relay.WrappedPayload{
			Id: "p-trace",
			Metadata: &relay.MessageMetadata{
				TraceId: "payload-trace",
			},
		}}},
		&relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Close{Close: &relay.StreamClose{Reason: "done"}}},
	)

	if err := rr.StreamReceive(stream); err != nil {
		t.Fatalf("StreamReceive error: %v", err)
	}

	out := waitForWrappedPayload(t, rr.DataCh)
	if out.GetMetadata().GetTraceId() != "payload-trace" {
		t.Fatalf("expected trace_id from payload, got %q", out.GetMetadata().GetTraceId())
	}
}

func TestStreamReceiveUsesDefaultTraceIDWhenPayloadMissing(t *testing.T) {
	ctx := context.Background()
	rr := NewReceivingRelay[*relay.WrappedPayload](ctx).(*ReceivingRelay[*relay.WrappedPayload])
	rr.SetPassthrough(true)

	stream := newBidiStream(
		ctx,
		&relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Open{Open: &relay.StreamOpen{
			StreamId: "s-default",
			Defaults: &relay.MessageMetadata{TraceId: "default-trace"},
			AckMode:  relay.AckMode_ACK_PER_MESSAGE,
		}}},
		&relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Payload{Payload: &relay.WrappedPayload{
			Id: "p-default",
		}}},
		&relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Close{Close: &relay.StreamClose{Reason: "done"}}},
	)

	if err := rr.StreamReceive(stream); err != nil {
		t.Fatalf("StreamReceive error: %v", err)
	}

	out := waitForWrappedPayload(t, rr.DataCh)
	if out.GetMetadata().GetTraceId() != "default-trace" {
		t.Fatalf("expected trace_id from defaults, got %q", out.GetMetadata().GetTraceId())
	}
}
