package httpclient

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/joeydtaylor/electrician/pkg/internal/codec"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

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
		// Handling binary data
		binaryDecoder := codec.NewBinaryDecoder()
		raw, decodeErr = binaryDecoder.Decode(resp.Body)
	case strings.Contains(contentType, "text/plain"):
		// Handling plain text as raw data
		raw, decodeErr = io.ReadAll(resp.Body)

	case strings.Contains(contentType, "text/html"):
		// Directly read the HTML content into the raw byte slice
		raw, decodeErr = io.ReadAll(resp.Body)
	default:
		// Fallback for other unsupported content types
		raw, decodeErr = io.ReadAll(resp.Body) // Read unknown content type as raw data
		if decodeErr != nil {
			return types.NewWrappedResponse[T](data, raw, decodeErr, resp.Header), fmt.Errorf("unsupported content type")
		}
	}

	if decodeErr != nil {
		return types.NewWrappedResponse[T](data, raw, decodeErr, resp.Header), decodeErr
	}
	return types.NewWrappedResponse[T](data, raw, nil, resp.Header), nil
}

// notifyStart notifies all sensors of the wire that the wire has started.
// It invokes the OnStart callback for each sensor.
func (hp *HTTPClientAdapter[T]) notifyHTTPClientRequestStart() {
	for _, sensor := range hp.sensors {
		hp.sensorLock.Lock()
		defer hp.sensorLock.Unlock()
		sensor.InvokeOnHTTPClientRequestStart(hp.componentMetadata)

		hp.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyHTTPClientRequestStart, target_component: %s => Invoked InvokeOnHTTPClientRequestStart for sensor", hp.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

// notifyStart notifies all sensors of the wire that the wire has started.
// It invokes the OnStart callback for each sensor.
func (hp *HTTPClientAdapter[T]) notifyHTTPClientError(err error) {
	for _, sensor := range hp.sensors {
		hp.sensorLock.Lock()
		defer hp.sensorLock.Unlock()
		sensor.InvokeOnHTTPClientError(hp.componentMetadata, err)

		hp.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyHTTPClientError, error: %v, target_component: %s => Invoked InvokeOnHTTPClientRequestStart for sensor", hp.GetComponentMetadata(), err, sensor.GetComponentMetadata())
	}
}

// notifyStart notifies all sensors of the wire that the wire has started.
// It invokes the OnStart callback for each sensor.
func (hp *HTTPClientAdapter[T]) notifyHTTPClientRequestReceived() {
	for _, sensor := range hp.sensors {
		hp.sensorLock.Lock()
		defer hp.sensorLock.Unlock()
		sensor.InvokeOnHTTPClientResponseReceived(hp.componentMetadata)

		hp.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyHTTPClientRequestReceived, target_component: %s => Invoked InvokeOnHTTPClientResponseReceived for sensor", hp.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

// notifyStart notifies all sensors of the wire that the wire has started.
// It invokes the OnStart callback for each sensor.
func (hp *HTTPClientAdapter[T]) notifyHTTPClientRequestComplete() {
	for _, sensor := range hp.sensors {
		hp.sensorLock.Lock()
		defer hp.sensorLock.Unlock()
		sensor.InvokeOnHTTPClientRequestComplete(hp.componentMetadata)

		hp.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, result: SUCCESS, event: notifyHTTPClientRequestComplete, target_component: %s => Invoked InvokeOnHTTPClientRequestComplete for sensor", hp.GetComponentMetadata(), sensor.GetComponentMetadata())
	}
}

// validateHeaderValue ensures that headers do not contain malicious characters.
func (hp *HTTPClientAdapter[T]) validateHeaderValue(value string) error {
	if regexp.MustCompile(`[\r\n]`).MatchString(value) {
		err := fmt.Errorf("header contains unsupported characeter")
		hp.NotifyLoggers(types.ErrorLevel, "%s => level: ERROR, result: FAILURE, event: AddHeader, error: %v, header: %s = %s => Failed to add header!", hp.componentMetadata, err, value, fmt.Errorf("header contains unsupported characeter"))
		hp.notifyHTTPClientError(err)
		return err
	}
	return nil
}

// verifyServerCertificate checks if the server's certificate matches the pinned certificate.
func (hp *HTTPClientAdapter[T]) verifyServerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if !hp.tlsPinningEnabled {
		return nil // TLS pinning is not enabled, skip verification
	}

	// Attempt to find the pinned certificate in the verified chains
	for _, chain := range verifiedChains {
		for _, cert := range chain {
			if bytes.Equal(cert.Raw, hp.pinnedCert) { // Compare raw certificate data
				return nil
			}
		}
	}

	return errors.New("TLS certificate pinning check failed")
}

// loadCertificate loads the PEM-encoded certificate from the specified file path using io and os packages.
func loadCertificate(certPath string) ([]byte, error) {
	file, err := os.Open(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate file: %v", err)
	}
	defer file.Close()

	// Read the entire file content
	certData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate file: %v", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing the certificate: %v", err)
	}
	return block.Bytes, nil
}
