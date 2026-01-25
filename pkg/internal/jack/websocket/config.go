package websocket

import (
	"strings"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func (s *serverAdapter[T]) SetAddress(address string) {
	s.requireNotFrozen("SetAddress")
	s.configLock.Lock()
	s.address = strings.TrimSpace(address)
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetEndpoint(path string) {
	s.requireNotFrozen("SetEndpoint")
	s.configLock.Lock()
	s.endpoint = strings.TrimSpace(path)
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) AddHeader(key, value string) {
	s.requireNotFrozen("AddHeader")
	if strings.TrimSpace(key) == "" {
		return
	}
	s.configLock.Lock()
	s.headers[key] = value
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetAllowedOrigins(origins ...string) {
	s.requireNotFrozen("SetAllowedOrigins")
	s.configLock.Lock()
	s.allowedOrigins = trimStrings(origins)
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetTokenQueryParam(name string) {
	s.requireNotFrozen("SetTokenQueryParam")
	s.configLock.Lock()
	s.tokenQueryParam = strings.TrimSpace(name)
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetMessageFormat(format string) {
	s.requireNotFrozen("SetMessageFormat")
	s.configLock.Lock()
	if strings.TrimSpace(format) != "" {
		s.format = strings.ToLower(strings.TrimSpace(format))
	}
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetReadLimit(limit int64) {
	s.requireNotFrozen("SetReadLimit")
	s.configLock.Lock()
	if limit > 0 {
		s.readLimit = limit
	}
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetWriteTimeout(timeout time.Duration) {
	s.requireNotFrozen("SetWriteTimeout")
	s.configLock.Lock()
	if timeout > 0 {
		s.writeTimeout = timeout
	}
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetIdleTimeout(timeout time.Duration) {
	s.requireNotFrozen("SetIdleTimeout")
	s.configLock.Lock()
	s.idleTimeout = timeout
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetSendBuffer(size int) {
	s.requireNotFrozen("SetSendBuffer")
	s.configLock.Lock()
	if size > 0 {
		s.sendBuffer = size
	}
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetMaxConnections(max int) {
	s.requireNotFrozen("SetMaxConnections")
	s.configLock.Lock()
	if max >= 0 {
		s.maxConnections = max
	}
	s.configLock.Unlock()
}

func (s *serverAdapter[T]) SetTLSConfig(tlsCfg types.TLSConfig) {
	s.requireNotFrozen("SetTLSConfig")
	cfg, err := buildTLSConfig(tlsCfg)
	s.configLock.Lock()
	s.tlsConfig = cfg
	s.tlsConfigErr = err
	s.configLock.Unlock()
}
