package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/codec"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// Fetch executes the configured request and decodes the response.
func (hp *HTTPClientAdapter[T]) Fetch() (types.HttpResponse[T], error) {
	cfg := hp.snapshotConfig()

	hp.notifyHTTPClientRequestStart()

	if cfg.oauthConfig != nil {
		if err := hp.EnsureValidToken(
			cfg.oauthConfig.ClientID,
			cfg.oauthConfig.Secret,
			cfg.oauthConfig.TokenURL,
			cfg.oauthConfig.Audience,
			cfg.oauthConfig.Scopes...,
		); err != nil {
			hp.notifyHTTPClientError(err)
			return types.HttpResponse[T]{}, err
		}
	}

	fetchCtx, fetchCancel := context.WithTimeout(cfg.ctx, cfg.timeout)
	defer fetchCancel()

	req, err := http.NewRequestWithContext(fetchCtx, cfg.request.Method, cfg.request.Endpoint, cfg.request.Body)
	if err != nil {
		hp.notifyHTTPClientError(err)
		return types.HttpResponse[T]{}, &types.HTTPError{StatusCode: 0, Err: err, Message: "failed to create HTTP request"}
	}

	for key, value := range cfg.headers {
		req.Header.Set(key, value)
	}

	resp, err := hp.httpClient.Do(req)
	if err != nil {
		hp.notifyHTTPClientError(err)
		return types.HttpResponse[T]{}, &types.HTTPError{StatusCode: 0, Err: err, Message: "http request execution failed"}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = fmt.Errorf("http request failed with status code: %d", resp.StatusCode)
		hp.notifyHTTPClientError(err)
		return types.HttpResponse[T]{StatusCode: resp.StatusCode}, &types.HTTPError{StatusCode: resp.StatusCode, Err: err, Message: "http request failed with non-success status"}
	}

	hp.notifyHTTPClientResponseReceived()

	wrappedResponse, err := hp.processResponse(resp)
	if err != nil {
		return types.HttpResponse[T]{}, err
	}

	hp.notifyHTTPClientRequestComplete()
	return types.HttpResponse[T]{StatusCode: wrappedResponse.StatusCode, Body: wrappedResponse.Data}, nil
}

func (hp *HTTPClientAdapter[T]) processResponse(resp *http.Response) (*types.WrappedResponse[T], error) {
	contentType := resp.Header.Get("Content-Type")
	var data T
	var raw []byte
	var decodeErr error

	switch {
	case strings.Contains(contentType, "application/json"):
		decoder := codec.NewJSONDecoder[T]()
		data, decodeErr = decoder.Decode(resp.Body)
	case strings.Contains(contentType, "text/xml"), strings.Contains(contentType, "application/xml"):
		decoder := codec.NewXMLDecoder[T]()
		data, decodeErr = decoder.Decode(resp.Body)
	case strings.Contains(contentType, "application/octet-stream"):
		binaryDecoder := codec.NewBinaryDecoder()
		raw, decodeErr = binaryDecoder.Decode(resp.Body)
	case strings.Contains(contentType, "text/plain"), strings.Contains(contentType, "text/html"):
		raw, decodeErr = io.ReadAll(resp.Body)
	default:
		raw, decodeErr = io.ReadAll(resp.Body)
		if decodeErr != nil {
			wrapped := types.NewWrappedResponse[T](data, raw, decodeErr, resp.Header)
			wrapped.StatusCode = resp.StatusCode
			return wrapped, fmt.Errorf("unsupported content type")
		}
	}

	wrapped := types.NewWrappedResponse[T](data, raw, decodeErr, resp.Header)
	wrapped.StatusCode = resp.StatusCode
	if decodeErr != nil {
		return wrapped, decodeErr
	}
	return wrapped, nil
}
