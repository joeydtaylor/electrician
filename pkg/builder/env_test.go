package builder

import (
	"os"
	"testing"
)

func TestEnvOr(t *testing.T) {
	const key = "ELECTRICIAN_TEST_ENV_OR"
	_ = os.Unsetenv(key)
	if got := EnvOr(key, "fallback"); got != "fallback" {
		t.Fatalf("expected fallback, got %q", got)
	}

	if err := os.Setenv(key, `"  value  "`); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv(key) })

	if got := EnvOr(key, "fallback"); got != "value" {
		t.Fatalf("expected trimmed value, got %q", got)
	}
}

func TestEnvIntOr(t *testing.T) {
	const key = "ELECTRICIAN_TEST_ENV_INT"
	_ = os.Unsetenv(key)
	if got := EnvIntOr(key, 7); got != 7 {
		t.Fatalf("expected default int, got %d", got)
	}

	if err := os.Setenv(key, "12"); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv(key) })

	if got := EnvIntOr(key, 7); got != 12 {
		t.Fatalf("expected 12, got %d", got)
	}

	if err := os.Setenv(key, "not-int"); err != nil {
		t.Fatalf("setenv failed: %v", err)
	}
	if got := EnvIntOr(key, 7); got != 7 {
		t.Fatalf("expected default on bad int, got %d", got)
	}
}
