package auth

import (
	"bytes"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

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

func rsaPublicKey(key interface{}) (*rsa.PublicKey, bool) {
	pub, ok := key.(*rsa.PublicKey)
	return pub, ok
}
