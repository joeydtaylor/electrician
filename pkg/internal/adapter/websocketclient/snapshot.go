package websocketclient

import (
	"crypto/tls"
	"time"
)

type clientConfig struct {
	url          string
	headers      map[string]string
	format       string
	readLimit    int64
	writeTimeout time.Duration
	idleTimeout  time.Duration
	tlsConfig    *tls.Config
	tlsConfigErr error
}

func (c *WebSocketClientAdapter[T]) snapshotConfig() clientConfig {
	c.configLock.Lock()
	defer c.configLock.Unlock()

	return clientConfig{
		url:          c.url,
		headers:      cloneHeaderMap(c.headers),
		format:       normalizeFormat(c.format),
		readLimit:    c.readLimit,
		writeTimeout: c.writeTimeout,
		idleTimeout:  c.idleTimeout,
		tlsConfig:    c.tlsConfig,
		tlsConfigErr: c.tlsConfigErr,
	}
}
