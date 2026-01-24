package httpclient

import (
	"encoding/base64"
	"io"
	"time"
)

// SetRequestConfig updates the HTTP request template.
func (hp *HTTPClientAdapter[T]) SetRequestConfig(method, endpoint string, body io.Reader) {
	hp.configLock.Lock()
	hp.request = RequestConfig{Method: method, Endpoint: endpoint, Body: body}
	hp.configLock.Unlock()
}

// SetInterval updates the polling interval used by Serve.
func (hp *HTTPClientAdapter[T]) SetInterval(interval time.Duration) {
	hp.configLock.Lock()
	hp.interval = interval
	hp.configLock.Unlock()
}

// SetTimeout updates the per-request timeout.
func (hp *HTTPClientAdapter[T]) SetTimeout(timeout time.Duration) {
	hp.configLock.Lock()
	hp.timeout = timeout
	if hp.httpClient != nil {
		hp.httpClient.Timeout = timeout
	}
	hp.configLock.Unlock()
}

// SetMaxRetries updates the maximum number of retry attempts.
func (hp *HTTPClientAdapter[T]) SetMaxRetries(retries int) {
	hp.configLock.Lock()
	hp.maxRetries = retries
	hp.configLock.Unlock()
}

// SetOAuth2Config stores OAuth2 client credentials.
func (hp *HTTPClientAdapter[T]) SetOAuth2Config(clientID, clientSecret, tokenURL string, audience string, scopes ...string) {
	hp.configLock.Lock()
	hp.oauthConfig = &OAuth2Config{
		ClientID: clientID,
		Secret:   clientSecret,
		TokenURL: tokenURL,
		Audience: audience,
		Scopes:   append([]string(nil), scopes...),
	}
	hp.configLock.Unlock()
}

// AddHeader adds a static header to outgoing requests.
func (hp *HTTPClientAdapter[T]) AddHeader(key, value string) {
	if err := hp.validateHeaderValue(value); err != nil {
		return
	}

	hp.configLock.Lock()
	if hp.headers == nil {
		hp.headers = make(map[string]string)
	}
	hp.headers[key] = value
	hp.configLock.Unlock()
}

// SetBasicAuth configures HTTP basic authentication.
func (hp *HTTPClientAdapter[T]) SetBasicAuth(username, password string) {
	basicAuth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	hp.AddHeader("Authorization", "Basic "+basicAuth)
}
