package quicrelay

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
)

type tokenCacheEntry struct {
	active bool
	scope  string
	exp    time.Time
}

type cachingIntrospectionValidator struct {
	introspectionURL string
	requiredScopes   []string
	bearerToken      string
	ttl              time.Duration

	mu sync.Mutex
	m  map[string]tokenCacheEntry
}

type introspectResp struct {
	Active bool   `json:"active"`
	Scope  string `json:"scope"`
}

func newCachingIntrospectionValidator(o *relay.OAuth2Options) *cachingIntrospectionValidator {
	ttl := time.Duration(o.GetIntrospectionCacheSeconds()) * time.Second
	if ttl == 0 {
		ttl = 30 * time.Second
	}
	return &cachingIntrospectionValidator{
		introspectionURL: strings.TrimRight(o.GetIntrospectionUrl(), "/"),
		requiredScopes:   o.GetRequiredScopes(),
		bearerToken:      o.GetIntrospectionBearerToken(),
		ttl:              ttl,
		m:                make(map[string]tokenCacheEntry),
	}
}

func (v *cachingIntrospectionValidator) hasAllScopes(granted string) bool {
	if len(v.requiredScopes) == 0 {
		return true
	}
	set := make(map[string]struct{})
	for _, s := range strings.Fields(granted) {
		set[s] = struct{}{}
	}
	for _, req := range v.requiredScopes {
		if _, ok := set[req]; !ok {
			return false
		}
	}
	return true
}

func (v *cachingIntrospectionValidator) validate(ctx context.Context, token string) error {
	v.mu.Lock()
	if e, ok := v.m[token]; ok && e.exp.After(time.Now()) {
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
	if v.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+v.bearerToken)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return errors.New("introspection 429")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return errors.New("introspection failed")
	}

	var ir introspectResp
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return err
	}

	v.mu.Lock()
	v.m[token] = tokenCacheEntry{active: ir.Active, scope: ir.Scope, exp: time.Now().Add(v.ttl)}
	v.mu.Unlock()

	if !ir.Active {
		return errors.New("token inactive")
	}
	if !v.hasAllScopes(ir.Scope) {
		return errors.New("insufficient scope")
	}
	return nil
}
