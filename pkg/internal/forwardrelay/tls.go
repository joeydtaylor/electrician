package forwardrelay

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc/credentials"
)

func (fr *ForwardRelay[T]) loadTLSCredentials(config *types.TLSConfig) (credentials.TransportCredentials, error) {
	loadedCreds := fr.tlsCredentials.Load()
	if loadedCreds != nil {
		return loadedCreds.(credentials.TransportCredentials), nil
	}

	fr.tlsCredentialsUpdate.Lock()
	defer fr.tlsCredentialsUpdate.Unlock()

	loadedCreds = fr.tlsCredentials.Load()
	if loadedCreds != nil {
		return loadedCreds.(credentials.TransportCredentials), nil
	}

	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		fr.NotifyLoggers(types.ErrorLevel, "loadTLSCredentials: load keypair failed: %v", err)
		return nil, err
	}
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(config.CAFile)
	if err != nil {
		fr.NotifyLoggers(types.ErrorLevel, "loadTLSCredentials: read CA failed: %v", err)
		return nil, err
	}
	if !certPool.AppendCertsFromPEM(ca) {
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

	newCreds := credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		ServerName:   config.SubjectAlternativeName,
		MinVersion:   minTLSVersion,
		MaxVersion:   maxTLSVersion,
	})
	fr.tlsCredentials.Store(newCreds)
	fr.NotifyLoggers(types.DebugLevel, "loadTLSCredentials: loaded")
	return newCreds, nil
}
