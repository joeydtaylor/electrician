package testoauth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"
)

// Config mirrors the mock OAuth defaults used by the relay examples.
type Config struct {
	Issuer       string
	Audience     string
	Scope        string
	ClientID     string
	ClientSecret string
	Kid          string
	TokenTTL     time.Duration
}

// Server is a minimal OAuth server for tests (token + JWKS).
type Server struct {
	URL      string
	JWKSURL  string
	TokenURL string
	Close    func()
}

// NewTLSServer starts a test OAuth server over TLS.
func NewTLSServer(cfg Config) (*Server, error) {
	applyDefaults(&cfg)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	jwk := buildJWK(&key.PublicKey, cfg.Kid)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/oauth/jwks.json", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"keys": []map[string]string{jwk},
		})
	})
	mux.HandleFunc("/api/auth/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != cfg.ClientID || pass != cfg.ClientSecret {
			http.Error(w, "invalid client", http.StatusUnauthorized)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		if grant := strings.TrimSpace(r.FormValue("grant_type")); grant != "client_credentials" {
			http.Error(w, "unsupported grant_type", http.StatusBadRequest)
			return
		}
		scope := strings.TrimSpace(r.FormValue("scope"))
		if scope == "" {
			scope = cfg.Scope
		}
		tok, err := signJWT(cfg, key, scope)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"access_token": tok,
			"token_type":   "Bearer",
			"expires_in":   int64(cfg.TokenTTL.Seconds()),
			"scope":        scope,
		})
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	srv := httptest.NewUnstartedServer(mux)
	srv.Listener = listener
	srv.StartTLS()

	return &Server{
		URL:      srv.URL,
		JWKSURL:  srv.URL + "/api/auth/oauth/jwks.json",
		TokenURL: srv.URL + "/api/auth/oauth/token",
		Close:    srv.Close,
	}, nil
}

func applyDefaults(cfg *Config) {
	if cfg.Issuer == "" {
		cfg.Issuer = "auth-service"
	}
	if cfg.Audience == "" {
		cfg.Audience = "your-api"
	}
	if cfg.Scope == "" {
		cfg.Scope = "write:data"
	}
	if cfg.ClientID == "" {
		cfg.ClientID = "steeze-local-cli"
	}
	if cfg.ClientSecret == "" {
		cfg.ClientSecret = "local-secret"
	}
	if cfg.Kid == "" {
		cfg.Kid = "mock-key-1"
	}
	if cfg.TokenTTL == 0 {
		cfg.TokenTTL = 5 * time.Minute
	}
}

func buildJWK(pub *rsa.PublicKey, kid string) map[string]string {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())
	return map[string]string{
		"kty": "RSA",
		"kid": kid,
		"use": "sig",
		"alg": "RS256",
		"n":   n,
		"e":   e,
	}
}

func signJWT(cfg Config, key *rsa.PrivateKey, scope string) (string, error) {
	now := time.Now().Unix()
	header := map[string]interface{}{
		"alg": "RS256",
		"kid": cfg.Kid,
		"typ": "JWT",
	}
	claims := map[string]interface{}{
		"iss":   cfg.Issuer,
		"sub":   cfg.ClientID,
		"aud":   cfg.Audience,
		"scope": scope,
		"iat":   now,
		"exp":   now + int64(cfg.TokenTTL.Seconds()),
		"jti":   newJTI(),
	}
	headJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	payload := encodeJWTPart(headJSON) + "." + encodeJWTPart(claimsJSON)
	hash := sha256.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	return payload + "." + encodeJWTPart(sig), nil
}

func encodeJWTPart(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func newJTI() string {
	buf := make([]byte, 16)
	_, _ = rand.Read(buf)
	return base64.RawURLEncoding.EncodeToString(buf)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
