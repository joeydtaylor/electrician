//go:build integration

package websocketrelay

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

func freeAddr(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	return addr
}

func waitForReadyWS(t *testing.T, url string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
		fr := NewForwardRelay[feedback](ctx, func(fr *ForwardRelay[feedback]) {
			fr.SetTargets(url)
			fr.SetAckMode(relay.AckMode_ACK_NONE)
		})
		_, err := fr.getOrCreateSession(ctx, url)
		cancel()
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for websocket listener: %v", err)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func waitForMessage[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatalf("timeout waiting for message")
	}
	var zero T
	return zero
}

func TestWebSocketRelayRoundTripJSON(t *testing.T) {
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[feedback](ctx,
		func(rr *ReceivingRelay[feedback]) {
			rr.SetAddress(addr)
		},
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	url := fmt.Sprintf("ws://%s/relay", addr)
	waitForReadyWS(t, url)

	fr := NewForwardRelay[feedback](ctx,
		func(fr *ForwardRelay[feedback]) { fr.SetTargets(url) },
		func(fr *ForwardRelay[feedback]) { fr.SetPayloadFormat("json") },
		func(fr *ForwardRelay[feedback]) { fr.SetOmitPayloadMetadata(false) },
	)

	in := feedback{CustomerID: "cust-1", Content: "hello", Category: "feedback", IsNegative: false, Tags: []string{"ws"}}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}

	out := waitForMessage(t, recv.DataCh)
	if out.CustomerID != in.CustomerID || out.Content != in.Content || out.IsNegative != in.IsNegative {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}

func TestWebSocketRelayEncryptionAESGCM(t *testing.T) {
	addr := freeAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	key := "0123456789abcdef0123456789abcdef"
	sec := &relay.SecurityOptions{Enabled: true, Suite: relay.EncryptionSuite_ENCRYPTION_AES_GCM}

	recv := NewReceivingRelay[feedback](ctx,
		func(rr *ReceivingRelay[feedback]) { rr.SetAddress(addr) },
		func(rr *ReceivingRelay[feedback]) { rr.SetDecryptionKey(key) },
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	url := fmt.Sprintf("ws://%s/relay", addr)
	waitForReadyWS(t, url)

	fr := NewForwardRelay[feedback](ctx,
		func(fr *ForwardRelay[feedback]) { fr.SetTargets(url) },
		func(fr *ForwardRelay[feedback]) { fr.SetPayloadFormat("json") },
		func(fr *ForwardRelay[feedback]) { fr.SetSecurityOptions(sec, key) },
		func(fr *ForwardRelay[feedback]) { fr.SetOmitPayloadMetadata(false) },
	)

	in := feedback{CustomerID: "cust-2", Content: "secure", Category: "feedback", IsNegative: true}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}

	out := waitForMessage(t, recv.DataCh)
	if out.CustomerID != in.CustomerID || out.Content != in.Content || out.IsNegative != in.IsNegative {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}

var _ types.ForwardRelay[feedback] = (*ForwardRelay[feedback])(nil)
