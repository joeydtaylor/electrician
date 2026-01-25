package websocketclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func buildTLSClientConfig(tlsCfg types.TLSConfig) (*tls.Config, error) {
	if !tlsCfg.UseTLS {
		return nil, nil
	}

	minTLSVersion := tlsCfg.MinTLSVersion
	if minTLSVersion == 0 {
		minTLSVersion = tls.VersionTLS12
	}
	maxTLSVersion := tlsCfg.MaxTLSVersion
	if maxTLSVersion == 0 {
		maxTLSVersion = tls.VersionTLS13
	}

	tlsConf := &tls.Config{
		MinVersion: minTLSVersion,
		MaxVersion: maxTLSVersion,
		ServerName: tlsCfg.SubjectAlternativeName,
	}

	if tlsCfg.CertFile != "" || tlsCfg.KeyFile != "" {
		if tlsCfg.CertFile == "" || tlsCfg.KeyFile == "" {
			return nil, fmt.Errorf("both CertFile and KeyFile are required for client certs")
		}
		cert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load client certificate/key: %w", err)
		}
		tlsConf.Certificates = []tls.Certificate{cert}
	}

	if tlsCfg.CAFile != "" {
		caData, err := os.ReadFile(tlsCfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("unable to read CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if ok := caPool.AppendCertsFromPEM(caData); !ok {
			return nil, fmt.Errorf("failed to parse CA certificate(s) in %s", tlsCfg.CAFile)
		}
		tlsConf.RootCAs = caPool
	}

	return tlsConf, nil
}
