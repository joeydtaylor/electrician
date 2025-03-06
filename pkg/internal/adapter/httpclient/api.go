package httpclient

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// WithRequestConfig allows setting the request configuration.
func (p *HTTPClientAdapter[T]) SetRequestConfig(method, endpoint string, body io.Reader) {
	p.requestConfig = RequestConfig{
		Method:   method,
		Endpoint: endpoint,
		Body:     body,
	}
}

// SetTlsPinnedCertificate configures TLS pinning with the specified certificate.
func (hp *HTTPClientAdapter[T]) SetTlsPinnedCertificate(certPath string) {
	// Store the pinned certificate for later verification
	cert, err := loadCertificate(certPath)
	if err != nil {
		hp.notifyHTTPClientError(err)
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, result: SUCCESS, event: SetTlsPinnedCertificate, error: %v, certPath: %v => Unable to load certificate", hp.componentMetadata, err, certPath)
	}

	hp.pinnedCert = cert
	hp.tlsPinningEnabled = true

	// Setup the custom TLS configuration
	tlsConfig := &tls.Config{
		VerifyPeerCertificate: hp.verifyServerCertificate,
	}

	hp.httpClient.Transport = &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	hp.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: SetTlsPinnedCertificate, certPath: %v => Pinned certificate", hp.componentMetadata, certPath)
}

// WithRequestConfig allows setting the request configuration.
func (p *HTTPClientAdapter[T]) SetOAuth2Config(clientID, clientSecret, tokenURL string, audience string, scopes ...string) {
	p.oauth2Config = &OAuth2Config{
		ClientID: clientID,
		Secret:   clientSecret,
		TokenURL: tokenURL,
		Audience: audience,
		Scopes:   scopes,
	}
}

// WithInterval sets the interval for making requests.
func (p *HTTPClientAdapter[T]) SetInterval(interval time.Duration) {
	p.interval = interval
}

// WithInterval sets the interval for making requests.
func (p *HTTPClientAdapter[T]) SetTimeout(timeout time.Duration) {
	p.timeout = timeout
}

// WithInterval sets the interval for making requests.
func (p *HTTPClientAdapter[T]) SetMaxRetries(retries int) {
	p.maxRetries = retries
}

// WithHeader adds a header to the request.
func (p *HTTPClientAdapter[T]) AddHeader(key, value string) {
	err := p.validateHeaderValue(value)
	if err == nil {
		p.headers[key] = value
	}
}

func (p *HTTPClientAdapter[T]) ConnectSensor(sensor ...types.Sensor[T]) {
	p.sensors = append(p.sensors, sensor...)
	for _, m := range sensor {
		p.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: ConnectSensor, target: %v => Connected sensor", p.componentMetadata, m.GetComponentMetadata())
	}
}

func (p *HTTPClientAdapter[T]) ConnectLogger(l ...types.Logger) {
	p.loggersLock.Lock()
	defer p.loggersLock.Unlock()
	p.loggers = append(p.loggers, l...)
}

func (p *HTTPClientAdapter[T]) NotifyLoggers(level types.LogLevel, format string, args ...interface{}) {
	if p.loggers != nil {
		msg := fmt.Sprintf(format, args...)
		for _, logger := range p.loggers {
			if logger == nil {
				continue
			}
			// Ensure we only acquire the lock once per logger to avoid deadlock or excessive locking overhead
			p.loggersLock.Lock()
			if logger.GetLevel() <= level {
				switch level {
				case types.DebugLevel:
					logger.Debug(msg)
				case types.InfoLevel:
					logger.Info(msg)
				case types.WarnLevel:
					logger.Warn(msg)
				case types.ErrorLevel:
					logger.Error(msg)
				case types.DPanicLevel:
					logger.DPanic(msg)
				case types.PanicLevel:
					logger.Panic(msg)
				case types.FatalLevel:
					logger.Fatal(msg)
				}
			}
			p.loggersLock.Unlock()
		}
	}
}

// GetComponentMetadata returns the metadata.
func (p *HTTPClientAdapter[T]) GetComponentMetadata() types.ComponentMetadata {
	return p.componentMetadata
}

// SetComponentMetadata sets the component metadata.
func (p *HTTPClientAdapter[T]) SetComponentMetadata(name string, id string) {
	p.componentMetadata = types.ComponentMetadata{Name: name, ID: id}
}

