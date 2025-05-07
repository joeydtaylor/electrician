// File: internal.go
package httpserver

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
)

// -----------------------------------------------------------------------------
// Helper Methods
// -----------------------------------------------------------------------------

// parseRequest reads and decodes the request as JSON (simple placeholder).
func (h *httpServerAdapter[T]) parseRequest(r *http.Request) (*types.WrappedResponse[T], error) {
	var decoded T
	err := utils.DecodeJSON(r.Body, &decoded)
	if err != nil {
		return nil, err
	}
	return types.NewWrappedResponse(decoded, nil, nil, r.Header), nil
}

// notifyHTTPServerError notifies sensors & logs an error event.
func (h *httpServerAdapter[T]) notifyHTTPServerError(err error) {
	h.sensorsLock.Lock()
	defer h.sensorsLock.Unlock()

	for _, sensor := range h.sensors {
		// Optionally define a dedicated OnHTTPServerError method in sensors
		// but we can reuse OnHTTPClientError for the example
		sensor.InvokeOnHTTPClientError(h.componentMetadata, err)
	}

	h.NotifyLoggers(types.ErrorLevel,
		"%s => level: ERROR, event: notifyHTTPServerError, error: %v => Notified sensors",
		h.componentMetadata, err)
}

// buildTLSConfig creates a *tls.Config by loading certs/keys if UseTLS is true.
// Returns nil if UseTLS=false, or if no cert/key are provided.
func buildTLSConfig(tlsCfg types.TLSConfig) (*tls.Config, error) {
	// If TLS is disabled, return nil to indicate plain HTTP.
	if !tlsCfg.UseTLS {
		return nil, nil
	}

	// Load the server certificate/key from files, if provided.
	cert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate/key: %w", err)
	}

	// Prepare our tls.Config
	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tlsCfg.MinTLSVersion, // e.g., tls.VersionTLS12
		MaxVersion:   tlsCfg.MaxTLSVersion, // e.g., tls.VersionTLS13
	}

	// If a CAFile is provided, load it to set up a custom CA pool (common for mutual TLS).
	if tlsCfg.CAFile != "" {
		caData, err := ioutil.ReadFile(tlsCfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if ok := caPool.AppendCertsFromPEM(caData); !ok {
			return nil, fmt.Errorf("failed to parse CA certificate(s) in %s", tlsCfg.CAFile)
		}
		tlsConf.ClientCAs = caPool
		// If you want to enforce client cert checking (mTLS), set:
		// tlsConf.ClientAuth = tls.RequireAndVerifyClientCert
	}

	return tlsConf, nil
}
