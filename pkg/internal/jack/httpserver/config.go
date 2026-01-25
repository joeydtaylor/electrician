package httpserver

import (
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetAddress configures the listen address.
func (h *httpServerAdapter[T]) SetAddress(address string) {
	h.requireNotFrozen("SetAddress")
	h.configLock.Lock()
	h.address = address
	h.configLock.Unlock()
}

// SetServerConfig configures the HTTP method and endpoint.
func (h *httpServerAdapter[T]) SetServerConfig(method, endpoint string) {
	h.requireNotFrozen("SetServerConfig")
	h.configLock.Lock()
	h.method = method
	h.endpoint = endpoint
	h.configLock.Unlock()
}

// AddHeader adds a default response header.
func (h *httpServerAdapter[T]) AddHeader(key, value string) {
	h.requireNotFrozen("AddHeader")
	if key == "" {
		return
	}
	h.configLock.Lock()
	h.headers[key] = value
	h.configLock.Unlock()
}

// SetTimeout sets read/write timeouts.
func (h *httpServerAdapter[T]) SetTimeout(timeout time.Duration) {
	h.requireNotFrozen("SetTimeout")
	h.configLock.Lock()
	h.timeout = timeout
	h.configLock.Unlock()
}

// SetTLSConfig configures the server to use TLS for inbound connections.
func (h *httpServerAdapter[T]) SetTLSConfig(tlsCfg types.TLSConfig) {
	h.requireNotFrozen("SetTLSConfig")
	h.configLock.Lock()
	defer h.configLock.Unlock()

	if !tlsCfg.UseTLS {
		h.tlsConfig = nil
		h.tlsConfigErr = nil
		return
	}

	cfg, err := buildTLSConfig(tlsCfg)
	h.tlsConfig = cfg
	h.tlsConfigErr = err
}

func (h *httpServerAdapter[T]) snapshotConfig() serverConfig {
	h.configLock.Lock()
	defer h.configLock.Unlock()

	h.ensureDefaultAuthValidatorLocked()

	headers := make(map[string]string, len(h.headers))
	for k, v := range h.headers {
		headers[k] = v
	}

	staticHeaders := make(map[string]string, len(h.staticHeaders))
	for k, v := range h.staticHeaders {
		staticHeaders[k] = v
	}

	return serverConfig{
		address:       h.address,
		method:        h.method,
		endpoint:      h.endpoint,
		headers:       headers,
		timeout:       h.timeout,
		tlsConfig:     h.tlsConfig,
		tlsConfigErr:  h.tlsConfigErr,
		authRequired:  h.authRequired,
		staticHeaders: staticHeaders,
		authValidator: h.dynamicAuthValidator,
	}
}