// SetBasicAuth configures the plug to use basic authentication.
func (hp *HTTPClientAdapter[T]) SetBasicAuth(username, password string) {
	hp.configLock.Lock()
	defer hp.configLock.Unlock()
	basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	hp.headers["Authorization"] = basicAuth
}

// Serve starts the plug that performs HTTP requests at configurable intervals.
func (hp *HTTPClientAdapter[T]) Serve(ctx context.Context, submitFunc func(ctx context.Context, elem T) error) error {
	// Ensure only one instance of Serve is running
	if !atomic.CompareAndSwapInt32(&hp.isServing, 0, 1) {
		return nil // Serve is already running
	}
	defer atomic.StoreInt32(&hp.isServing, 0) // Reset when done

	ticker := time.NewTicker(hp.interval)
	defer ticker.Stop()

	hp.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Serve, result: SUCCESS, interval: %d => HTTPClientAdapter Serve started", hp.componentMetadata, hp.interval)

	for {
		select {
		case <-ctx.Done():
			hp.NotifyLoggers(types.WarnLevel, "%s => level: WARN, event: Cancel, result: PENDING => HTTP Adapter stopped due to context cancellation.", hp.componentMetadata)
			return nil
		case <-ticker.C:
			err := hp.attemptFetchAndSubmit(ctx, submitFunc, hp.maxRetries)
			if err != nil {
				hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Serve, result: FAILURE, message: %v => All retries exhausted, propagating error.", hp.componentMetadata, err)
				return err // Only return an error after all retries are exhausted
			}
		}
	}
}

func (hp *HTTPClientAdapter[T]) attemptFetchAndSubmit(ctx context.Context, submitFunc func(ctx context.Context, elem T) error, maxRetries int) error {
	var attempt int
	var httpError error
	for attempt = 0; attempt <= maxRetries; attempt++ {
		expBackoff := time.Duration(math.Pow(2, float64(attempt))) * time.Second

		// Apply full jitter
		jitter := time.Duration(rand.Int63n(int64(expBackoff)))
		sleepDuration := jitter
		hp.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Serve, result: PENDING, attempt: %d, endpoint: %s%s => Attempting to fetch data...", hp.componentMetadata, attempt, hp.baseURL, hp.requestConfig.Endpoint)
		response, err := hp.Fetch()
		httpError = err
		if err != nil {
			hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Fetch, result: FAILURE, attempt: %d, endpoint: %s%s, error: %v => Error during request!", hp.componentMetadata, attempt, hp.baseURL, hp.requestConfig.Endpoint, err)
			time.Sleep(sleepDuration)
			continue // Retry the request
		}

		if err = submitFunc(ctx, response.Body); err != nil {
			hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Submit, result: FAILURE, attempt: %d, endpoint: %s%s, error: %v => Error submitting response!", hp.componentMetadata, attempt, hp.baseURL, hp.requestConfig.Endpoint, err)
			time.Sleep(sleepDuration)
			continue // Retry the submission
		}

		return nil // Success, exit the retry loop
	}
	return fmt.Errorf("retries exhausted after %d attempts, last error: %v", maxRetries, httpError) // Final error after retries are done
}

func (hp *HTTPClientAdapter[T]) GetTokenFromOAuthServer(clientID, clientSecret, tokenURL string, audience string, scopes ...string) (string, error) {
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("grant_type", "client_credentials")
	values.Set("audience", audience)
	if len(scopes) > 0 {
		values.Set("scope", strings.Join(scopes, " "))
	}

	requestBody := values.Encode()
	hp.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: OAuthTokenRequest, status: START, message: Preparing token request for URL: %s with payload: %s", hp.componentMetadata, tokenURL, requestBody)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(requestBody))
	if err != nil {
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: OAuthTokenRequest, status: ERROR, message: Failed to create request: %v", hp.componentMetadata, err)
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := hp.httpClient.Do(req)
	if err != nil {
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: OAuthTokenRequest, status: ERROR, message: Request failed: %v", hp.componentMetadata, err)
		return "", err
	}
	defer resp.Body.Close()

	responseBytes, _ := io.ReadAll(resp.Body)
	hp.NotifyLoggers(types.DebugLevel, "%s => level: INFO, event: OAuthTokenRequest, status: RESPONSE, message: Received raw response: %s", hp.componentMetadata, string(responseBytes))

	var tokenResponse struct {
		AccessToken      string `json:"access_token"`
		TokenType        string `json:"token_type"`
		ExpiresIn        int    `json:"expires_in"`
		Scope            string `json:"scope"`
		Error            string `json:"error"`
		ErrorDescription string `json:"error_description"` // Additional field to capture more detailed error messages
	}

	if err := json.Unmarshal(responseBytes, &tokenResponse); err != nil {
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: OAuthTokenRequest, status: ERROR, message: Failed to decode response: %v", hp.componentMetadata, err)
		return "", err
	}

	if tokenResponse.Error != "" {
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: OAuthTokenRequest, status: ERROR, message: OAuth error received: %s - %s", hp.componentMetadata, tokenResponse.Error, tokenResponse.ErrorDescription)
		return "", fmt.Errorf("OAuth error: %s - %s", tokenResponse.Error, tokenResponse.ErrorDescription)
	}

	// Calculate the token's expiry time
	expiresAt := time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second)
	hp.tokenMutex.Lock()
	hp.oAuthToken = &OAuthToken{
		AccessToken: tokenResponse.AccessToken,
		ExpiresAt:   expiresAt,
	}
	hp.tokenMutex.Unlock()

	hp.NotifyLoggers(types.DebugLevel, "%s => level: INFO, event: OAuthTokenRequest, status: SUCCESS, tokenURL: %v, response: %v => Token retrieved successfully", hp.componentMetadata, hp.oauth2Config.TokenURL, tokenResponse.AccessToken)
	return tokenResponse.AccessToken, nil
}

