package auth

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

var jwtValidAlgs = map[string]struct{}{
	"RS256": {}, "RS384": {}, "RS512": {},
	"PS256": {}, "PS384": {}, "PS512": {},
	"ES256": {}, "ES384": {}, "ES512": {},
}

type jwksDocument struct {
	Keys []jwkKey `json:"keys"`
}

type jwkKey struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	Alg string `json:"alg"`

	N string `json:"n"`
	E string `json:"e"`

	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type JWTValidator struct {
	issuer           string
	jwksURL          string
	requiredAudience []string
	requiredScopes   []string

	ttl time.Duration
	hc  *http.Client

	mu      sync.RWMutex
	keys    map[string]interface{}
	allKeys []interface{}
	expires time.Time
}

func NewJWTValidator(o *relay.OAuth2Options) *JWTValidator {
	if o == nil {
		return nil
	}
	jwksURL := strings.TrimSpace(o.GetJwksUri())
	if jwksURL == "" {
		return nil
	}
	ttl := time.Duration(o.GetJwksCacheSeconds()) * time.Second
	if ttl <= 0 {
		ttl = 300 * time.Second
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS13,
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true,
		},
	}
	return &JWTValidator{
		issuer:           o.GetIssuer(),
		jwksURL:          strings.TrimRight(jwksURL, "/"),
		requiredAudience: cloneStrings(o.GetRequiredAudience()),
		requiredScopes:   cloneStrings(o.GetRequiredScopes()),
		ttl:              ttl,
		hc:               &http.Client{Timeout: 8 * time.Second, Transport: tr},
		keys:             make(map[string]interface{}),
	}
}

func LooksLikeJWT(token string) bool {
	return strings.Count(token, ".") == 2
}

func (v *JWTValidator) Validate(ctx context.Context, token string) error {
	if v == nil {
		return errors.New("jwt validator not configured")
	}
	if token == "" {
		return errors.New("missing bearer token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return errors.New("invalid jwt format")
	}

	headBytes, err := decodeJWTPart(parts[0])
	if err != nil {
		return fmt.Errorf("jwt header decode: %w", err)
	}
	payloadBytes, err := decodeJWTPart(parts[1])
	if err != nil {
		return fmt.Errorf("jwt payload decode: %w", err)
	}
	sig, err := decodeJWTPart(parts[2])
	if err != nil {
		return fmt.Errorf("jwt signature decode: %w", err)
	}

	var header jwtHeader
	if err := json.Unmarshal(headBytes, &header); err != nil {
		return fmt.Errorf("jwt header parse: %w", err)
	}
	if header.Alg == "" {
		return errors.New("jwt missing alg")
	}
	if _, ok := jwtValidAlgs[header.Alg]; !ok {
		return fmt.Errorf("unsupported jwt alg %q", header.Alg)
	}

	key, err := v.lookupKey(ctx, header.Kid)
	if err != nil {
		return err
	}

	signingInput := parts[0] + "." + parts[1]
	if err := verifySignature(header.Alg, signingInput, sig, key); err != nil {
		return err
	}

	claims, err := decodeClaims(payloadBytes)
	if err != nil {
		return err
	}

	if err := validateStandardClaims(claims, v.issuer, v.requiredAudience); err != nil {
		return err
	}
	if !hasAllScopes(v.requiredScopes, claims) {
		return errors.New("insufficient scope")
	}

	return nil
}

func decodeClaims(payload []byte) (map[string]interface{}, error) {
	var claims map[string]interface{}
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	if err := dec.Decode(&claims); err != nil {
		return nil, fmt.Errorf("jwt claims parse: %w", err)
	}
	return claims, nil
}

func validateStandardClaims(claims map[string]interface{}, issuer string, audiences []string) error {
	now := time.Now().Unix()
	const skewSeconds int64 = 5

	if issuer != "" {
		iss, _ := claims["iss"].(string)
		if iss == "" || iss != issuer {
			return errors.New("issuer mismatch")
		}
	}

	if len(audiences) > 0 {
		got := extractAudience(claims)
		if !audienceMatch(got, audiences) {
			return errors.New("audience mismatch")
		}
	}

	if exp, ok := claimInt64(claims, "exp"); ok {
		if now > exp+skewSeconds {
			return errors.New("token expired")
		}
	}
	if nbf, ok := claimInt64(claims, "nbf"); ok {
		if now+skewSeconds < nbf {
			return errors.New("token not yet valid")
		}
	}
	return nil
}

func extractAudience(claims map[string]interface{}) []string {
	v, ok := claims["aud"]
	if !ok {
		return nil
	}
	switch t := v.(type) {
	case string:
		if t == "" {
			return nil
		}
		return []string{t}
	case []string:
		return cloneStrings(t)
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					out = append(out, trimmed)
				}
			}
		}
		return out
	default:
		return nil
	}
}

