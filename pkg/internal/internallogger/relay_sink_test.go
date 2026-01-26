package internallogger

import (
	"bytes"
	"crypto/tls"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/logschema"
)

func TestParseRelaySinkConfigMissingConfig(t *testing.T) {
	_, err := parseRelaySinkConfig(types.SinkConfig{})
	if err == nil {
		t.Fatalf("expected error for missing config")
	}
}

func TestParseRelaySinkConfigTargets(t *testing.T) {
	cfg, err := parseRelaySinkConfig(types.SinkConfig{
		Config: map[string]interface{}{
			"targets": "a:1, b:2 , ,",
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfg.targets) != 2 {
		t.Fatalf("expected 2 targets, got %d", len(cfg.targets))
	}
	if cfg.targets[0] != "a:1" || cfg.targets[1] != "b:2" {
		t.Fatalf("unexpected targets: %#v", cfg.targets)
	}
}

func TestParseRelaySinkConfigQueueSizeInvalid(t *testing.T) {
	_, err := parseRelaySinkConfig(types.SinkConfig{
		Config: map[string]interface{}{
			"targets":    []string{"a:1"},
			"queue_size": 0,
		},
	})
	if err == nil {
		t.Fatalf("expected error for invalid queue_size")
	}
}

func TestParseRelaySinkConfigTLSMapString(t *testing.T) {
	cfg, err := parseRelaySinkConfig(types.SinkConfig{
		Config: map[string]interface{}{
			"targets": []string{"a:1"},
			"tls": map[string]string{
				"cert":        "client.crt",
				"key":         "client.key",
				"ca":          "ca.crt",
				"server_name": "localhost",
				"min_version": "1.2",
				"max_version": "1.3",
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.tlsConfig == nil {
		t.Fatalf("expected tls config to be set")
	}
	if !cfg.tlsConfig.UseTLS {
		t.Fatalf("expected UseTLS true")
	}
	if cfg.tlsConfig.SubjectAlternativeName != "localhost" {
		t.Fatalf("unexpected server name: %q", cfg.tlsConfig.SubjectAlternativeName)
	}
	if cfg.tlsConfig.MinTLSVersion != tls.VersionTLS12 {
		t.Fatalf("unexpected min TLS version: %v", cfg.tlsConfig.MinTLSVersion)
	}
	if cfg.tlsConfig.MaxTLSVersion != tls.VersionTLS13 {
		t.Fatalf("unexpected max TLS version: %v", cfg.tlsConfig.MaxTLSVersion)
	}
}

func TestRelayWriteSyncerNewPayload(t *testing.T) {
	r := &relayWriteSyncer{}
	p := r.newPayload([]byte("hello"))
	if p == nil {
		t.Fatalf("expected payload")
	}
	if !bytes.Equal(p.Payload, []byte("hello")) {
		t.Fatalf("unexpected payload bytes: %q", p.Payload)
	}
	if p.Metadata == nil {
		t.Fatalf("expected metadata")
	}
	if p.Metadata.ContentType != logPayloadContentType {
		t.Fatalf("unexpected content type: %q", p.Metadata.ContentType)
	}
	if got := p.Metadata.Headers[logschema.FieldSchema]; got != logschema.SchemaID {
		t.Fatalf("unexpected schema header: %q", got)
	}
	if p.PayloadType != logschema.SchemaID {
		t.Fatalf("unexpected payload type: %q", p.PayloadType)
	}
}
