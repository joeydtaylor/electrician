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
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/golang/snappy"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/klauspost/compress/zstd"
	"github.com/pierrec/lz4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Local aliases for compression, same as forward side:
const (
	COMPRESS_NONE    relay.CompressionAlgorithm = 0
	COMPRESS_DEFLATE relay.CompressionAlgorithm = 1
	COMPRESS_SNAPPY  relay.CompressionAlgorithm = 2
	COMPRESS_ZSTD    relay.CompressionAlgorithm = 3
	COMPRESS_BROTLI  relay.CompressionAlgorithm = 4
	COMPRESS_LZ4     relay.CompressionAlgorithm = 5

	// Local aliases for encryption (must match your proto’s enum values!)
	ENCRYPTION_NONE    relay.EncryptionSuite = 0
	ENCRYPTION_AES_GCM relay.EncryptionSuite = 1
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

// decryptData checks SecurityOptions. If encryption is enabled and AES-GCM is selected,
// it extracts the nonce (the first gcm.NonceSize() bytes) and decrypts the remainder.
// Otherwise, it returns data as-is (no decryption).
func decryptData(data []byte, secOpts *relay.SecurityOptions, key string) ([]byte, error) {
	if secOpts == nil || !secOpts.Enabled || secOpts.Suite != ENCRYPTION_AES_GCM {
		// No encryption needed or not AES-GCM.
		return data, nil
	}

	// Convert key string to bytes (16, 24, or 32 bytes for AES-128/192/256).
	aesKey := []byte(key)

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return nil, fmt.Errorf("decryptData: invalid AES key: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("decryptData: failed to create GCM: %w", err)
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("decryptData: ciphertext too short for nonce")
	}

	// Split out the nonce (first nonceSize bytes) from the rest (the actual ciphertext).
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("decryptData: GCM decryption failed: %w", err)
	}
	return plaintext, nil
}

// -------------------- NEW: auth policy + interceptor wiring --------------------

// buildUnaryPolicyInterceptor enforces static headers and the dynamic auth validator for unary RPCs.
func (rr *ReceivingRelay[T]) buildUnaryPolicyInterceptor() grpc.UnaryServerInterceptor {
	needPolicy := rr.dynamicAuthValidator != nil || len(rr.staticHeaders) > 0
	if !needPolicy {
		return nil
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		mdMap := rr.collectIncomingMD(ctx)

		// Static headers
		if err := rr.checkStaticHeaders(mdMap); err != nil {
			return nil, err
		}

		// Dynamic validator
		if rr.dynamicAuthValidator != nil {
			if err := rr.dynamicAuthValidator(ctx, mdMap); err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "auth validation failed: %v", err)
			}
		}

		return handler(ctx, req)
	}
}

// buildStreamPolicyInterceptor enforces the same policy for streaming RPCs (e.g., StreamReceive).
func (rr *ReceivingRelay[T]) buildStreamPolicyInterceptor() grpc.StreamServerInterceptor {
	needPolicy := rr.dynamicAuthValidator != nil || len(rr.staticHeaders) > 0
	if !needPolicy {
		return nil
	}
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		mdMap := rr.collectIncomingMD(ctx)

		// Static headers
		if err := rr.checkStaticHeaders(mdMap); err != nil {
			return err
		}

		// Dynamic validator
		if rr.dynamicAuthValidator != nil {
			if err := rr.dynamicAuthValidator(ctx, mdMap); err != nil {
				return status.Errorf(codes.Unauthenticated, "auth validation failed: %v", err)
			}
		}

		return handler(srv, ss)
	}
}

func (rr *ReceivingRelay[T]) appendAuthServerOptions(opts []grpc.ServerOption) []grpc.ServerOption {
	// Auto-install a built-in validator if OAuth2 introspection is configured
	rr.ensureDefaultAuthValidator()

	// Unary chain: policy + user-provided (order: policy first, then custom)
	var unaryChain []grpc.UnaryServerInterceptor
	if p := rr.buildUnaryPolicyInterceptor(); p != nil {
		unaryChain = append(unaryChain, p)
	}
	if rr.authUnary != nil {
		unaryChain = append(unaryChain, rr.authUnary)
	}
	if len(unaryChain) > 0 {
		opts = append(opts, grpc.ChainUnaryInterceptor(unaryChain...))
	}

	// Stream chain
	if p := rr.buildStreamPolicyInterceptor(); p != nil {
		opts = append(opts, grpc.ChainStreamInterceptor(p))
	}
	return opts
}

// collectIncomingMD flattens incoming metadata into a case-insensitive map.
func (rr *ReceivingRelay[T]) collectIncomingMD(ctx context.Context) map[string]string {
	out := make(map[string]string)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for k, vals := range md {
			if len(vals) > 0 {
				out[strings.ToLower(k)] = vals[0]
			}
		}
	}
	return out
}

// checkStaticHeaders enforces exact-match static metadata headers.
func (rr *ReceivingRelay[T]) checkStaticHeaders(md map[string]string) error {
	for k, v := range rr.staticHeaders {
		lk := strings.ToLower(k)
		got, ok := md[lk]
		if !ok || got != v {
			return status.Errorf(codes.Unauthenticated, "missing/invalid header %s", k)
		}
	}
	return nil
}

// tokenCacheEntry holds the cached decision for a bearer token.
type tokenCacheEntry struct {
	active bool
	scope  string
	exp    time.Time
}

type cachingIntrospectionValidator struct {
	introspectionURL string
	authType         string // "basic" | "bearer" | "none"
	clientID         string
	clientSecret     string
	bearerToken      string
	requiredScopes   []string

	hc  *http.Client
	mu  sync.Mutex
	m   map[string]tokenCacheEntry
	ttl time.Duration

	// backoff on server 429 / overload
	backoffUntil time.Time
	backoffStep  time.Duration // grows up to maxBackoff
	maxBackoff   time.Duration
}

