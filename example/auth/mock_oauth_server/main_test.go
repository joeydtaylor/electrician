package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

type testLogger struct{}

func (testLogger) Info(string, ...interface{})  {}
func (testLogger) Debug(string, ...interface{}) {}

func newTestServer(t *testing.T) *server {
	t.Helper()
	priv, pub, err := generateRSAKey()
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	return &server{
		cfg: config{
			issuer:            "auth-service",
			audience:          "your-api",
			scope:             "write:data",
			tokenTTL:          5 * time.Minute,
			kid:               "test-kid",
			subject:           "user-local",
			staticToken:       "token-123",
			oauthClientID:     "steeze-local-cli",
			oauthClientSecret: "local-secret",
			introspectAuth:    "basic",
			introspectID:      "steeze-local-cli",
			introspectSecret:  "local-secret",
		},
		privKey: priv,
		pubKey:  pub,
		logger:  testLogger{},
	}
}

func TestHandleOAuthTokenSuccess(t *testing.T) {
	s := newTestServer(t)
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	form.Set("scope", "write:data")
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.cfg.oauthClientID, s.cfg.oauthClientSecret)
	rr := httptest.NewRecorder()

	s.handleOAuthToken(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	defer res.Body.Close()

	var tr tokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if tr.AccessToken == "" {
		t.Fatalf("expected access_token")
	}
	if tr.TokenType != "Bearer" {
		t.Fatalf("unexpected token_type: %s", tr.TokenType)
	}
	if tr.Scope != "write:data" {
		t.Fatalf("unexpected scope: %s", tr.Scope)
	}
	if tr.ExpiresIn <= 0 {
		t.Fatalf("unexpected expires_in: %d", tr.ExpiresIn)
	}

	claims, err := s.verifyJWT(tr.AccessToken)
	if err != nil {
		t.Fatalf("verify jwt: %v", err)
	}
	if iss, _ := claims["iss"].(string); iss != s.cfg.issuer {
		t.Fatalf("unexpected iss: %s", iss)
	}
	if aud, _ := claims["aud"].(string); aud != s.cfg.audience {
		t.Fatalf("unexpected aud: %s", aud)
	}
}

func TestHandleOAuthTokenUnauthorized(t *testing.T) {
	s := newTestServer(t)
	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("bad", "creds")
	rr := httptest.NewRecorder()

	s.handleOAuthToken(rr, req)

	if rr.Result().StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Result().StatusCode)
	}
}

func TestHandleIntrospectJWT(t *testing.T) {
	s := newTestServer(t)
	token, err := s.signJWT("write:data")
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	form := url.Values{}
	form.Set("token", token)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/introspect", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.cfg.introspectID, s.cfg.introspectSecret)
	rr := httptest.NewRecorder()

	s.handleIntrospect(rr, req)

	res := rr.Result()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", res.StatusCode)
	}
	defer res.Body.Close()

	var ir introspectResponse
	if err := json.NewDecoder(res.Body).Decode(&ir); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !ir.Active {
		t.Fatalf("expected active=true")
	}
	if ir.Scope != "write:data" {
		t.Fatalf("unexpected scope: %s", ir.Scope)
	}
}

func TestHandleIntrospectStaticToken(t *testing.T) {
	s := newTestServer(t)
	form := url.Values{}
	form.Set("token", s.cfg.staticToken)
	req := httptest.NewRequest(http.MethodPost, "/api/auth/oauth/introspect", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.cfg.introspectID, s.cfg.introspectSecret)
	rr := httptest.NewRecorder()

	s.handleIntrospect(rr, req)

	var ir introspectResponse
	if err := json.NewDecoder(rr.Result().Body).Decode(&ir); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !ir.Active {
		t.Fatalf("expected active=true")
	}
}

func TestHandleJWKS(t *testing.T) {
	s := newTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/auth/oauth/jwks.json", nil)
	rr := httptest.NewRecorder()

	s.handleJWKS(rr, req)

	if rr.Result().StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", rr.Result().StatusCode)
	}
	var doc struct {
		Keys []map[string]interface{} `json:"keys"`
	}
	if err := json.NewDecoder(rr.Result().Body).Decode(&doc); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(doc.Keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(doc.Keys))
	}
	if kid, _ := doc.Keys[0]["kid"].(string); kid != s.cfg.kid {
		t.Fatalf("unexpected kid: %v", kid)
	}
}
