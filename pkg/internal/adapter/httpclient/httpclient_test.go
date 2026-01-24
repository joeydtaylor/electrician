package httpclient

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type testPayload struct {
	Message string `json:"message"`
}

func TestHTTPClientAdapter_FetchJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testPayload{Message: "ok"})
	}))
	defer server.Close()

	adapter := NewHTTPClientAdapter[testPayload](context.Background(), WithRequestConfig[testPayload]("GET", server.URL, nil)).(*HTTPClientAdapter[testPayload])

	resp, err := adapter.Fetch()
	if err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
	}
	if resp.Body.Message != "ok" {
		t.Fatalf("expected decoded message, got %q", resp.Body.Message)
	}
}

func TestHTTPClientAdapter_FetchNonSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	adapter := NewHTTPClientAdapter[testPayload](context.Background(), WithRequestConfig[testPayload]("GET", server.URL, nil)).(*HTTPClientAdapter[testPayload])

	_, err := adapter.Fetch()
	if err == nil {
		t.Fatalf("expected error")
	}

	var httpErr *types.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, httpErr.StatusCode)
	}
}

func TestHTTPClientAdapter_AddHeaderAndBasicAuth(t *testing.T) {
	var gotAuth string
	var gotCustom string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotCustom = r.Header.Get("X-Test")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(testPayload{Message: "ok"})
	}))
	defer server.Close()

	adapter := NewHTTPClientAdapter[testPayload](context.Background(), WithRequestConfig[testPayload]("GET", server.URL, nil)).(*HTTPClientAdapter[testPayload])
	adapter.AddHeader("X-Test", "value")
	adapter.SetBasicAuth("user", "pass")

	if _, err := adapter.Fetch(); err != nil {
		t.Fatalf("Fetch() error: %v", err)
	}

	expectedAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte("user:pass"))
	if gotAuth != expectedAuth {
		t.Fatalf("expected Authorization header %q, got %q", expectedAuth, gotAuth)
	}
	if gotCustom != "value" {
		t.Fatalf("expected X-Test header value, got %q", gotCustom)
	}
}

func TestHTTPClientAdapter_AddHeaderRejectsInvalidValue(t *testing.T) {
	adapter := NewHTTPClientAdapter[testPayload](context.Background()).(*HTTPClientAdapter[testPayload])
	adapter.AddHeader("X-Test", "bad\nvalue")

	adapter.configLock.Lock()
	_, exists := adapter.headers["X-Test"]
	adapter.configLock.Unlock()

	if exists {
		t.Fatalf("expected header to be rejected")
	}
}

func TestHTTPClientAdapter_OAuthTokenCaching(t *testing.T) {
	var hits int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "token",
			"expires_in":   3600,
		})
	}))
	defer server.Close()

	adapter := NewHTTPClientAdapter[testPayload](context.Background()).(*HTTPClientAdapter[testPayload])

	if err := adapter.EnsureValidToken("id", "secret", server.URL, "audience"); err != nil {
		t.Fatalf("EnsureValidToken() error: %v", err)
	}
	if err := adapter.EnsureValidToken("id", "secret", server.URL, "audience"); err != nil {
		t.Fatalf("EnsureValidToken() error: %v", err)
	}

	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 token request, got %d", hits)
	}
}

func TestHTTPClientAdapter_AttemptFetchAndSubmit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	adapter := NewHTTPClientAdapter[int](context.Background(), WithRequestConfig[int]("GET", server.URL, nil)).(*HTTPClientAdapter[int])

	err := adapter.attemptFetchAndSubmit(context.Background(), func(context.Context, int) error {
		return nil
	}, 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}
