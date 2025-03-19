// Package receivingrelay implements the internal mechanisms for a data relay system that focuses on receiving,
// processing, and securely transmitting data to subsequent stages within a distributed service architecture.
// This package specifically handles the low-level operations including data decompression, secure communication
// setup using TLS, and efficient data streaming through gRPC.

// The primary functionalities include decompressing incoming data payloads that have been compressed using
// various standard algorithms like gzip, snappy, zstd, brotli, and lz4. This ensures that the system can handle
// diverse data formats and compression techniques used in modern distributed systems.

// Additionally, this package takes responsibility for setting up and managing TLS configurations for secure
// data transmission. This includes loading X509 key pairs and CA certificates, configuring server names, and
// ensuring that all transmitted data adheres to specified security protocols.

// The combination of gRPC for networking and advanced compression and security techniques makes this package
// a critical component of the receiving relay's infrastructure, ensuring data integrity and confidentiality
// while facilitating high-throughput data processing capabilities.

// The internal.go file contains the detailed implementations of these functionalities, focusing on:
// - Data decompression based on the specified algorithms.
// - Dynamic loading of TLS credentials based on configuration.
// - Preparation and management of server options for secure and reliable data reception.

// These utilities are designed to be robust and efficient, making them suitable for high-performance
// scenarios typical in microservices and cloud-native environments where data security and efficiency are paramount.

package receivingrelay

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"os"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"google.golang.org/grpc/credentials"
)

// decompressData takes a byte slice and a compression algorithm identifier, and returns a decompressed
// byte buffer. This function is vital for processing data that has been compressed according to various
// algorithms such as gzip, snappy, zstd, brotli, and lz4. Each compression type is handled according
// to its specific library requirements and capabilities.
//
// Parameters:
//   - data: The compressed data as a byte slice.
//   - algorithm: The compression algorithm used, as defined by relay.CompressionAlgorithm.
//
// Returns:
//   - *bytes.Buffer: The decompressed data.
//   - error: Error encountered during decompression, if any.
func decompressData(data []byte, algorithm relay.CompressionAlgorithm) (*bytes.Buffer, error) {
	var b bytes.Buffer
	var r io.Reader

	switch algorithm {
	case COMPRESS_DEFLATE:
		var err error
		r, err = gzip.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	case COMPRESS_SNAPPY:
		r = snappy.NewReader(bytes.NewReader(data))
	case COMPRESS_ZSTD:
		var err error
		r, err = zstd.NewReader(bytes.NewReader(data))
		if err != nil {
			return nil, err
		}
	case COMPRESS_BROTLI:
		r = brotli.NewReader(bytes.NewReader(data))
	case COMPRESS_LZ4:
		r = lz4.NewReader(bytes.NewReader(data))
	default:
		r = bytes.NewReader(data) // No compression
	}

	if _, err := io.Copy(&b, r); err != nil {
		return nil, err
	}
	return &b, nil
}

// loadTLSCredentials loads the TLS credentials for a ReceivingRelay from the provided TLSConfig.
// This method is critical for setting up secure communications using TLS. It includes detailed error
// handling and logging, which are crucial for diagnosing issues related to TLS configuration and
// certificate management. The method loads a X509 key pair and CA certificates from the file system,
// and prepares a tls.Config with these certificates.
//
// If TLS is not enabled in the configuration, the method logs a warning and returns an error indicating
// that TLS is disabled. This method assumes that TLS credentials (certificate and key files) are necessary
// for the operation of the ReceivingRelay and are properly formatted and accessible.
//
// Parameters:
//   - config: A pointer to a types.TLSConfig containing the necessary TLS settings such as certificate
//     and key file locations, and the expected server name for the certificate.
//
// Returns:
//   - credentials.TransportCredentials: The loaded transport credentials on success.
//   - error: An error object detailing what went wrong during loading, if anything.
func (rr *ReceivingRelay[T]) loadTLSCredentials(config *types.TLSConfig) (credentials.TransportCredentials, error) {
	if config.UseTLS {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			rr.NotifyLoggers(
				types.ErrorLevel,
				"Component: %s, address: %s, event: loadTLSCredentials, error: %v => Failed to load key pair",
				rr.componentMetadata, rr.Address, err,
			)
			return nil, err
		}

		certPool := x509.NewCertPool()
		ca, err := os.ReadFile(config.CAFile)
		if err != nil {
			rr.NotifyLoggers(
				types.ErrorLevel,
				"Component: %s, address: %s, event: loadTLSCredentials, error: %v => Failed to read CA file",
				rr.componentMetadata, rr.Address, err,
			)
			return nil, err
		}
		if ok := certPool.AppendCertsFromPEM(ca); !ok {
			rr.NotifyLoggers(
				types.ErrorLevel,
				"Component: %s, address: %s, event: loadTLSCredentials, error: %v => Failed to append CA certificate",
				rr.componentMetadata, rr.Address, err,
			)
			return nil, fmt.Errorf("failed to append CA certificate")
		}

		// Set Min and Max TLS versions (defaulting to TLS 1.2 - 1.3 if unspecified)
		minTLSVersion := config.MinTLSVersion
		if minTLSVersion == 0 {
			minTLSVersion = tls.VersionTLS12
		}

		maxTLSVersion := config.MaxTLSVersion
		if maxTLSVersion == 0 {
			maxTLSVersion = tls.VersionTLS13
		}

		return credentials.NewTLS(&tls.Config{
			ServerName:   config.SubjectAlternativeName, // Ensure this matches the certificate name
			Certificates: []tls.Certificate{cert},
			RootCAs:      certPool,
			MinVersion:   minTLSVersion,
			MaxVersion:   maxTLSVersion,
		}), nil
	} else {
		rr.NotifyLoggers(
			types.WarnLevel,
			"Component: %s, address: %s, event: loadTLSCredentials => TLS IS DISABLED!",
			rr.componentMetadata, rr.Address,
		)
		return nil, fmt.Errorf("TLS is disabled")
	}
}
