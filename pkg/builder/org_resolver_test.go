package builder

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

func TestOrgIDFromJWT(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none"}`))
	payload, err := json.Marshal(map[string]any{"org": "demo-org"})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}
	body := base64.RawURLEncoding.EncodeToString(payload)
	token := header + "." + body + "."

	got, ok := OrgIDFromJWT(token, []string{"org"})
	if !ok || got != "demo-org" {
		t.Fatalf("expected demo-org, got %q (ok=%v)", got, ok)
	}
}

func TestResolveOrgIDFromEnv(t *testing.T) {
	t.Setenv("ORG_ID", "")
	t.Setenv("ASSERT_JWT", "")
	t.Setenv("BEARER_JWT", "")

	cfg := DefaultOrgResolverConfig("fallback")
	if got := ResolveOrgIDFromEnv(cfg); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}

	t.Setenv("ORG_ID", "env-org")
	if got := ResolveOrgIDFromEnv(cfg); got != "env-org" {
		t.Fatalf("expected env-org, got %q", got)
	}
}
