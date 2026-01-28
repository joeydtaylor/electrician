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

func TestUnwrapPayloadFallbackDecryptMissingSecurity(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(fallbackPayload{Name: "alpha", ID: 7}); err != nil {
		t.Fatalf("gob encode: %v", err)
	}

	ciphertext := encryptAESGCM(t, key, buf.Bytes())
	wp := &relay.WrappedPayload{
		Payload:         ciphertext,
		PayloadEncoding: relay.PayloadEncoding_PAYLOAD_ENCODING_GOB,
		Metadata:        &relay.MessageMetadata{},
	}

	var out fallbackPayload
	if err := UnwrapPayload(wp, key, &out); err != nil {
		t.Fatalf("unwrap failed: %v", err)
	}
	if out.Name != "alpha" || out.ID != 7 {
		t.Fatalf("unexpected output: %+v", out)
	}
}
