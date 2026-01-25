package websocketclient

import (
	"context"
	"errors"
	"net/http"

	"nhooyr.io/websocket"
)

func (c *WebSocketClientAdapter[T]) dial(ctx context.Context, cfg clientConfig) (*websocket.Conn, error) {
	if cfg.tlsConfigErr != nil {
		return nil, cfg.tlsConfigErr
	}
	if cfg.url == "" {
		return nil, errors.New("url not configured")
	}
	if err := validateFormat(cfg.format); err != nil {
		return nil, err
	}

	hdr := http.Header{}
	for k, v := range cfg.headers {
		hdr.Add(k, v)
	}

	opts := &websocket.DialOptions{HTTPHeader: hdr}
	if cfg.tlsConfig != nil {
		opts.HTTPClient = &http.Client{
			Transport: &http.Transport{TLSClientConfig: cfg.tlsConfig},
		}
	}

	conn, resp, err := websocket.Dial(ctx, cfg.url, opts)
	if resp != nil && resp.Body != nil {
		_ = resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	if cfg.readLimit > 0 {
		conn.SetReadLimit(cfg.readLimit)
	}
	return conn, nil
}
