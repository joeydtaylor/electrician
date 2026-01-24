package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// GetTokenFromOAuthServer requests a client-credentials token from the OAuth server.
func (hp *HTTPClientAdapter[T]) GetTokenFromOAuthServer(clientID, clientSecret, tokenURL string, audience string, scopes ...string) (string, error) {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("grant_type", "client_credentials")
	values.Set("audience", audience)
	if len(scopes) > 0 {
		values.Set("scope", strings.Join(scopes, " "))
	}

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(values.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := hp.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	responseBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("oauth token request failed: %s", resp.Status)
	}

	var tokenResponse struct {
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int    `json:"expires_in"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"`
	}

	if err := json.Unmarshal(responseBytes, &tokenResponse); err != nil {
		return "", err
	}

	if tokenResponse.Error != "" {
		return "", fmt.Errorf("oauth error: %s - %s", tokenResponse.Error, tokenResponse.ErrorDescription)
	}

	expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	hp.tokenLock.Lock()
	hp.oAuthToken = &OAuthToken{
		AccessToken: tokenResponse.AccessToken,
		ExpiresAt:   expiresAt,
	}
	hp.tokenLock.Unlock()

	return tokenResponse.AccessToken, nil
}

// EnsureValidToken refreshes the OAuth token when needed.
func (hp *HTTPClientAdapter[T]) EnsureValidToken(clientID, clientSecret, tokenURL, audience string, scopes ...string) error {
	hp.tokenLock.Lock()
	token := hp.oAuthToken
	hp.tokenLock.Unlock()

	if token != nil && time.Now().Before(token.ExpiresAt) {
		return nil
	}

	_, err := hp.GetTokenFromOAuthServer(clientID, clientSecret, tokenURL, audience, scopes...)
	return err
}
