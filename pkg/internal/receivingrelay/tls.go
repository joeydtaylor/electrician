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
		rr.logKV(
			types.WarnLevel,
			"TLS disabled",
			"event", "TLSCredentials",
			"result", "SKIPPED",
		)
		return nil, fmt.Errorf("TLS is disabled")
	}

	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		rr.logKV(
			types.ErrorLevel,
			"TLS load keypair failed",
			"event", "TLSCredentials",
			"result", "FAILURE",
			"cert_file", config.CertFile,
			"key_file", config.KeyFile,
			"error", err,
		)
		return nil, err
	}

	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(config.CAFile)
	if err != nil {
		rr.logKV(
			types.ErrorLevel,
			"TLS read CA failed",
			"event", "TLSCredentials",
			"result", "FAILURE",
			"ca_file", config.CAFile,
			"error", err,
		)
		return nil, err
	}
	if ok := certPool.AppendCertsFromPEM(ca); !ok {
		rr.logKV(
			types.ErrorLevel,
			"TLS append CA failed",
			"event", "TLSCredentials",
			"result", "FAILURE",
			"ca_file", config.CAFile,
		)
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

	rr.logKV(
		types.DebugLevel,
		"TLS credentials loaded",
		"event", "TLSCredentials",
		"result", "SUCCESS",
		"server_name", config.SubjectAlternativeName,
		"min_tls_version", minTLSVersion,
		"max_tls_version", maxTLSVersion,
	)

	return credentials.NewTLS(&tls.Config{
		ServerName:   config.SubjectAlternativeName,
		Certificates: []tls.Certificate{cert},
		RootCAs:      certPool,
		MinVersion:   minTLSVersion,
		MaxVersion:   maxTLSVersion,
	}), nil
}