func audienceMatch(got, required []string) bool {
	if len(required) == 0 {
		return true
	}
	if len(got) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(got))
	for _, a := range got {
		set[a] = struct{}{}
	}
	for _, need := range required {
		if _, ok := set[need]; ok {
			return true
		}
	}
	return false
}

func claimInt64(claims map[string]interface{}, key string) (int64, bool) {
	v, ok := claims[key]
	if !ok {
		return 0, false
	}
	switch t := v.(type) {
	case json.Number:
		i, err := t.Int64()
		return i, err == nil
	case float64:
		return int64(t), true
	case int64:
		return t, true
	case int:
		return int64(t), true
	case string:
		if t == "" {
			return 0, false
		}
		var n json.Number = json.Number(t)
		i, err := n.Int64()
		return i, err == nil
	default:
		return 0, false
	}
}

func (v *JWTValidator) lookupKey(ctx context.Context, kid string) (interface{}, error) {
	keys, all, err := v.getKeys(ctx, false)
	if err != nil {
		return nil, err
	}
	if kid != "" {
		if key, ok := keys[kid]; ok {
			return key, nil
		}
	} else if len(all) == 1 {
		return all[0], nil
	}

	keys, all, err = v.getKeys(ctx, true)
	if err != nil {
		return nil, err
	}
	if kid != "" {
		if key, ok := keys[kid]; ok {
			return key, nil
		}
		return nil, fmt.Errorf("jwks key not found for kid %q", kid)
	}
	if len(all) == 1 {
		return all[0], nil
	}
	return nil, fmt.Errorf("jwt kid missing and jwks has %d keys", len(all))
}

func (v *JWTValidator) getKeys(ctx context.Context, force bool) (map[string]interface{}, []interface{}, error) {
	now := time.Now()
	v.mu.RLock()
	if !force && len(v.allKeys) > 0 && now.Before(v.expires) {
		keys := v.keys
		all := v.allKeys
		v.mu.RUnlock()
		return keys, all, nil
	}
	v.mu.RUnlock()

	keys, all, err := v.fetchJWKS(ctx)
	if err != nil {
		return nil, nil, err
	}

	v.mu.Lock()
	v.keys = keys
	v.allKeys = all
	v.expires = now.Add(v.ttl)
	v.mu.Unlock()
	return keys, all, nil
}

func (v *JWTValidator) fetchJWKS(ctx context.Context) (map[string]interface{}, []interface{}, error) {
	if v.hc == nil {
		v.hc = &http.Client{Timeout: 8 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, v.jwksURL, nil)
	if err != nil {
		return nil, nil, err
	}
	resp, err := v.hc.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, nil, fmt.Errorf("jwks status %s", resp.Status)
	}

	var doc jwksDocument
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, nil, err
	}

	keys := make(map[string]interface{}, len(doc.Keys))
	all := make([]interface{}, 0, len(doc.Keys))
	for _, k := range doc.Keys {
		key, err := parseJWK(k)
		if err != nil {
			continue
		}
		all = append(all, key)
		if k.Kid != "" {
			keys[k.Kid] = key
		}
	}
	if len(all) == 0 {
		return nil, nil, errors.New("jwks: no usable keys")
	}
	return keys, all, nil
}

func parseJWK(k jwkKey) (interface{}, error) {
	switch strings.ToUpper(k.Kty) {
	case "RSA":
		return parseRSAPublicKey(k.N, k.E)
	case "EC":
		return parseECPublicKey(k.Crv, k.X, k.Y)
	default:
		return nil, fmt.Errorf("unsupported jwk kty %q", k.Kty)
	}
}

func parseRSAPublicKey(nStr, eStr string) (*rsa.PublicKey, error) {
	if nStr == "" || eStr == "" {
		return nil, errors.New("rsa jwk missing n/e")
	}
	nb, err := base64.RawURLEncoding.DecodeString(nStr)
	if err != nil {
		return nil, err
	}
	eb, err := base64.RawURLEncoding.DecodeString(eStr)
	if err != nil {
		return nil, err
	}
	e := 0
	for _, b := range eb {
		e = (e << 8) + int(b)
	}
	if e == 0 {
		return nil, errors.New("rsa jwk invalid exponent")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nb), E: e}, nil
}

func parseECPublicKey(crv, xStr, yStr string) (*ecdsa.PublicKey, error) {
	if crv == "" || xStr == "" || yStr == "" {
		return nil, errors.New("ec jwk missing crv/x/y")
	}
	var curve elliptic.Curve
	switch crv {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported ec curve %q", crv)
	}
	xb, err := base64.RawURLEncoding.DecodeString(xStr)
	if err != nil {
		return nil, err
	}
	yb, err := base64.RawURLEncoding.DecodeString(yStr)
	if err != nil {
		return nil, err
	}
	x := new(big.Int).SetBytes(xb)
	y := new(big.Int).SetBytes(yb)
	if !curve.IsOnCurve(x, y) {
		return nil, errors.New("ec jwk point not on curve")
	}
	return &ecdsa.PublicKey{Curve: curve, X: x, Y: y}, nil
}

