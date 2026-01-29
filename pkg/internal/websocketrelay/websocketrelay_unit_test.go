package websocketrelay

import (
	"context"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

type feedbackUnit struct {
	CustomerID string `json:"customerId"`
	Content    string `json:"content"`
}

func TestWebSocketWrapUnwrapJSON(t *testing.T) {
	ctx := context.Background()
	fr := NewForwardRelay[feedbackUnit](ctx, func(fr *ForwardRelay[feedbackUnit]) {
		fr.SetPayloadFormat("json")
	})

	in := feedbackUnit{CustomerID: "cust-1", Content: "hello"}
	wp, err := fr.wrapItem(ctx, in)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}

	var out feedbackUnit
	if err := receivingrelay.UnwrapPayload(wp, "", &out); err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if out.CustomerID != in.CustomerID || out.Content != in.Content {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}

func TestWebSocketWrapUnwrapEncrypted(t *testing.T) {
	ctx := context.Background()
	key := "0123456789abcdef0123456789abcdef"
	sec := &relay.SecurityOptions{Enabled: true, Suite: relay.EncryptionSuite_ENCRYPTION_AES_GCM}

	fr := NewForwardRelay[feedbackUnit](ctx, func(fr *ForwardRelay[feedbackUnit]) {
		fr.SetPayloadFormat("json")
		fr.SetSecurityOptions(sec, key)
		fr.SetOmitPayloadMetadata(false)
	})

	in := feedbackUnit{CustomerID: "cust-2", Content: "secure"}
	wp, err := fr.wrapItem(ctx, in)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}

	var out feedbackUnit
	if err := receivingrelay.UnwrapPayload(wp, key, &out); err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if out.CustomerID != in.CustomerID || out.Content != in.Content {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}
