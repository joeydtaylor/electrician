package main

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/joeydtaylor/electrician/pkg/builder"
)

type config struct {
	addr              string
	useTLS            bool
	tlsCert           string
	tlsKey            string
	issuer            string
	audience          string
	scope             string
	tokenTTL          time.Duration
	kid               string
	subject           string
	staticToken       string
	oauthClientID     string
	oauthClientSecret string
	introspectAuth    string
	introspectID      string
	introspectSecret  string
	introspectBearer  string
}

type server struct {
	cfg     config
	privKey *rsa.PrivateKey
	pubKey  *rsa.PublicKey
	logger  logSink
}

type logSink interface {
	Info(msg string, keysAndValues ...interface{})
	Debug(msg string, keysAndValues ...interface{})
}

type tokenRequest struct {
	Scope string `json:"scope"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type introspectResponse struct {
	Active bool   `json:"active"`
	Scope  string `json:"scope"`
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envOrDuration(k string, def time.Duration) time.Duration {
	if v := os.Getenv(k); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func envOrBool(k string, def bool) bool {
	if v := os.Getenv(k); v != "" {
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "1", "true", "yes", "y", "on":
			return true
		case "0", "false", "no", "n", "off":
			return false
		}
	}
	return def
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func findRepoRoot() (string, bool) {
	dir, err := os.Getwd()
	if err != nil {
		return "", false
	}
	for {
		if fileExists(filepath.Join(dir, "go.mod")) {
			return dir, true
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", false
		}
		dir = parent
	}
}

func defaultTLSAssetPath(filename string) string {
	candidates := []string{
		filepath.Join("example", "relay_example", "tls", filename),
		filepath.Join("..", "relay_example", "tls", filename),
		filepath.Join("..", "..", "relay_example", "tls", filename),
	}
	if root, ok := findRepoRoot(); ok {
		candidates = append([]string{filepath.Join(root, "example", "relay_example", "tls", filename)}, candidates...)
	}
	for _, candidate := range candidates {
		if fileExists(candidate) {
			return candidate
		}
	}
	return filepath.Join("example", "relay_example", "tls", filename)
}

func loadConfig() config {
	useTLS := !envOrBool("OAUTH_TLS_DISABLE", false)
	tlsCert := envOr("TLS_CERT", "")
	if tlsCert == "" {
		tlsCert = defaultTLSAssetPath("server.crt")
	}
	tlsKey := envOr("TLS_KEY", "")
	if tlsKey == "" {
		tlsKey = defaultTLSAssetPath("server.key")
	}
	return config{
		addr:              envOr("OAUTH_ADDR", "localhost:3000"),
		useTLS:            useTLS,
		tlsCert:           tlsCert,
		tlsKey:            tlsKey,
		issuer:            envOr("OAUTH_ISSUER_BASE", "auth-service"),
		audience:          envOr("OAUTH_AUDIENCE", "your-api"),
		scope:             envOr("OAUTH_SCOPE", "write:data"),
		tokenTTL:          envOrDuration("OAUTH_TOKEN_TTL", 300*time.Second),
		kid:               envOr("OAUTH_KID", "mock-key-1"),
		subject:           envOr("OAUTH_SUBJECT", "user-local"),
		staticToken:       envOr("OAUTH_STATIC_TOKEN", "token-123"),
		oauthClientID:     envOr("OAUTH_CLIENT_ID", "steeze-local-cli"),
		oauthClientSecret: envOr("OAUTH_CLIENT_SECRET", "local-secret"),
		introspectAuth:    strings.ToLower(envOr("INTROSPECT_AUTH", "basic")),
		introspectID:      envOr("INTROSPECT_CLIENT_ID", "steeze-local-cli"),
		introspectSecret:  envOr("INTROSPECT_CLIENT_SECRET", "local-secret"),
		introspectBearer:  envOr("INTROSPECT_BEARER_TOKEN", ""),
	}
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := loadConfig()
	logLevel := envOr("LOG_LEVEL", "info")
	logger := builder.NewLogger(
		builder.LoggerWithDevelopment(true),
		builder.LoggerWithLevel(logLevel),
	)

	priv, pub, err := generateRSAKey()
	if err != nil {
		log.Fatalf("rsa keygen: %v", err)
	}

	s := &server{
		cfg:     cfg,
		privKey: priv,
		pubKey:  pub,
		logger:  logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/auth/oauth/jwks.json", s.handleJWKS)
	mux.HandleFunc("/api/auth/oauth/token", s.handleOAuthToken)
	mux.HandleFunc("/api/auth/session/token", s.handleSessionToken)
	mux.HandleFunc("/api/auth/oauth/introspect", s.handleIntrospect)
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := withCORS(mux)
	srv := &http.Server{
		Addr:    cfg.addr,
		Handler: handler,
	}

	go func() {
		<-ctx.Done()
		ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctxShutdown)
	}()

	logger.Info("Mock OAuth server starting",
		"event", "Start",
		"result", "SUCCESS",
		"address", cfg.addr,
		"use_tls", cfg.useTLS,
		"issuer", cfg.issuer,
		"audience", cfg.audience,
		"scope", cfg.scope,
		"token_ttl", cfg.tokenTTL.String(),
		"introspect_auth", cfg.introspectAuth,
	)

	if cfg.useTLS {
		srv.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
			MaxVersion: tls.VersionTLS13,
		}
		if err := srv.ListenAndServeTLS(cfg.tlsCert, cfg.tlsKey); err != nil && err != http.ErrServerClosed {
			log.Fatalf("https server: %v", err)
		}
		return
	}

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server: %v", err)
	}
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "authorization, content-type")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *server) handleJWKS(w http.ResponseWriter, _ *http.Request) {
	jwks := map[string]interface{}{
		"keys": []map[string]interface{}{
			buildJWK(s.pubKey, s.cfg.kid),
		},
	}
	writeJSON(w, jwks)
}

func (s *server) handleSessionToken(w http.ResponseWriter, r *http.Request) {
	scope := s.cfg.scope
	if r.Method == http.MethodPost {
		if req := readTokenRequest(r); req.Scope != "" {
			scope = req.Scope
		}
	}
	token, err := s.signJWT(scope)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}
	s.logger.Debug("Issued session token",
		"event", "TokenIssued",
		"result", "SUCCESS",
		"scope", scope,
		"issuer", s.cfg.issuer,
		"audience", s.cfg.audience,
		"token", token,
	)
	resp := tokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.cfg.tokenTTL.Seconds()),
		Scope:       scope,
	}
	writeJSON(w, resp)
}

func (s *server) handleOAuthToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	id, secret, ok := r.BasicAuth()
	if !ok || id != s.cfg.oauthClientID || secret != s.cfg.oauthClientSecret {
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
		scope = s.cfg.scope
	}
	token, err := s.signJWT(scope)
	if err != nil {
		http.Error(w, "token error", http.StatusInternalServerError)
		return
	}
	resp := tokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.cfg.tokenTTL.Seconds()),
		Scope:       scope,
	}
	writeJSON(w, resp)
}

func (s *server) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	if err := s.authorizeIntrospection(r); err != nil {
		s.logger.Debug("Introspection auth failed",
			"event", "Introspect",
			"result", "FAILURE",
			"error", err,
		)
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	token := readIntrospectionToken(r)
	if token == "" {
		s.logger.Debug("Introspection missing token",
			"event", "Introspect",
			"result", "FAILURE",
		)
		writeJSON(w, introspectResponse{Active: false})
		return
	}
	if token == s.cfg.staticToken {
		s.logger.Debug("Introspection static token",
			"event", "Introspect",
			"result", "SUCCESS",
			"scope", s.cfg.scope,
			"token", token,
		)
		writeJSON(w, introspectResponse{Active: true, Scope: s.cfg.scope})
		return
	}
	claims, err := s.verifyJWT(token)
	if err != nil {
		s.logger.Debug("Introspection JWT invalid",
			"event", "Introspect",
			"result", "FAILURE",
			"error", err,
			"token", token,
		)
		writeJSON(w, introspectResponse{Active: false})
		return
	}
	scope := s.cfg.scope
	if v, ok := claims["scope"].(string); ok && strings.TrimSpace(v) != "" {
		scope = v
	}
	s.logger.Debug("Introspection JWT valid",
		"event", "Introspect",
		"result", "SUCCESS",
		"scope", scope,
		"token", token,
	)
	writeJSON(w, introspectResponse{Active: true, Scope: scope})
}

func (s *server) authorizeIntrospection(r *http.Request) error {
	switch s.cfg.introspectAuth {
	case "none":
		return nil
	case "basic":
		user, pass, ok := r.BasicAuth()
		if !ok || user != s.cfg.introspectID || pass != s.cfg.introspectSecret {
			return fmt.Errorf("basic auth failed")
		}
		return nil
	case "bearer":
		want := s.cfg.introspectBearer
		if want == "" {
			return fmt.Errorf("missing introspect bearer token")
		}
		if got := bearerFromHeader(r.Header.Get("Authorization")); got != want {
			return fmt.Errorf("bearer auth failed")
		}
		return nil
	default:
		if s.cfg.introspectID == "" && s.cfg.introspectSecret == "" && s.cfg.introspectBearer == "" {
			return nil
		}
		user, pass, ok := r.BasicAuth()
		if ok && user == s.cfg.introspectID && pass == s.cfg.introspectSecret {
			return nil
		}
		if s.cfg.introspectBearer != "" && bearerFromHeader(r.Header.Get("Authorization")) == s.cfg.introspectBearer {
			return nil
		}
		return fmt.Errorf("auth failed")
	}
}

func readTokenRequest(r *http.Request) tokenRequest {
	var req tokenRequest
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		_ = json.NewDecoder(r.Body).Decode(&req)
		return req
	}
	_ = r.ParseForm()
	if scope := strings.TrimSpace(r.FormValue("scope")); scope != "" {
		req.Scope = scope
	}
	return req
}

func readIntrospectionToken(r *http.Request) string {
	if strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		var payload struct {
			Token string `json:"token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&payload)
		return strings.TrimSpace(payload.Token)
	}
	_ = r.ParseForm()
	return strings.TrimSpace(r.FormValue("token"))
}

