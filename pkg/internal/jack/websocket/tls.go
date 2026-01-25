package websocket

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

func buildTLSConfig(tlsCfg types.TLSConfig) (*tls.Config, error) {
	if !tlsCfg.UseTLS {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(tlsCfg.CertFile, tlsCfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate/key: %w", err)
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tlsCfg.MinTLSVersion,
		MaxVersion:   tlsCfg.MaxTLSVersion,
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
		tlsConf.ClientCAs = caPool
	}

	return tlsConf, nil
}
