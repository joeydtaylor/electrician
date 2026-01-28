package httpclient

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

// SetTlsPinnedCertificate enables TLS pinning with the provided certificate file.
func (hp *HTTPClientAdapter[T]) SetTlsPinnedCertificate(certPath string) {
	cert, err := loadCertificate(certPath)
	if err != nil {
		hp.notifyHTTPClientError(err)
		hp.NotifyLoggers(
			types.ErrorLevel,
			"httpclient tls pinning failed",
			"component", hp.GetComponentMetadata(),
			"event", "TLSPinning",
			"cert_path", certPath,
			"error", err,
		)
		return
	}

	hp.configLock.Lock()
	hp.pinnedCert = cert
	hp.pinEnabled = true
	hp.configLock.Unlock()

	tlsConfig := &tls.Config{VerifyPeerCertificate: hp.verifyServerCertificate}

	hp.configLock.Lock()
	if hp.httpClient != nil {
		hp.httpClient.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	}
	hp.configLock.Unlock()
}

func (hp *HTTPClientAdapter[T]) verifyServerCertificate(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
	hp.configLock.Lock()
	enabled := hp.pinEnabled
	pinned := append([]byte(nil), hp.pinnedCert...)
	hp.configLock.Unlock()

	if !enabled {
		return nil
	}

	for _, chain := range verifiedChains {
		for _, cert := range chain {
			if bytes.Equal(cert.Raw, pinned) {
				return nil
			}
		}
	}

	return errors.New("TLS certificate pinning check failed")
}

func loadCertificate(certPath string) ([]byte, error) {
	file, err := os.Open(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate file: %w", err)
	}
	defer file.Close()

	certData, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	block, _ := pem.Decode(certData)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM block containing the certificate")
	}
	return block.Bytes, nil
}
