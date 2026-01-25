package httpserver

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type sampleRequest struct {
	Name string `json:"name"`
}

func newTestServer[T any]() *httpServerAdapter[T] {
	return &httpServerAdapter[T]{
		headers:       make(map[string]string),
		staticHeaders: make(map[string]string),
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

func TestHandlerUnauthorizedWithStaticHeaders(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{
		method:        http.MethodPost,
		endpoint:      "/hook",
		headers:       map[string]string{},
		authRequired:  true,
		staticHeaders: map[string]string{"X-Auth": "token"},
	}

	h := adapter.buildHandler(cfg, func(ctx context.Context, req sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"name":"alpha"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandlerAuthSoftFail(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{
		method:        http.MethodPost,
		endpoint:      "/hook",
		headers:       map[string]string{},
		authRequired:  false,
		staticHeaders: map[string]string{"X-Auth": "token"},
	}

	h := adapter.buildHandler(cfg, func(ctx context.Context, req sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{StatusCode: http.StatusOK}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"name":"alpha"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandlerDynamicAuthValidator(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{
		method:       http.MethodPost,
		endpoint:     "/hook",
		headers:      map[string]string{},
		authRequired: true,
		authValidator: func(ctx context.Context, headers map[string]string) error {
			if headers["authorization"] != "Bearer good" {
				return errors.New("bad token")
			}
			return nil
		},
	}

	h := adapter.buildHandler(cfg, func(ctx context.Context, req sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{StatusCode: http.StatusOK}, nil
	})

	req := httptest.NewRequest(http.MethodPost, "/hook", strings.NewReader(`{"name":"alpha"}`))
	req.Header.Set("Authorization", "Bearer good")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestServeValidationErrors(t *testing.T) {
	adapter := newTestServer[sampleRequest]()

	if err := adapter.Serve(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil submit func")
	}

	adapter.configLock.Lock()
	adapter.endpoint = "/hook"
	adapter.configLock.Unlock()
	if err := adapter.Serve(context.Background(), func(context.Context, sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, nil
	}); err == nil || !strings.Contains(err.Error(), "method not configured") {
		t.Fatalf("expected method not configured error, got %v", err)
	}

	adapter = newTestServer[sampleRequest]()
	adapter.configLock.Lock()
	adapter.method = http.MethodPost
	adapter.configLock.Unlock()
	if err := adapter.Serve(context.Background(), func(context.Context, sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, nil
	}); err == nil || !strings.Contains(err.Error(), "endpoint not configured") {
		t.Fatalf("expected endpoint not configured error, got %v", err)
	}
}

func TestServeTLSConfigError(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	adapter.configLock.Lock()
	adapter.method = http.MethodPost
	adapter.endpoint = "/hook"
	adapter.tlsConfigErr = errors.New("tls err")
	adapter.configLock.Unlock()

	if err := adapter.Serve(context.Background(), func(context.Context, sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, nil
	}); err == nil || !strings.Contains(err.Error(), "tls err") {
		t.Fatalf("expected tls config error, got %v", err)
	}
}

func TestServeAlreadyStarted(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	adapter.server = &http.Server{}

	if err := adapter.Serve(context.Background(), func(context.Context, sampleRequest) (types.HTTPServerResponse, error) {
		return types.HTTPServerResponse{}, nil
	}); err == nil || !strings.Contains(err.Error(), "already started") {
		t.Fatalf("expected already started error, got %v", err)
	}
}

func TestBuildTLSConfig(t *testing.T) {
	cfg := tlsFixture(t)
	cfg.MinTLSVersion = tls.VersionTLS12
	cfg.MaxTLSVersion = tls.VersionTLS13

	tlsCfg, err := buildTLSConfig(cfg)
	if err != nil {
		t.Fatalf("buildTLSConfig error: %v", err)
	}
	if tlsCfg == nil || len(tlsCfg.Certificates) == 0 {
		t.Fatal("expected TLS config with certificates")
	}
	if tlsCfg.MinVersion != cfg.MinTLSVersion || tlsCfg.MaxVersion != cfg.MaxTLSVersion {
		t.Fatalf("unexpected tls versions: min=%d max=%d", tlsCfg.MinVersion, tlsCfg.MaxVersion)
	}
}

func TestBuildTLSConfigDisabled(t *testing.T) {
	tlsCfg, err := buildTLSConfig(types.TLSConfig{UseTLS: false})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if tlsCfg != nil {
		t.Fatalf("expected nil TLS config")
	}
}

func TestBuildTLSConfigMissingCA(t *testing.T) {
	cfg := tlsFixture(t)
	cfg.CAFile = filepath.Join(os.TempDir(), "missing-ca.pem")
	if _, err := buildTLSConfig(cfg); err == nil {
		t.Fatalf("expected error for missing CA file")
	}
}

func TestSnapshotConfigCopiesHeaders(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	adapter.AddHeader("X-Default", "default")
	staticHeaders := map[string]string{"X-Auth": "token"}
	adapter.SetStaticHeaders(staticHeaders)

	cfg := adapter.snapshotConfig()
	adapter.configLock.Lock()
	adapter.headers["X-Default"] = "mutated"
	adapter.staticHeaders["X-Auth"] = "mutated"
	adapter.configLock.Unlock()

	if cfg.headers["X-Default"] != "default" {
		t.Fatalf("expected snapshot to retain default header, got %q", cfg.headers["X-Default"])
	}
	if cfg.staticHeaders["X-Auth"] != "token" {
		t.Fatalf("expected snapshot to retain static header, got %q", cfg.staticHeaders["X-Auth"])
	}

	adapter2 := newTestServer[sampleRequest]()
	cloneSource := map[string]string{"X-Auth": "token"}
	adapter2.SetStaticHeaders(cloneSource)
	cloneSource["X-Auth"] = "mutated"
	cfg2 := adapter2.snapshotConfig()
	if cfg2.staticHeaders["X-Auth"] != "token" {
		t.Fatalf("expected adapter to use cloned headers")
	}
}

func TestHeaderHelpers(t *testing.T) {
	if got := normalizeHeaderKey("  X-Test "); got != "x-test" {
		t.Fatalf("expected normalized header key, got %q", got)
	}

	headers := map[string]string{"x-auth": "token"}
	if err := checkStaticHeaders(headers, map[string]string{"X-Auth": "token"}); err != nil {
		t.Fatalf("expected static headers to validate, got %v", err)
	}
	if err := checkStaticHeaders(headers, map[string]string{"X-Auth": "bad"}); err == nil {
		t.Fatalf("expected static header mismatch error")
	}
}

func TestCollectHeadersAndBearerToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/hook", nil)
	req.Header.Add("X-Token", "a")
	req.Header.Add("X-Token", "b")
	req.Header.Set("Authorization", "Bearer token")

	headers := collectHeaders(req)
	if headers["x-token"] != "a" {
		t.Fatalf("expected first header value, got %q", headers["x-token"])
	}
	if got := extractBearerToken(headers); got != "token" {
		t.Fatalf("expected bearer token, got %q", got)
	}

	if got := extractBearerToken(map[string]string{"authorization": "Basic abc"}); got != "" {
		t.Fatalf("expected empty token for non-bearer auth, got %q", got)
	}
}

func TestApplyAuthPolicy(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	err := adapter.applyAuthPolicy(errors.New("bad"), false)
	if err != nil {
		t.Fatalf("expected soft-fail to return nil, got %v", err)
	}
	if err := adapter.applyAuthPolicy(errors.New("bad"), true); err == nil {
		t.Fatalf("expected error when auth required")
	}
}

func TestAuthorizeRequestSoftFail(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	cfg := serverConfig{
		authRequired:  false,
		staticHeaders: map[string]string{"X-Auth": "token"},
	}
	req := httptest.NewRequest(http.MethodPost, "/hook", nil)
	if err := adapter.authorizeRequest(context.Background(), req, cfg); err != nil {
		t.Fatalf("expected soft-fail auth to return nil, got %v", err)
	}
}

func TestAuthenticationOptionsClone(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	opts := NewAuthenticationOptionsOAuth2(&relay.OAuth2Options{
		AcceptIntrospection: true,
		IntrospectionUrl:    "https://auth.local/introspect",
		RequiredScopes:      []string{"scope-a"},
	})
	adapter.SetAuthenticationOptions(opts)

	if adapter.authOptions == opts {
		t.Fatalf("expected auth options to be cloned")
	}
	opts.Enabled = false
	opts.Oauth2.RequiredScopes[0] = "mutated"
	if !adapter.authOptions.GetEnabled() {
		t.Fatalf("expected cloned auth options to remain enabled")
	}
	if got := adapter.authOptions.GetOauth2().GetRequiredScopes()[0]; got != "scope-a" {
		t.Fatalf("expected cloned scopes to remain unchanged, got %q", got)
	}
}

func TestEnsureDefaultAuthValidator(t *testing.T) {
	adapter := newTestServer[sampleRequest]()
	adapter.authOptions = NewAuthenticationOptionsOAuth2(&relay.OAuth2Options{
		AcceptIntrospection: true,
		IntrospectionUrl:    "https://auth.local/introspect",
	})

	cfg := adapter.snapshotConfig()
	if cfg.authValidator == nil {
		t.Fatalf("expected auth validator to be installed")
	}

	adapter2 := newTestServer[sampleRequest]()
	adapter2.authOptions = NewAuthenticationOptionsOAuth2(&relay.OAuth2Options{
		AcceptIntrospection: false,
		IntrospectionUrl:    "https://auth.local/introspect",
	})
	cfg2 := adapter2.snapshotConfig()
	if cfg2.authValidator != nil {
		t.Fatalf("expected no auth validator without introspection")
	}
}

func TestOAuth2Helpers(t *testing.T) {
	dst := &relay.OAuth2Options{
		Issuer:             "old",
		ForwardBearerToken: true,
		ForwardMetadataKey: "old",
		RequiredAudience:   []string{"old"},
	}
	src := &relay.OAuth2Options{
		AcceptJwt:                 true,
		AcceptIntrospection:       true,
		Issuer:                    "new",
		JwksUri:                   "jwks",
		RequiredAudience:          []string{"new"},
		RequiredScopes:            []string{"scope"},
		IntrospectionUrl:          "https://introspect",
		IntrospectionAuthType:     "basic",
		IntrospectionClientId:     "client",
		IntrospectionClientSecret: "secret",
		ForwardMetadataKey:        "auth",
		JwksCacheSeconds:          30,
	}

	out := MergeOAuth2Options(dst, src)
	if out != dst {
		t.Fatalf("expected dst to be returned")
	}
	if dst.Issuer != "new" || dst.JwksUri != "jwks" || !dst.AcceptJwt || !dst.AcceptIntrospection {
		t.Fatalf("unexpected merge result: %+v", dst)
	}
	if dst.ForwardMetadataKey != "auth" || dst.ForwardBearerToken {
		t.Fatalf("expected forwarding settings to be overwritten, got %+v", dst)
	}
	src.RequiredAudience[0] = "mutated"
	if dst.RequiredAudience[0] != "new" {
		t.Fatalf("expected RequiredAudience to be cloned")
	}
	src.RequiredScopes[0] = "mutated"
	if dst.RequiredScopes[0] != "scope" {
		t.Fatalf("expected RequiredScopes to be cloned")
	}

	aud := []string{"aud"}
	scopes := []string{"scope"}
	opts := NewOAuth2JWTOptions("issuer", "jwks", aud, scopes, 10)
	aud[0] = "mutated"
	scopes[0] = "mutated"
	if got := opts.GetRequiredAudience()[0]; got != "aud" {
		t.Fatalf("expected audience to be cloned, got %q", got)
	}
	if got := opts.GetRequiredScopes()[0]; got != "scope" {
		t.Fatalf("expected scopes to be cloned, got %q", got)
	}
}

func TestIntrospectionValidatorCacheAndBackoff(t *testing.T) {
	opts := &relay.OAuth2Options{
		AcceptIntrospection:       true,
		IntrospectionUrl:          "https://auth.local/introspect",
		IntrospectionAuthType:     "basic",
		IntrospectionClientId:     "client",
		IntrospectionClientSecret: "secret",
		RequiredScopes:            []string{"scope-a"},
		IntrospectionCacheSeconds: 30,
	}
	v := newCachingIntrospectionValidator(opts)

	var hits int32
	errCh := make(chan error, 1)
	v.hc = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&hits, 1)
			body, _ := io.ReadAll(r.Body)
			values, _ := url.ParseQuery(string(body))
			if values.Get("token") != "token" {
				errCh <- fmt.Errorf("unexpected token: %s", values.Get("token"))
			}
			user, pass, ok := r.BasicAuth()
			if !ok || user != opts.IntrospectionClientId || pass != opts.IntrospectionClientSecret {
				errCh <- fmt.Errorf("unexpected basic auth")
			}
			return jsonResponse(http.StatusOK, `{"active":true,"scope":"scope-a"}`), nil
		}),
	}

	if err := v.validate(context.Background(), "token"); err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if err := v.validate(context.Background(), "token"); err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("expected cache hit, got %d", got)
	}
	select {
	case err := <-errCh:
		t.Fatal(err)
	default:
	}

	v.hc = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{"active":true,"scope":"scope-b"}`), nil
		}),
	}
	if err := v.validate(context.Background(), "other"); err == nil {
		t.Fatalf("expected scope error")
	}

	var backoffHits int32
	v.hc = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&backoffHits, 1)
			return jsonResponse(http.StatusTooManyRequests, `{"error":"rate"}`), nil
		}),
	}
	if err := v.validate(context.Background(), "rate"); err == nil {
		t.Fatalf("expected 429 error")
	}
	if err := v.validate(context.Background(), "rate"); err == nil {
		t.Fatalf("expected backoff error")
	}
	if got := atomic.LoadInt32(&backoffHits); got != 1 {
		t.Fatalf("expected single call during backoff, got %d", got)
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

func tlsFixture(t *testing.T) types.TLSConfig {
	t.Helper()

	base := filepath.Join("..", "..", "..", "..", "example", "relay_example", "tls")
	cert := filepath.Join(base, "server.crt")
	key := filepath.Join(base, "server.key")
	ca := filepath.Join(base, "ca.crt")

	if _, err := os.Stat(cert); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(key); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(ca); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}

	return types.TLSConfig{
		UseTLS:                 true,
		CertFile:               cert,
		KeyFile:                key,
		CAFile:                 ca,
		MinTLSVersion:          tls.VersionTLS12,
		MaxTLSVersion:          tls.VersionTLS13,
	}
}
