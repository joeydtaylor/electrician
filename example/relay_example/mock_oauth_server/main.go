package main

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"log"
	"math/big"
	"net"
	"net/http"
	"strings"
	"time"
)

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int32  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type introspectResponse struct {
	Active bool   `json:"active"`
	Scope  string `json:"scope"`
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type server struct {
	issuer   string
	audience string
	kid      string

	clientID     string
	clientSecret string

	key *rsa.PrivateKey
	pub *rsa.PublicKey
}

const (
	addr           = "localhost:3000"
	issuer         = "auth-service"
	audience       = "your-api"
	clientID       = "steeze-local-cli"
	clientSecret   = "local-secret"
	keyID          = "dev-key"
	tokenTTL       = 3600
	certValidDays  = 30
	oauthTokenPath = "/api/auth/oauth/token"
	oauthJWKSPath  = "/api/auth/oauth/jwks.json"
	oauthIntPath   = "/api/auth/oauth/introspect"
	sessionToken   = "/api/auth/session/token"
)

func main() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generate key: %v", err)
	}

	s := &server{
		issuer:       issuer,
		audience:     audience,
		kid:          keyID,
		clientID:     clientID,
		clientSecret: clientSecret,
		key:          key,
		pub:          &key.PublicKey,
	}

	h := http.NewServeMux()
	h.HandleFunc(oauthTokenPath, s.handleToken)
	h.HandleFunc(oauthJWKSPath, s.handleJWKS)
	h.HandleFunc(oauthIntPath, s.handleIntrospect)
	h.HandleFunc(sessionToken, s.handleSessionToken)

	tlsCfg, err := buildTLSConfig()
	if err != nil {
		log.Fatalf("tls config: %v", err)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	server := &http.Server{
		Handler:      loggingMiddleware(h),
		TLSConfig:    tlsCfg,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Mock OAuth server up: https://%s", addr)
	log.Printf("Issuer: %s", issuer)
	log.Printf("JWKS:   https://%s%s", addr, oauthJWKSPath)
	log.Printf("Token:  https://%s%s", addr, oauthTokenPath)

	if err := server.Serve(tls.NewListener(listener, tlsCfg)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func (s *server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	id, secret, ok := r.BasicAuth()
	if !ok || id != s.clientID || secret != s.clientSecret {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "invalid client"})
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad form"})
		return
	}

	grant := r.FormValue("grant_type")
	if grant != "client_credentials" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "unsupported grant_type"})
		return
	}

	scope := r.FormValue("scope")
	if scope == "" {
		scope = "write:data"
	}

	tok, err := s.signToken(scope)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token sign failed"})
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: tok,
		TokenType:   "Bearer",
		ExpiresIn:   tokenTTL,
		Scope:       scope,
	})
}

func (s *server) handleSessionToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}
	var body struct {
		Scope string `json:"scope"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	scope := strings.TrimSpace(body.Scope)
	if scope == "" {
		scope = "write:data"
	}

	tok, err := s.signToken(scope)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "token sign failed"})
		return
	}

	writeJSON(w, http.StatusOK, tokenResponse{
		AccessToken: tok,
		TokenType:   "Bearer",
		ExpiresIn:   tokenTTL,
		Scope:       scope,
	})
}

func (s *server) handleJWKS(w http.ResponseWriter, r *http.Request) {
	jwks := jwksDocument{Keys: []jwkKey{s.publicJWK()}}
	writeJSON(w, http.StatusOK, jwks)
}

func (s *server) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
		return
	}

	if err := r.ParseForm(); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "bad form"})
		return
	}
	ok, scope := s.validateToken(r.FormValue("token"))
	writeJSON(w, http.StatusOK, introspectResponse{Active: ok, Scope: scope})
}

func (s *server) publicJWK() jwkKey {
	n := base64.RawURLEncoding.EncodeToString(s.pub.N.Bytes())
	e := base64.RawURLEncoding.EncodeToString(big.NewInt(int64(s.pub.E)).Bytes())
	return jwkKey{
		Kty: "RSA",
		Kid: s.kid,
		Use: "sig",
		Alg: "RS256",
		N:   n,
		E:   e,
	}
}

func (s *server) signToken(scope string) (string, error) {
	now := time.Now().Unix()
	claims := map[string]interface{}{
		"jti":   newJTI(),
		"sub":   s.clientID,
		"cid":   s.clientID,
		"org":   "local",
		"scope": scope,
		"iss":   s.issuer,
		"aud":   s.audience,
		"iat":   now,
		"exp":   now + tokenTTL,
	}

	header := map[string]interface{}{
		"alg": "RS256",
		"kid": s.kid,
		"typ": "JWT",
	}

	headJSON, _ := json.Marshal(header)
	claimsJSON, _ := json.Marshal(claims)

	payload := encodeJWTPart(headJSON) + "." + encodeJWTPart(claimsJSON)
	hash := sha256.Sum256([]byte(payload))
	sig, err := rsa.SignPKCS1v15(rand.Reader, s.key, cryptoHash(), hash[:])
	if err != nil {
		return "", err
	}

	return payload + "." + encodeJWTPart(sig), nil
}

func (s *server) validateToken(token string) (bool, string) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, ""
	}

	head, err := decodeJWTPart(parts[0])
	if err != nil {
		return false, ""
	}
	var hdr struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(head, &hdr); err != nil {
		return false, ""
	}
	if hdr.Alg != "RS256" {
		return false, ""
	}

	payload, err := decodeJWTPart(parts[1])
	if err != nil {
		return false, ""
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return false, ""
	}

	signingInput := parts[0] + "." + parts[1]
	sig, err := decodeJWTPart(parts[2])
	if err != nil {
		return false, ""
	}

	hash := sha256.Sum256([]byte(signingInput))
	if err := rsa.VerifyPKCS1v15(s.pub, cryptoHash(), hash[:], sig); err != nil {
		return false, ""
	}

	if iss, _ := claims["iss"].(string); iss != s.issuer {
		return false, ""
	}
	if aud, _ := claims["aud"].(string); aud != s.audience {
		return false, ""
	}

	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return false, ""
		}
	}

	scope, _ := claims["scope"].(string)
	return true, scope
}

func cryptoHash() crypto.Hash {
	return crypto.SHA256
}

func encodeJWTPart(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeJWTPart(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
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

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func buildTLSConfig() (*tls.Config, error) {
	certPEM, keyPEM, err := selfSignedCert("localhost")
	if err != nil {
		return nil, err
	}
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}}, nil
}

func selfSignedCert(host string) ([]byte, []byte, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return nil, nil, err
	}

	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			Organization: []string{"Electrician Mock Auth"},
		},
		NotBefore:             now.Add(-1 * time.Hour),
		NotAfter:              now.Add(time.Duration(certValidDays) * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{host},
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return certPEM, keyPEM, nil
}
