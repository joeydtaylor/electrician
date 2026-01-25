package websocketclient

import (
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetURL sets the WebSocket endpoint URL.
func (c *WebSocketClientAdapter[T]) SetURL(url string) {
	c.configLock.Lock()
	c.url = strings.TrimSpace(url)
	c.configLock.Unlock()
}

// SetHeaders replaces all request headers.
func (c *WebSocketClientAdapter[T]) SetHeaders(headers map[string]string) {
	c.configLock.Lock()
	c.headers = cloneHeaderMap(headers)
	c.configLock.Unlock()
}

// AddHeader adds a single request header.
func (c *WebSocketClientAdapter[T]) AddHeader(key, value string) {
	if strings.TrimSpace(key) == "" {
		return
	}
	c.configLock.Lock()
	if c.headers == nil {
		c.headers = make(map[string]string)
	}
	c.headers[key] = value
	c.configLock.Unlock()
}

// SetMessageFormat sets the message encoding format.
func (c *WebSocketClientAdapter[T]) SetMessageFormat(format string) {
	c.configLock.Lock()
	if strings.TrimSpace(format) != "" {
		c.format = strings.ToLower(strings.TrimSpace(format))
	}
	c.configLock.Unlock()
}

// SetReadLimit sets the maximum inbound message size.
func (c *WebSocketClientAdapter[T]) SetReadLimit(limit int64) {
	c.configLock.Lock()
	if limit > 0 {
		c.readLimit = limit
	}
	c.configLock.Unlock()
}

// SetWriteTimeout sets the timeout for writes.
func (c *WebSocketClientAdapter[T]) SetWriteTimeout(timeout time.Duration) {
	c.configLock.Lock()
	if timeout > 0 {
		c.writeTimeout = timeout
	}
	c.configLock.Unlock()
}

// SetIdleTimeout sets the maximum idle duration for reads.
func (c *WebSocketClientAdapter[T]) SetIdleTimeout(timeout time.Duration) {
	c.configLock.Lock()
	c.idleTimeout = timeout
	c.configLock.Unlock()
}

// SetTLSConfig configures TLS for outbound connections.
func (c *WebSocketClientAdapter[T]) SetTLSConfig(tlsCfg types.TLSConfig) {
	cfg, err := buildTLSClientConfig(tlsCfg)
	c.configLock.Lock()
	c.tlsConfig = cfg
	c.tlsConfigErr = err
	c.configLock.Unlock()
}
