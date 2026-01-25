package httpserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type sampleRequest struct {
	Name string `json:"name"`
}

func newTestServer[T any]() *httpServerAdapter[T] {
	return &httpServerAdapter[T]{
		headers: make(map[string]string),
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{method: http.MethodPost, endpoint: "/hook", headers: map[string]string{}}

	h := adapter.buildHandler(cfg, func(ctx context.Context, req sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, nil
	})

	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestHandlerDecodeError(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{method: http.MethodPost, endpoint: "/hook", headers: map[string]string{}}

	h := adapter.buildHandler(cfg, func(ctx context.Context, req sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader("{bad"))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandlerDefaultHeaders(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{
		method:   http.MethodPost,
		endpoint: "/hook",
		headers: map[string]string{
			"X-Default": "default",
			"X-Other":   "other",
		},
	}

	h := adapter.buildHandler(cfg, func(ctx context.Context, req sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{
			Headers: map[string]string{
				"X-Default": "override",
				"X-Custom":  "custom",
			},
		}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"name":"alpha"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if got := w.Header().Get("X-Default"); got != "override" {
		t.Fatalf("expected X-Default override, got %q", got)
	}
	if got := w.Header().Get("X-Other"); got != "other" {
		t.Fatalf("expected X-Other default, got %q", got)
	}
	if got := w.Header().Get("X-Custom"); got != "custom" {
		t.Fatalf("expected X-Custom header, got %q", got)
	}
}

func TestHandlerHTTPServerError(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{method: http.MethodPost, endpoint: "/hook", headers: map[string]string{}}

	h := adapter.buildHandler(cfg, func(ctx context.Context, req sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, &types.HTTPServerError{StatusCode: http.StatusTeapot, Message: "nope"}
	})

	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"name":"alpha"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusTeapot {
		t.Fatalf("expected status %d, got %d", http.StatusTeapot, w.Code)
	}
	if !strings.Contains(w.Body.String(), "nope") {
		t.Fatalf("expected response body to include message")
	}
}
