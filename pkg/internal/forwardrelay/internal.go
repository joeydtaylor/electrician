// Package forwardrelay contains internal utilities for the ForwardRelay component
// which facilitate secure communication, data compression, and input data handling.
// These utilities ensure the ForwardRelay can securely connect to other network components,
// efficiently compress data before transmission, and manage continuous data input.
// This package includes methods to load TLS credentials for secure connections,
// compress data using various algorithms, and continuously read from input conduits.
package forwardrelay

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

const (
	COMPRESS_NONE    relay.CompressionAlgorithm = 0
	COMPRESS_DEFLATE relay.CompressionAlgorithm = 1
	COMPRESS_SNAPPY  relay.CompressionAlgorithm = 2
	COMPRESS_ZSTD    relay.CompressionAlgorithm = 3
	COMPRESS_BROTLI  relay.CompressionAlgorithm = 4
	COMPRESS_LZ4     relay.CompressionAlgorithm = 5
)

// loadTLSCredentials loads the TLS credentials from the provided configuration settings.
// This method is responsible for setting up a secure context for the ForwardRelay by loading
// X509 certificates and configuring the TLS settings based on the given TLSConfig.
// It attempts to load cached credentials and if unavailable, it reads the certificate and key files
// from the filesystem to create new credentials, caching them for future use.
//
// Parameters:
//   - config: A pointer to types.TLSConfig which includes certificate file paths and other TLS-related settings.
//
// Returns:
//   - credentials.TransportCredentials: Configured transport credentials for establishing a secure connection.
//   - error: An error object detailing issues encountered during the loading of TLS credentials.
//
// This method logs all steps, including successful loads and errors, providing traceability and debugging insights.
func (fr *ForwardRelay[T]) loadTLSCredentials(config *types.TLSConfig) (credentials.TransportCredentials, error) {
	loadedCreds := fr.tlsCredentials.Load()
	if loadedCreds != nil {
		return loadedCreds.(credentials.TransportCredentials), nil
	}

	fr.tlsCredentialsUpdate.Lock()
	defer fr.tlsCredentialsUpdate.Unlock()

	// Check again in case credentials were loaded while acquiring the lock
	loadedCreds = fr.tlsCredentials.Load()
	if loadedCreds != nil {
		return loadedCreds.(credentials.TransportCredentials), nil
	}

	// Load from file system
	cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
	if err != nil {
		fr.NotifyLoggers(types.ErrorLevel, "component: %v, level: ERROR, result: FAILURE, event: loadTLSCredentials, error: %v, ca_cert: %v => Error Loading Keys", fr.componentMetadata, err, cert)
		return nil, err
	}
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(config.CAFile)
	if err != nil {
		fr.NotifyLoggers(types.ErrorLevel, "component: %v, level: ERROR, result: FAILURE, event: loadTLSCredentials, error: %v, ca_cert: %v => Error Reading CA File", fr.componentMetadata, err, ca)
		return nil, err
	}
	if !certPool.AppendCertsFromPEM(ca) {
		fr.NotifyLoggers(types.ErrorLevel, "component: %v, level: ERROR, result: FAILURE, event: loadTLSCredentials, error: %v, ca_cert: %v => Error loading Tls Credentials", fr.componentMetadata, ca, err)
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	// Set Min and Max TLS versions (default to TLS 1.2 - 1.3 if not specified)
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
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, result: SUCCESS, event: loadTLSCredentials, credentials: %v => Loaded Tls Credentials", fr.componentMetadata, newCreds)
	return newCreds, nil
}

// readFromInput manages the continuous reading of data from a specified input channel.
// It repeatedly reads data from the given Receiver's output channel until the channel is closed or the context is cancelled.
// Each received data item is immediately submitted for further processing using the Submit method of ForwardRelay.
// This function is typically executed in a separate goroutine for each input conduit and is crucial for the real-time
// processing capabilities of the ForwardRelay.
//
// Parameters:
//   - input: The types.Receiver[T] from which to continuously read data.
//
// No return values, but the function handles all operational logging internally,
// documenting key events like data reception and errors during data forwarding.

func (fr *ForwardRelay[T]) readFromInput(input types.Receiver[T]) {
	for {
		select {
		case <-fr.ctx.Done():
			fr.NotifyLoggers(types.InfoLevel, "component: %s, level: INFO, result: SUCCESS, event: readFromInput => Context canceled, stopping read from conduit", fr.componentMetadata)
			return
		case data, ok := <-input.GetOutputChannel():
			if !ok {
				return // Channel is closed, exit the goroutine
			}
			if err := fr.Submit(fr.ctx, data); err != nil {
				fr.NotifyLoggers(types.ErrorLevel, "component: %s, level: ERROR, result: FAILURE, event: readFromInput, error: %v => Error forwarding data", fr.componentMetadata, err)
			}
		}
	}
}

// compressData applies the specified compression algorithm to the provided data.
// This function supports multiple compression algorithms such as gzip, snappy, zstd, brotli, and lz4.
// It initializes the appropriate compressor according to the specified algorithm, processes the input data,
// and returns the compressed byte slice. This utility is critical for reducing data size before network transmission
// in the ForwardRelay component.
//
// Parameters:
//   - data: A byte slice containing the data to be compressed.
//   - algorithm: The compression algorithm to use as defined by relay.CompressionAlgorithm.
//
// Returns:
//   - []byte: The compressed data.
//   - error: An error object indicating any issues encountered during the compression process.
//
// This method includes detailed error handling and logging for diagnostics and performance tracking.

func compressData(data []byte, algorithm relay.CompressionAlgorithm) ([]byte, error) {
	var b bytes.Buffer
	var w io.WriteCloser // Declare the writer variable here with appropriate interface

	switch algorithm {
	case COMPRESS_DEFLATE:
		w = gzip.NewWriter(&b)
	case COMPRESS_SNAPPY:
		w = snappy.NewBufferedWriter(&b)
	case COMPRESS_ZSTD:
		var err error
		w, err = zstd.NewWriter(&b) // You should handle this error.
		if err != nil {
			return nil, err
		}
	case COMPRESS_BROTLI:
		w = brotli.NewWriterLevel(&b, brotli.BestCompression)
	case COMPRESS_LZ4:
		w = lz4.NewWriter(&b)
	default:
		return data, nil // No compression, directly return the data.
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
