package receivingrelay

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc/credentials"
)

func (rr *ReceivingRelay[T]) loadTLSCredentials(config *types.TLSConfig) (credentials.TransportCredentials, error) {
	if !config.UseTLS {
		rr.NotifyLoggers(types.WarnLevel, "loadTLSCredentials: TLS disabled")
		return nil, fmt.Errorf("TLS is disabled")
	}

	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		rr.NotifyLoggers(types.ErrorLevel, "loadTLSCredentials: load keypair failed: %v", err)
		return nil, err
	}

	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(config.CAFile)
	if err != nil {
		rr.NotifyLoggers(types.ErrorLevel, "loadTLSCredentials: read CA failed: %v", err)
		return nil, err
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		rr.NotifyLoggers(types.ErrorLevel, "loadTLSCredentials: append CA failed")
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	minTLSVersion := config.MinTLSVersion
	if minTLSVersion == 0 {
		minTLSVersion = tls.VersionTLS12
	}
	maxTLSVersion := config.MaxTLSVersion
	if maxTLSVersion == 0 {
		maxTLSVersion = tls.VersionTLS13
	}

	return credentials.NewTLS(&tls.Config{
		ServerName:   config.SubjectAlternativeName,
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		MinVersion:   minTLSVersion,
		MaxVersion:   maxTLSVersion,
	}), nil
}
