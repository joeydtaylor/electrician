package httpclient

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type testPayload struct {
	Message string `json:"message"`
}

func TestHTTPClientAdapter_FetchJSON(t *testing.T) {
	adapter := NewHTTPClientAdapter[testPayload](context.Background(), WithRequestConfig[testPayload]("GET", "https://example.test/api", nil)).(*HTTPClientAdapter[testPayload])
	adapter.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			if r.Method != http.MethodGet {
				return nil, fmt.Errorf("unexpected method %s", r.Method)
			}
			return jsonResponse(http.StatusOK, `{"message":"ok"}`), nil
		}),
	}

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
	adapter := NewHTTPClientAdapter[testPayload](context.Background(), WithRequestConfig[testPayload]("GET", "https://example.test/fail", nil)).(*HTTPClientAdapter[testPayload])
	adapter.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusInternalServerError, `{"error":"fail"}`), nil
		}),
	}

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

	adapter := NewHTTPClientAdapter[testPayload](context.Background(), WithRequestConfig[testPayload]("GET", "https://example.test/headers", nil)).(*HTTPClientAdapter[testPayload])
	adapter.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			gotAuth = r.Header.Get("Authorization")
			gotCustom = r.Header.Get("X-Test")
			return jsonResponse(http.StatusOK, `{"message":"ok"}`), nil
		}),
	}
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

	adapter := NewHTTPClientAdapter[testPayload](context.Background()).(*HTTPClientAdapter[testPayload])
	adapter.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&hits, 1)
			body, _ := io.ReadAll(r.Body)
			values, _ := url.ParseQuery(string(body))
			if values.Get("grant_type") != "client_credentials" {
				return nil, fmt.Errorf("unexpected grant_type %s", values.Get("grant_type"))
			}
			return jsonResponse(http.StatusOK, `{"access_token":"token","expires_in":3600}`), nil
		}),
	}

	if err := adapter.EnsureValidToken("id", "secret", "https://auth.test/token", "audience"); err != nil {
		t.Fatalf("EnsureValidToken() error: %v", err)
	}
	if err := adapter.EnsureValidToken("id", "secret", "https://auth.test/token", "audience"); err != nil {
		t.Fatalf("EnsureValidToken() error: %v", err)
	}

	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 token request, got %d", hits)
	}
}

func TestHTTPClientAdapter_AttemptFetchAndSubmit(t *testing.T) {
	adapter := NewHTTPClientAdapter[int](context.Background(), WithRequestConfig[int]("GET", "https://example.test/fail", nil)).(*HTTPClientAdapter[int])
	adapter.httpClient = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusInternalServerError, `{"error":"fail"}`), nil
		}),
	}

	err := adapter.attemptFetchAndSubmit(context.Background(), func(context.Context, int) error {
		return nil
	}, 0)
	if err == nil {
		t.Fatalf("expected error")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
