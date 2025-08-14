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
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"os"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/joeydtaylor/electrician/pkg/internal/utils"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	COMPRESS_NONE    relay.CompressionAlgorithm = 0
	COMPRESS_DEFLATE relay.CompressionAlgorithm = 1
	COMPRESS_SNAPPY  relay.CompressionAlgorithm = 2
	COMPRESS_ZSTD    relay.CompressionAlgorithm = 3
	COMPRESS_BROTLI  relay.CompressionAlgorithm = 4
	COMPRESS_LZ4     relay.CompressionAlgorithm = 5

	// Encryption suites
	ENCRYPT_NONE    relay.EncryptionSuite = 0
	ENCRYPT_AES_GCM relay.EncryptionSuite = 1
)

// ---------------- TLS credentials cache ----------------

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
		fr.NotifyLoggers(types.ErrorLevel, "component: %v, level: ERROR, event: loadTLSCredentials => load keypair failed: %v", fr.componentMetadata, err)
		return nil, err
	}
	certPool := x509.NewCertPool()
	ca, err := os.ReadFile(config.CAFile)
	if err != nil {
		fr.NotifyLoggers(types.ErrorLevel, "component: %v, level: ERROR, event: loadTLSCredentials => read CA failed: %v", fr.componentMetadata, err)
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
	fr.NotifyLoggers(types.DebugLevel, "component: %v, level: DEBUG, event: loadTLSCredentials => loaded", fr.componentMetadata)
	return newCreds, nil
}

// ---------------- Input pump ----------------

func (fr *ForwardRelay[T]) readFromInput(input types.Receiver[T]) {
	for {
		select {
		case <-fr.ctx.Done():
			fr.NotifyLoggers(types.InfoLevel, "component: %v, level: INFO, event: readFromInput => context canceled", fr.componentMetadata)
			return
		case data, ok := <-input.GetOutputChannel():
			if !ok {
				return
			}
			if err := fr.Submit(fr.ctx, data); err != nil {
				fr.NotifyLoggers(types.ErrorLevel, "component: %v, level: ERROR, event: readFromInput => submit failed: %v", fr.componentMetadata, err)
			}
		}
	}
}

// ---------------- Compression ----------------

func compressData(data []byte, algorithm relay.CompressionAlgorithm) ([]byte, error) {
	var b bytes.Buffer
	var w io.WriteCloser

	switch algorithm {
	case COMPRESS_DEFLATE:
		w = gzip.NewWriter(&b)
	case COMPRESS_SNAPPY:
		w = snappy.NewBufferedWriter(&b)
	case COMPRESS_ZSTD:
		var err error
		w, err = zstd.NewWriter(&b)
		if err != nil {
			return nil, err
		}
	case COMPRESS_BROTLI:
		w = brotli.NewWriterLevel(&b, brotli.BestCompression)
	case COMPRESS_LZ4:
		w = lz4.NewWriter(&b)
	default:
		return data, nil
	}

	if _, err := w.Write(data); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// ---------------- Encryption ----------------

func normalizeAESKey(s string) ([]byte, error) {
	k := []byte(s)
	switch len(k) {
	case 16, 24, 32:
		return k, nil
	default:
		if b, err := hex.DecodeString(s); err == nil && (len(b) == 16 || len(b) == 24 || len(b) == 32) {
			return b, nil
		}
		if b, err := base64.StdEncoding.DecodeString(s); err == nil && (len(b) == 16 || len(b) == 24 || len(b) == 32) {
			return b, nil
		}
		return nil, fmt.Errorf("invalid AES key length: got %d; need 16/24/32 bytes or hex/base64 of those", len(k))
	}
}

// encryptData checks SecurityOptions. If AES-GCM enabled, returns nonce||ciphertext.
func encryptData(data []byte, secOpts *relay.SecurityOptions, key string) ([]byte, error) {
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != ENCRYPT_AES_GCM {
		return data, nil
	}
	aesKey, err := normalizeAESKey(key)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

// ---------------- Per-RPC auth/metadata helpers ----------------

// buildPerRPCContext constructs outgoing metadata:
// - Authorization: Bearer <token> (when tokenSource configured; omitted if fetch fails and authRequired==false)
// - static headers (SetStaticHeaders)
// - dynamic headers (SetDynamicHeaders)
// - trace-id (generated if absent)
func (fr *ForwardRelay[T]) buildPerRPCContext(ctx context.Context) (context.Context, error) {
	// Start with existing md if any
	var md metadata.MD
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = existing.Copy()
	} else {
		md = metadata.New(nil)
	}

	// Static headers
	for k, v := range fr.staticHeaders {
		if k == "" {
			continue
		}
		md.Set(k, v)
	}

	// Dynamic headers
	if fr.dynamicHeaders != nil {
		for k, v := range fr.dynamicHeaders(ctx) {
			if k == "" || v == "" {
				continue
			}
			md.Set(k, v)
		}
	}

	// Authorization (best-effort if authRequired == false)
	if fr.tokenSource != nil {
		tok, err := fr.tokenSource.AccessToken(ctx)
		if err != nil || tok == "" {
			if fr.authRequired {
				if err == nil {
					err = fmt.Errorf("oauth token empty")
				}
				return ctx, fmt.Errorf("oauth token source error: %w", err)
			}
			// auth not required => proceed without Authorization
		} else {
			md.Set("authorization", "Bearer "+tok)
		}
	}

	// Trace ID: set only if not present
	if _, present := md["trace-id"]; !present {
		md.Set("trace-id", utils.GenerateUniqueHash())
	}

	return metadata.NewOutgoingContext(ctx, md), nil
}

// makeDialOptions enforces TLS iff OAuth2 is enabled AND authRequired==true.
// Returns (creds, useInsecure, err).
func (fr *ForwardRelay[T]) makeDialOptions() ([]credentials.TransportCredentials, bool, error) {
	// If TLS configured, use it.
	if fr.TlsConfig != nil && fr.TlsConfig.UseTLS {
		creds, err := fr.loadTLSCredentials(fr.TlsConfig)
		if err != nil {
			return nil, false, err
		}
		return []credentials.TransportCredentials{creds}, false, nil
	}

	// No TLS configured.
	// If OAuth2 is enabled AND required, refuse plaintext.
	if fr.tokenSource != nil && fr.authRequired {
		return nil, false, fmt.Errorf("oauth2 enabled and required but TLS is disabled; refuse to dial insecure")
	}

	// Dev-friendly path: no TLS; caller will use grpc.WithInsecure().
	return nil, true, nil
}