func bearerFromHeader(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 {
		return ""
	}
	if strings.ToLower(strings.TrimSpace(parts[0])) != "bearer" {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func generateRSAKey() (*rsa.PrivateKey, *rsa.PublicKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}
	return priv, &priv.PublicKey, nil
}

func buildJWK(pub *rsa.PublicKey, kid string) map[string]interface{} {
	n := base64.RawURLEncoding.EncodeToString(pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes())
	return map[string]interface{}{
		"kty": "RSA",
		"kid": kid,
		"use": "sig",
		"alg": "RS256",
		"n":   n,
		"e":   e,
	}
}

func (s *server) signJWT(scope string) (string, error) {
	now := time.Now()
	claims := map[string]interface{}{
		"iss":   s.cfg.issuer,
		"sub":   s.cfg.subject,
		"aud":   s.cfg.audience,
		"scope": scope,
		"iat":   now.Unix(),
		"exp":   now.Add(s.cfg.tokenTTL).Unix(),
		"jti":   randomTokenID(),
	}
	header := map[string]interface{}{
		"alg": "RS256",
		"kid": s.cfg.kid,
		"typ": "JWT",
	}
	return signRS256(header, claims, s.privKey)
}

func (s *server) verifyJWT(token string) (map[string]interface{}, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid jwt format")
	}
	headBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header struct {
		Alg string `json:"alg"`
	}
	if err := json.Unmarshal(headBytes, &header); err != nil {
		return nil, err
	}
	if header.Alg != "RS256" {
		return nil, fmt.Errorf("unsupported alg %s", header.Alg)
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	signingInput := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(s.pubKey, crypto.SHA256, hash[:], sig); err != nil {
		return nil, err
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(payloadBytes))
	dec.UseNumber()
	if err := dec.Decode(&claims); err != nil {
		return nil, err
	}
	if exp, ok := claims["exp"].(json.Number); ok {
		if expInt, err := exp.Int64(); err == nil {
			if time.Now().Unix() > expInt {
				return nil, fmt.Errorf("token expired")
			}
		}
	}
	return claims, nil
}

func signRS256(header, claims map[string]interface{}, key *rsa.PrivateKey) (string, error) {
	headBytes, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	claimBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	headB64 := base64.RawURLEncoding.EncodeToString(headBytes)
	claimB64 := base64.RawURLEncoding.EncodeToString(claimBytes)
	signingInput := headB64 + "." + claimB64
	hash := sha256.Sum256([]byte(signingInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, hash[:])
	if err != nil {
		return "", err
	}
	sigB64 := base64.RawURLEncoding.EncodeToString(sig)
	return signingInput + "." + sigB64, nil
}

func randomTokenID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("tok-%d", time.Now().UnixNano())
	}
	return fmt.Sprintf("%x", b)
}