func newCachingIntrospectionValidator(o *relay.OAuth2Options) *cachingIntrospectionValidator {
	ttl := time.Duration(o.GetIntrospectionCacheSeconds()) * time.Second
	if ttl <= 0 {
		ttl = 30 * time.Second
	}
	// TLS 1.3 client that accepts local/self-signed (dev); prod users can override by wiring a custom hc later if needed.
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion:         tls.VersionTLS13,
			MaxVersion:         tls.VersionTLS13,
			InsecureSkipVerify: true, // dev only
		},
	}
	return &cachingIntrospectionValidator{
		introspectionURL: strings.TrimRight(o.GetIntrospectionUrl(), "/"),
		authType:         strings.ToLower(o.GetIntrospectionAuthType()),
		clientID:         o.GetIntrospectionClientId(),
		clientSecret:     o.GetIntrospectionClientSecret(),
		bearerToken:      o.GetIntrospectionBearerToken(),
		requiredScopes:   append([]string(nil), o.GetRequiredScopes()...),
		hc:               &http.Client{Timeout: 8 * time.Second, Transport: tr},
		m:                make(map[string]tokenCacheEntry),
		ttl:              ttl,
		backoffStep:      250 * time.Millisecond,
		maxBackoff:       5 * time.Second,
	}
}

func (v *cachingIntrospectionValidator) hasAllScopes(granted string) bool {
	if len(v.requiredScopes) == 0 {
		return true
	}
	parts := strings.Fields(granted) // space separated
	set := make(map[string]struct{}, len(parts))
	for _, s := range parts {
		set[s] = struct{}{}
	}
	for _, need := range v.requiredScopes {
		if _, ok := set[need]; !ok {
			return false
		}
	}
	return true
}

type introspectResp struct {
	Active bool   `json:"active"`
	Scope  string `json:"scope"`
}

func (v *cachingIntrospectionValidator) validate(ctx context.Context, token string) error {
	now := time.Now()

	// global backoff if auth server is overloaded
	if until := v.backoffUntil; until.After(now) {
		return errors.New("auth server backoff in effect")
	}

	// cache hit still valid?
	v.mu.Lock()
	if e, ok := v.m[token]; ok && e.exp.After(now) {
		v.mu.Unlock()
		if !e.active {
			return errors.New("token inactive")
		}
		if !v.hasAllScopes(e.scope) {
			return errors.New("insufficient scope")
		}
		return nil
	}
	v.mu.Unlock()

	// call RFC 7662
	form := url.Values{}
	form.Set("token", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.introspectionURL, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	switch v.authType {
	case "basic":
		req.SetBasicAuth(v.clientID, v.clientSecret)
	case "bearer":
		if v.bearerToken != "" {
			req.Header.Set("Authorization", "Bearer "+v.bearerToken)
		}
	case "none":
		// no auth header
	default:
		// default to basic if misconfigured but fields exist
		if v.clientID != "" || v.clientSecret != "" {
			req.SetBasicAuth(v.clientID, v.clientSecret)
		}
	}

	resp, err := v.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		// exponential-ish backoff
		v.mu.Lock()
		if v.backoffStep < v.maxBackoff {
			v.backoffStep *= 2
			if v.backoffStep > v.maxBackoff {
				v.backoffStep = v.maxBackoff
			}
		}
		v.backoffUntil = now.Add(v.backoffStep)
		v.mu.Unlock()
		return errors.New("introspection 429")
	}
	// reset backoff on success/non-429
	v.mu.Lock()
	v.backoffStep = 250 * time.Millisecond
	v.backoffUntil = time.Time{}
	v.mu.Unlock()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return errors.New(resp.Status)
	}

	var ir introspectResp
	if err := json.NewDecoder(resp.Body).Decode(&ir); err != nil {
		return err
	}

	// cache result
	v.mu.Lock()
	v.m[token] = tokenCacheEntry{
		active: ir.Active,
		scope:  ir.Scope,
		exp:    now.Add(v.ttl),
	}
	v.mu.Unlock()

	if !ir.Active {
		return errors.New("token inactive")
	}
	if !v.hasAllScopes(ir.Scope) {
		return errors.New("insufficient scope")
	}
	return nil
}

// ensureDefaultAuthValidator installs a built-in introspection validator
// when rr.dynamicAuthValidator is nil and OAuth2 introspection is configured.
func (rr *ReceivingRelay[T]) ensureDefaultAuthValidator() {
	if rr.dynamicAuthValidator != nil {
		return
	}
	if rr.authOptions == nil || rr.authOptions.Mode != relay.AuthMode_AUTH_OAUTH2 {
		return
	}
	o := rr.authOptions.GetOauth2()
	if o == nil || !o.GetAcceptIntrospection() || o.GetIntrospectionUrl() == "" {
		return
	}
	validator := newCachingIntrospectionValidator(o)
	rr.dynamicAuthValidator = func(ctx context.Context, md map[string]string) error {
		// Find bearer in metadata
		var token string
		if v, ok := md["authorization"]; ok && strings.HasPrefix(strings.ToLower(v), "bearer ") {
			token = strings.TrimSpace(v[len("bearer "):])
		}
		if token == "" {
			return errors.New("missing bearer token")
		}
		return validator.validate(ctx, token)
	}
	rr.NotifyLoggers(types.InfoLevel, "component: %s, address: %s, level: INFO, event: Auth => installed built-in OAuth2 introspection validator", rr.componentMetadata, rr.Address)
}
