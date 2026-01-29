//go:build webtransport

package webtransportrelay

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/quic-go/quic-go/http3"
)

func (fr *ForwardRelay[T]) buildClientTLSConfig() (*tls.Config, error) {
	if fr.TlsConfig == nil || !fr.TlsConfig.UseTLS {
		if fr.tokenSource != nil && fr.authRequired {
			return nil, fmt.Errorf("oauth2 enabled and required but TLS is disabled; refuse to dial insecure")
		}
		return &tls.Config{InsecureSkipVerify: true, NextProtos: []string{http3.NextProtoH3}, MinVersion: tls.VersionTLS13, MaxVersion: tls.VersionTLS13}, nil
	}

	certPool := x509.NewCertPool()
	if fr.TlsConfig.CAFile != "" {
		ca, err := os.ReadFile(fr.TlsConfig.CAFile)
		if err != nil {
			return nil, err
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
	}

	var certs []tls.Certificate
	if fr.TlsConfig.CertFile != "" && fr.TlsConfig.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(fr.TlsConfig.CertFile, fr.TlsConfig.KeyFile)
		if err != nil {
			return nil, err
		}
		certs = []tls.Certificate{cert}
	}

	return &tls.Config{
		ServerName:         fr.TlsConfig.SubjectAlternativeName,
		Certificates:       certs,
		RootCAs:            certPool,
		NextProtos:         []string{http3.NextProtoH3},
		MinVersion:         tls.VersionTLS13,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: fr.TlsConfig.CAFile == "",
	}, nil
}

func (rr *ReceivingRelay[T]) buildServerTLSConfig() (*tls.Config, error) {
	cfg := rr.TlsConfig
	if cfg == nil || !cfg.UseTLS {
		return nil, fmt.Errorf("TLS required for WebTransport listeners")
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, err
	}

	var certPool *x509.CertPool
	if cfg.CAFile != "" {
		certPool = x509.NewCertPool()
		ca, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, err
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			return nil, fmt.Errorf("failed to append CA certificate")
		}
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    certPool,
		NextProtos:   []string{http3.NextProtoH3},
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}, nil
}
