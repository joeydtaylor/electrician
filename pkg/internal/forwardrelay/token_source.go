package forwardrelay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type staticTokenSource struct{ tok string }

func (s staticTokenSource) AccessToken(_ context.Context) (string, error) { return s.tok, nil }

// NewStaticBearerTokenSource returns a token source that always returns the provided token.
func NewStaticBearerTokenSource(token string) types.OAuth2TokenSource {
	return staticTokenSource{tok: token}
}

type envTokenSource struct{ key string }

func (e envTokenSource) AccessToken(_ context.Context) (string, error) {
	return os.Getenv(e.key), nil
}

// NewEnvBearerTokenSource returns a token source that pulls the token from an env var at call time.
func NewEnvBearerTokenSource(envVar string) types.OAuth2TokenSource {
	return envTokenSource{key: envVar}
}

type refreshingCCSource struct {
	tokenURL     string
	clientID     string
	clientSecret string
	scopes       []string

	hc   *http.Client
	mu   sync.Mutex
	tok  string
	exp  time.Time
	skew time.Duration
}

type ccTokenResp struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int32  `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

// NewClientCredentialsBearerTokenSource returns a token source that fetches client_credentials tokens.
func NewClientCredentialsBearerTokenSource(
	baseURL, clientID, clientSecret string,
	scopes []string,
	httpClient *http.Client,
) types.OAuth2TokenSource {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &refreshingCCSource{
		tokenURL:     strings.TrimRight(baseURL, "/") + "/api/auth/oauth/token",
		clientID:     clientID,
		clientSecret: clientSecret,
		scopes:       append([]string(nil), scopes...),
		hc:           httpClient,
		skew:         20 * time.Second,
	}
}

// NewRefreshingClientCredentialsSource is a convenience alias with custom refresh leeway.
func NewRefreshingClientCredentialsSource(
	baseURL, clientID, clientSecret string,
	scopes []string,
	leeway time.Duration,
	httpClient *http.Client,
) types.OAuth2TokenSource {
	ts := NewClientCredentialsBearerTokenSource(baseURL, clientID, clientSecret, scopes, httpClient)
	if s, ok := ts.(*refreshingCCSource); ok {
		if leeway > 0 {
			s.skew = leeway
		}
	}
	return ts
}

func (s *refreshingCCSource) AccessToken(ctx context.Context) (string, error) {
	s.mu.Lock()
	tok, exp := s.tok, s.exp
	s.mu.Unlock()

	if tok != "" && time.Now().Add(s.skew).Before(exp) {
		return tok, nil
	}

	form := url.Values{}
	form.Set("grant_type", "client_credentials")
	if len(s.scopes) > 0 {
		form.Set("scope", strings.Join(s.scopes, " "))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(s.clientID, s.clientSecret)

	resp, err := s.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("token endpoint status %s", resp.Status)
	}

	var tr ccTokenResp
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	if tr.AccessToken == "" {
		return "", errors.New("empty access_token")
	}

	expAt := time.Now().Add(time.Duration(tr.ExpiresIn) * time.Second)
	if tr.ExpiresIn == 0 {
		expAt = time.Now().Add(60 * time.Second)
	}

	s.mu.Lock()
	s.tok = tr.AccessToken
	s.exp = expAt
	s.mu.Unlock()

	return tr.AccessToken, nil
}