func verifySignature(alg, signingInput string, sig []byte, key interface{}) error {
	switch alg {
	case "RS256":
		return verifyRSAPKCS1v15(crypto.SHA256, signingInput, sig, key)
	case "RS384":
		return verifyRSAPKCS1v15(crypto.SHA384, signingInput, sig, key)
	case "RS512":
		return verifyRSAPKCS1v15(crypto.SHA512, signingInput, sig, key)
	case "PS256":
		return verifyRSAPSS(crypto.SHA256, signingInput, sig, key)
	case "PS384":
		return verifyRSAPSS(crypto.SHA384, signingInput, sig, key)
	case "PS512":
		return verifyRSAPSS(crypto.SHA512, signingInput, sig, key)
	case "ES256":
		return verifyECDSA(crypto.SHA256, signingInput, sig, key)
	case "ES384":
		return verifyECDSA(crypto.SHA384, signingInput, sig, key)
	case "ES512":
		return verifyECDSA(crypto.SHA512, signingInput, sig, key)
	default:
		return fmt.Errorf("unsupported jwt alg %q", alg)
	}
}

func verifyRSAPKCS1v15(hash crypto.Hash, signingInput string, sig []byte, key interface{}) error {
	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("rsa public key required")
	}
	hashed, err := hashBytes(hash, signingInput)
	if err != nil {
		return err
	}
	return rsa.VerifyPKCS1v15(pub, hash, hashed, sig)
}

func verifyRSAPSS(hash crypto.Hash, signingInput string, sig []byte, key interface{}) error {
	pub, ok := key.(*rsa.PublicKey)
	if !ok {
		return errors.New("rsa public key required")
	}
	hashed, err := hashBytes(hash, signingInput)
	if err != nil {
		return err
	}
	return rsa.VerifyPSS(pub, hash, hashed, sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: hash})
}

func verifyECDSA(hash crypto.Hash, signingInput string, sig []byte, key interface{}) error {
	pub, ok := key.(*ecdsa.PublicKey)
	if !ok {
		return errors.New("ecdsa public key required")
	}
	hashed, err := hashBytes(hash, signingInput)
	if err != nil {
		return err
	}
	keyBytes := (pub.Curve.Params().BitSize + 7) / 8
	r, s, err := parseECDSASignature(sig, keyBytes)
	if err != nil {
		return err
	}
	if !ecdsa.Verify(pub, hashed, r, s) {
		return errors.New("ecdsa signature invalid")
	}
	return nil
}

func parseECDSASignature(sig []byte, keyBytes int) (*big.Int, *big.Int, error) {
	if len(sig) != 2*keyBytes {
		return nil, nil, fmt.Errorf("invalid ecdsa signature length %d", len(sig))
	}
	r := new(big.Int).SetBytes(sig[:keyBytes])
	s := new(big.Int).SetBytes(sig[keyBytes:])
	return r, s, nil
}

func hashBytes(hash crypto.Hash, input string) ([]byte, error) {
	switch hash {
	case crypto.SHA256:
		sum := sha256.Sum256([]byte(input))
		return sum[:], nil
	case crypto.SHA384:
		sum := sha512.Sum384([]byte(input))
		return sum[:], nil
	case crypto.SHA512:
		sum := sha512.Sum512([]byte(input))
		return sum[:], nil
	default:
		if !hash.Available() {
			return nil, errors.New("hash unavailable")
		}
		h := hash.New()
		_, _ = h.Write([]byte(input))
		return h.Sum(nil), nil
	}
}

func decodeJWTPart(part string) ([]byte, error) {
	if part == "" {
		return nil, errors.New("empty jwt part")
	}
	return base64.RawURLEncoding.DecodeString(part)
}

func hasAllScopes(required []string, claims map[string]interface{}) bool {
	if len(required) == 0 {
		return true
	}
	granted := extractScopes(claims)
	if len(granted) == 0 {
		return false
	}
	set := make(map[string]struct{}, len(granted))
	for _, s := range granted {
		set[s] = struct{}{}
	}
	for _, req := range required {
		if _, ok := set[req]; !ok {
			return false
		}
	}
	return true
}

func extractScopes(claims map[string]interface{}) []string {
	for _, key := range []string{"scope", "scp", "scopes"} {
		if v, ok := claims[key]; ok {
			if scopes := parseScopeValue(v); len(scopes) > 0 {
				return scopes
			}
		}
	}
	return nil
}

func parseScopeValue(v interface{}) []string {
	switch t := v.(type) {
	case string:
		return strings.Fields(t)
	case []string:
		return cloneStrings(t)
	case []interface{}:
		out := make([]string, 0, len(t))
		for _, item := range t {
			if s, ok := item.(string); ok {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					out = append(out, trimmed)
				}
			}
		}
		return out
	default:
		return nil
	}
}

func cloneStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}
