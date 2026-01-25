package websocket

import (
	"context"
	"crypto/tls"
	"time"
)

type serverConfig struct {
	address         string
	endpoint        string
	headers         map[string]string
	allowedOrigins  []string
	tokenQueryParam string
	readLimit       int64
	writeTimeout    time.Duration
	idleTimeout     time.Duration
	sendBuffer      int
	maxConnections  int
	format          string
	tlsConfig       *tls.Config
	tlsConfigErr    error
	staticHeaders   map[string]string
	authValidator   func(ctx context.Context, headers map[string]string) error
	authRequired    bool
}

func (s *serverAdapter[T]) snapshotConfig() serverConfig {
	s.configLock.Lock()
	defer s.configLock.Unlock()

	s.ensureDefaultAuthValidatorLocked()

	return serverConfig{
		address:         s.address,
		endpoint:        s.endpoint,
		headers:         cloneHeaderMap(s.headers),
		allowedOrigins:  cloneStrings(s.allowedOrigins),
		tokenQueryParam: s.tokenQueryParam,
		readLimit:       s.readLimit,
		writeTimeout:    s.writeTimeout,
		idleTimeout:     s.idleTimeout,
		sendBuffer:      s.sendBuffer,
		maxConnections:  s.maxConnections,
		format:          normalizeFormat(s.format),
		tlsConfig:       s.tlsConfig,
		tlsConfigErr:    s.tlsConfigErr,
		staticHeaders:   cloneHeaderMap(s.staticHeaders),
		authValidator:   s.dynamicAuthValidator,
		authRequired:    s.authRequired,
	}
}