func (hp *HTTPClientAdapter[T]) EnsureValidToken(clientID, clientSecret, tokenURL, audience string, scopes ...string) error {
	hp.tokenMutex.Lock()
	defer hp.tokenMutex.Unlock()

	// Check if the token is still valid
	if hp.oAuthToken != nil && time.Now().Before(hp.oAuthToken.ExpiresAt) {
		return nil
	}

	// Token is invalid or expired, fetch a new one
	_, err := hp.GetTokenFromOAuthServer(clientID, clientSecret, tokenURL, audience, scopes...)
	return err
}

func (hp *HTTPClientAdapter[T]) Fetch() (types.HttpResponse[T], error) {
	hp.notifyHTTPClientRequestStart()
	if hp.oauth2Config != nil {
		if err := hp.EnsureValidToken(hp.oauth2Config.ClientID, hp.oauth2Config.Secret, hp.oauth2Config.TokenURL, hp.oauth2Config.Audience, hp.oauth2Config.Scopes...); err != nil {
			return types.HttpResponse[T]{}, err
		}
	}

	fetchCtx, fetchCtxCancel := context.WithTimeout(hp.ctx, hp.timeout)
	defer fetchCtxCancel()

	req, err := http.NewRequestWithContext(fetchCtx, hp.requestConfig.Method, hp.baseURL+hp.requestConfig.Endpoint, hp.requestConfig.Body)
	if err != nil {
		hp.notifyHTTPClientError(err)
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Fetch, result: FAILURE, error: %v => Failed to create HTTP request!", hp.componentMetadata, err)
		return types.HttpResponse[T]{}, &types.HTTPError{StatusCode: 0, Err: err, Message: "Failed to create HTTP request"}
	}

	for key, value := range hp.headers {
		req.Header.Set(key, value)
	}

	resp, err := hp.httpClient.Do(req)
	if err != nil {
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Fetch, result: FAILURE, url: %s, method: %s, error: %v => Error executing request!", hp.componentMetadata, req.URL, req.Method, err)
		hp.notifyHTTPClientError(err)
		return types.HttpResponse[T]{}, &types.HTTPError{StatusCode: 0, Err: err, Message: "HTTP request execution failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Handle non-success status codes
		err = fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
		hp.notifyHTTPClientError(err)
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, event: Fetch, result: FAILURE, url: %s, method: %s, statusCode: %d => Non-success HTTP status received", hp.componentMetadata, req.URL, req.Method, resp.StatusCode)
		return types.HttpResponse[T]{StatusCode: resp.StatusCode}, &types.HTTPError{StatusCode: resp.StatusCode, Err: err, Message: "HTTP request failed with non-success status"}
	}

	hp.notifyHTTPClientRequestReceived()

	wrappedResponse, err := hp.processResponse(resp)
	if err != nil {
		return types.HttpResponse[T]{}, err
	}

	// Extract the response data from the wrapped response
	hp.notifyHTTPClientRequestComplete()
	return types.HttpResponse[T]{StatusCode: wrappedResponse.StatusCode, Body: wrappedResponse.Data}, nil
}
