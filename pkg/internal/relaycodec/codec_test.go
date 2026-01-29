package relaycodec_test

import (
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/relaycodec"
)

func TestWrapPayloadRequiresEncryptionWhenKeyProvided(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	_, err := relaycodec.WrapPayload(map[string]string{"hello": "world"}, relaycodec.FormatJSON, "", nil, nil, key, nil)
	if err == nil {
		t.Fatalf("expected error when key is provided without encryption enabled")
	}
}

func TestWrapPayloadRequiresKeyWhenEncryptionEnabled(t *testing.T) {
	sec := &relay.SecurityOptions{Enabled: true, Suite: relay.EncryptionSuite_ENCRYPTION_AES_GCM}
	_, err := relaycodec.WrapPayload(map[string]string{"hello": "world"}, relaycodec.FormatJSON, "", nil, sec, "", nil)
	if err == nil {
		t.Fatalf("expected error when encryption is enabled without a key")
	}
}
