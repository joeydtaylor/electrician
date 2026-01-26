package builder

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestRelaySinkConfigToSinkConfig(t *testing.T) {
	tlsCfg := NewTlsServerConfig(true, "cert", "key", "ca", "example", tls.VersionTLS12, tls.VersionTLS13)
	drop := BoolPtr(false)
	auth := BoolPtr(false)

	cfg := RelaySinkConfig{
		Targets:        []string{"a:1", "b:2"},
		QueueSize:      50,
		SubmitTimeout:  2 * time.Second,
		FlushTimeout:   5 * time.Second,
		DropOnFull:     drop,
		StaticHeaders:  map[string]string{"x-tenant": "dev"},
		TLS:            tlsCfg,
		BearerTokenEnv: "TOKEN_ENV",
		AuthRequired:   auth,
	}

	sink := cfg.ToSinkConfig()
	if sink.Type != string(RelaySink) {
		t.Fatalf("expected relay sink type, got %q", sink.Type)
	}

	targets, ok := sink.Config["targets"].([]string)
	if !ok || len(targets) != 2 {
		t.Fatalf("expected targets list, got %#v", sink.Config["targets"])
	}

	if got := sink.Config["queue_size"]; got != 50 {
		t.Fatalf("expected queue_size 50, got %#v", got)
	}
	if got := sink.Config["submit_timeout"]; got != 2*time.Second {
		t.Fatalf("expected submit_timeout 2s, got %#v", got)
	}
	if got := sink.Config["flush_timeout"]; got != 5*time.Second {
		t.Fatalf("expected flush_timeout 5s, got %#v", got)
	}
	if got := sink.Config["drop_on_full"]; got != false {
		t.Fatalf("expected drop_on_full false, got %#v", got)
	}
	if got := sink.Config["auth_required"]; got != false {
		t.Fatalf("expected auth_required false, got %#v", got)
	}

	headers, ok := sink.Config["static_headers"].(map[string]string)
	if !ok || headers["x-tenant"] != "dev" {
		t.Fatalf("expected static headers, got %#v", sink.Config["static_headers"])
	}

	tlsMap, ok := sink.Config["tls"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected tls map, got %#v", sink.Config["tls"])
	}
	if tlsMap["cert"] != "cert" || tlsMap["key"] != "key" || tlsMap["ca"] != "ca" {
		t.Fatalf("unexpected tls map: %#v", tlsMap)
	}
	if tlsMap["min_version"] != "1.2" || tlsMap["max_version"] != "1.3" {
		t.Fatalf("unexpected tls versions: %#v", tlsMap)
	}
	if sink.Config["bearer_token_env"] != "TOKEN_ENV" {
		t.Fatalf("expected bearer_token_env, got %#v", sink.Config["bearer_token_env"])
	}
}

func TestBoolPtr(t *testing.T) {
	val := BoolPtr(true)
	if val == nil || *val != true {
		t.Fatalf("expected true bool pointer, got %#v", val)
	}
}
