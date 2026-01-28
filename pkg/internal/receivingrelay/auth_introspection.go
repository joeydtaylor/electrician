package receivingrelay

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/auth"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type tokenCacheEntry struct {
	active bool
	scope  string
	exp    time.Time
}

type cachingIntrospectionValidator struct {
	introspectionURL string
	authType         string
	clientID         string
	clientSecret     string
	bearerToken      string
	requiredScopes   []string

	hc  *http.Client
	mu  sync.Mutex
	m   map[string]tokenCacheEntry
	ttl time.Duration

	backoffUntil time.Time
	backoffStep  time.Duration
	maxBackoff   time.Duration
}

func newCachingIntrospectionValidator(o *relay.OAuth2Options) *cachingIntrospectionValidator {
	ttl := time.Duration(o.GetIntrospectionCacheSeconds()) * time.Second
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS13,
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true,
		},
	}
	return &cachingIntrospectionValidator{
		introspectionURL: strings.TrimRight(o.GetIntrospectionUrl(), "/"),
		authType:         strings.ToLower(o.GetIntrospectionAuthType()),
		clientID:         o.GetIntrospectionClientId(),
		clientSecret:     o.GetIntrospectionClientSecret(),
		bearerToken:      o.GetIntrospectionBearerToken(),
		requiredScopes:   append([]string(nil), o.GetRequiredScopes()...),
		hc:               &http.Client{Timeout: 8 * time.Second, Transport: tr},
		m:                make(map[string]tokenCacheEntry),
		ttl:              ttl,
		backoffStep:      250 * time.Millisecond,
		maxBackoff:       5 * time.Second,
	}
}

func (v *cachingIntrospectionValidator) hasAllScopes(granted string) bool {
	if len(v.requiredScopes) == 0 {
		return true
	}
	parts := strings.Fields(granted)
	set := make(map[string]struct{}, len(parts))
	for _, s := range parts {
		set[s] = struct{}{}
	}
	for _, need := range v.requiredScopes {
		if _, ok := set[need]; !ok {
			return false
		}
	}
	return true
}

type introspectResp struct {
	Active bool   `json:"active"`
	Scope  string `json:"scope"`
}

func (v *cachingIntrospectionValidator) validate(ctx context.Context, token string) error {
	now := time.Now()

	if until := v.backoffUntil; until.After(now) {
		return errors.New("auth server backoff in effect")
	}

	v.mu.Lock()
	if e, ok := v.m[token]; ok && e.exp.After(now) {
		v.mu.Unlock()
		if !e.active {
			return errors.New("token inactive")
		}
		if !v.hasAllScopes(e.scope) {
			return errors.New("insufficient scope")
		}
		return nil
	}
	v.mu.Unlock()

	form := url.Values{}
	form.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.introspectionURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	switch v.authType {
	case "basic":
		req.SetBasicAuth(v.clientID, v.clientSecret)
	case "bearer":
		if v.bearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+v.bearerToken)
		}
	case "none":
	default:
		if v.clientID != "" || v.clientSecret != "" {
			req.SetBasicAuth(v.clientID, v.clientSecret)
		}
	}

	resp, err := v.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		v.mu.Lock()
		if v.backoffStep < v.maxBackoff {
			v.backoffStep *= 2
			if v.backoffStep > v.maxBackoff {
				v.backoffStep = v.maxBackoff
			}
		}
		v.backoffUntil = now.Add(v.backoffStep)
		v.mu.Unlock()
		return errors.New("introspection 429")
	}
	v.mu.Lock()
	v.backoffStep = 250 * time.Millisecond
	v.backoffUntil = time.Time{}
	v.mu.Unlock()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New(resp.Status)
	}

	var ir introspectResp
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return err
	}

	v.mu.Lock()
	v.m[token] = tokenCacheEntry{
		active: ir.Active,
		scope:  ir.Scope,
		exp:    now.Add(v.ttl),
	}
	v.mu.Unlock()

	if !ir.Active {
		return errors.New("token inactive")
	}
	if !v.hasAllScopes(ir.Scope) {
		return errors.New("insufficient scope")
	}
	return nil
}

func (rr *ReceivingRelay[T]) ensureDefaultAuthValidator() {
	if rr.dynamicAuthValidator != nil {
		return
	}
	if rr.authOptions == nil || rr.authOptions.Mode != relay.AuthMode_AUTH_OAUTH2 {
		return
	}
	o := rr.authOptions.GetOauth2()
	if o == nil {
		return
	}

	var jwtValidator *auth.JWTValidator
	if o.GetAcceptJwt() && o.GetJwksUri() != "" {
		jwtValidator = auth.NewJWTValidator(o)
		if jwtValidator != nil {
			rr.logKV(
				types.InfoLevel,
				"Auth JWT validator installed",
				"event", "AuthJWT",
				"result", "SUCCESS",
				"issuer", o.GetIssuer(),
				"jwks_url", o.GetJwksUri(),
				"required_audience", o.GetRequiredAudience(),
				"required_scopes", o.GetRequiredScopes(),
			)
		}
	}

	var introspectionValidator *cachingIntrospectionValidator
	if o.GetAcceptIntrospection() && o.GetIntrospectionUrl() != "" {
		introspectionValidator = newCachingIntrospectionValidator(o)
		rr.logKV(
			types.InfoLevel,
			"Auth introspection validator installed",
			"event", "AuthIntrospection",
			"result", "SUCCESS",
			"introspection_url", o.GetIntrospectionUrl(),
			"auth_type", o.GetIntrospectionAuthType(),
			"required_scopes", o.GetRequiredScopes(),
		)
	}

	if jwtValidator == nil && introspectionValidator == nil {
		return
	}

	rr.dynamicAuthValidator = func(ctx context.Context, md map[string]string) error {
		var token string
		if v, ok := md["authorization"]; ok && strings.HasPrefix(strings.ToLower(v), "bearer ") {
			token = strings.TrimSpace(v[len("bearer "):])
		}
		if token == "" {
			return errors.New("missing bearer token")
		}
		if jwtValidator != nil && (auth.LooksLikeJWT(token) || introspectionValidator == nil) {
			return jwtValidator.Validate(ctx, token)
		}
		if introspectionValidator != nil {
			return introspectionValidator.validate(ctx, token)
		}
		return errors.New("missing bearer token")
	}
}
