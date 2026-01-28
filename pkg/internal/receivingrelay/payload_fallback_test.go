package receivingrelay

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

type fallbackPayload struct {
	Name string
	ID   int
}

func TestUnwrapPayloadRequiresEncryptionWhenKeySet(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(fallbackPayload{Name: "alpha", ID: 7}); err != nil {
		t.Fatalf("gob encode: %v", err)
	}

	wp := &relay.WrappedPayload{
		Payload:         buf.Bytes(),
		PayloadEncoding: relay.PayloadEncoding_PAYLOAD_ENCODING_GOB,
		Metadata:        &relay.MessageMetadata{},
	}

	var out fallbackPayload
	if err := UnwrapPayload(wp, key, &out); err == nil {
		t.Fatalf("expected encryption required error")
	}
}
