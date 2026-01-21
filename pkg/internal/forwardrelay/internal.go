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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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

type streamSession struct {
	target string

	conn   *grpc.ClientConn
	client relay.RelayServiceClient
	stream relay.RelayService_StreamReceiveClient

	sendCh chan *relay.RelayEnvelope
	doneCh chan struct{}
}

func (fr *ForwardRelay[T]) ensureStreamsInit() {
	fr.streamsMu.Lock()
	defer fr.streamsMu.Unlock()
	if fr.streams == nil {
		fr.streams = make(map[string]*streamSession, len(fr.Targets))
	}
	if fr.streamSendBuf <= 0 {
		fr.streamSendBuf = 8192
	}
}

func (fr *ForwardRelay[T]) getOrCreateStreamSession(ctx context.Context, address string) (*streamSession, error) {
	fr.ensureStreamsInit()

	// Fast path: existing session
	fr.streamsMu.Lock()
	s := fr.streams[address]
	fr.streamsMu.Unlock()

	if s != nil {
		// If session is still alive, reuse it.
		select {
		case <-s.doneCh:
			// dead, rebuild
		default:
			return s, nil
		}

		// Session is dead; remove it so we don't hand it out again.
		fr.streamsMu.Lock()
		if fr.streams[address] == s {
			delete(fr.streams, address)
		}
		fr.streamsMu.Unlock()
	}

	// Build new session
	ns, err := fr.openStreamSession(ctx, address)
	if err != nil {
		return nil, err
	}

	// Publish (handle races)
	fr.streamsMu.Lock()
	if existing := fr.streams[address]; existing != nil {
		// Someone else created one; keep theirs and close ours.
		fr.streamsMu.Unlock()
		_ = ns.close("race lost")
		return existing, nil
	}
	fr.streams[address] = ns
	fr.streamsMu.Unlock()

	return ns, nil
}

func (fr *ForwardRelay[T]) openStreamSession(ctx context.Context, address string) (*streamSession, error) {
	// Build outgoing context ONCE for the stream (auth + static headers + dynamic headers + trace-id)
	outCtx, err := fr.buildPerRPCContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("metadata build failed: %w", err)
	}

	// Extract trace-id for stream defaults (optional)
	traceID := ""
	if md, ok := metadata.FromOutgoingContext(outCtx); ok && len(md["trace-id"]) > 0 {
		traceID = md["trace-id"][0]
	}
	if traceID == "" {
		traceID = utils.GenerateUniqueHash()
	}

	// Dial options
	creds, useInsecure, err := fr.makeDialOptions()
	if err != nil {
		return nil, fmt.Errorf("dial options: %w", err)
	}
	var dialOpts []grpc.DialOption
	if !useInsecure {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(creds[0]))
	} else {
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	conn, err := grpc.DialContext(outCtx, address, dialOpts...)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	client := relay.NewRelayServiceClient(conn)
	st, err := client.StreamReceive(outCtx)
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("stream open failed: %w", err)
	}

	s := &streamSession{
		target: address,
		conn:   conn,
		client: client,
		stream: st,
		sendCh: make(chan *relay.RelayEnvelope, fr.streamSendBuf),
		doneCh: make(chan struct{}),
	}

	// Send StreamOpen once (uses oneof wrapper type)
	open := &relay.StreamOpen{
		StreamId: utils.GenerateUniqueHash(),
		Defaults: &relay.MessageMetadata{
			Headers: map[string]string{
				"source": "go",
			},
			ContentType: "application/octet-stream",
			Version: &relay.VersionInfo{
				Major: 1,
				Minor: 0,
			},
			Performance: fr.PerformanceOptions,
			Security:    fr.SecurityOptions,
			TraceId:     traceID,
		},
		AckMode:             relay.AckMode_ACK_BATCH,
		AckEveryN:           1024,
		MaxInFlight:         uint32(fr.streamSendBuf),
		OmitPayloadMetadata: true,
	}

	openEnv := &relay.RelayEnvelope{
		Msg: &relay.RelayEnvelope_Open{Open: open},
	}

	if err := st.Send(openEnv); err != nil {
		_ = st.CloseSend()
		_ = conn.Close()
		return nil, fmt.Errorf("send open failed: %w", err)
	}

	// Sender loop (serializes stream.Send)
	go fr.streamSendLoop(s)

	// Ack reader (prevents server from blocking on flow control)
	go fr.streamAckLoop(s)

	fr.NotifyLoggers(types.InfoLevel,
		"component: %v, level: INFO, event: stream => opened StreamReceive to %s stream_id=%s trace_id=%s",
		fr.componentMetadata, address, open.StreamId, traceID,
	)

	return s, nil
}

func (fr *ForwardRelay[T]) streamSendLoop(s *streamSession) {
	defer close(s.doneCh)

	for env := range s.sendCh {
		if err := s.stream.Send(env); err != nil {
			// Expected shutdown / peer close cases:
			// - EOF: peer closed stream
			// - Canceled: local ctx canceled
			// - Unavailable: connection torn down during shutdown
			code := status.Code(err)
			if err == io.EOF || code == codes.Canceled || code == codes.Unavailable {
				fr.NotifyLoggers(types.DebugLevel,
					"component: %v, level: DEBUG, event: stream => send ended target=%s code=%v err=%v",
					fr.componentMetadata, s.target, code, err,
				)
			} else {
				fr.NotifyLoggers(types.ErrorLevel,
					"component: %v, level: ERROR, event: stream => send failed target=%s code=%v err=%v",
					fr.componentMetadata, s.target, code, err,
				)
			}

			_ = s.stream.CloseSend()
			_ = s.conn.Close()
			return
		}
	}

	_ = s.stream.CloseSend()
	_ = s.conn.Close()
}

func (fr *ForwardRelay[T]) streamAckLoop(s *streamSession) {
	for {
		ack, err := s.stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			// Common when sender closes first or connection resets
			fr.NotifyLoggers(types.DebugLevel,
				"component: %v, level: DEBUG, event: stream => ack recv ended target=%s err=%v",
				fr.componentMetadata, s.target, err,
			)
			return
		}
		// Keep this at DEBUG unless you want spam
		fr.NotifyLoggers(types.DebugLevel,
			"component: %v, level: DEBUG, event: stream => ack target=%s stream_id=%s last_seq=%d ok=%d err=%d msg=%s",
			fr.componentMetadata, s.target, ack.GetStreamId(), ack.GetLastSeq(), ack.GetOkCount(), ack.GetErrCount(), ack.GetMessage(),
		)
	}
}

func (s *streamSession) close(reason string) error {
	// If already done, nothing to do.
	select {
	case <-s.doneCh:
		return nil
	default:
	}

	// Best-effort: enqueue a StreamClose (don’t block).
	closeEnv := &relay.RelayEnvelope{
		Msg: &relay.RelayEnvelope_Close{
			Close: &relay.StreamClose{Reason: reason},
		},
	}
	select {
	case s.sendCh <- closeEnv:
	default:
		// Channel full: skip close message and rely on CloseSend/connection close.
	}

	// Closing sendCh makes streamSendLoop exit, which calls CloseSend().
	close(s.sendCh)

	// Wait until sender loop completes cleanup.
	<-s.doneCh
	return nil
}

func (fr *ForwardRelay[T]) closeAllStreams(reason string) {
	fr.streamsMu.Lock()
	streams := fr.streams
	fr.streams = nil
	fr.streamsMu.Unlock()

	for _, s := range streams {
		if s != nil {
			_ = s.close(reason)
		}
	}
}
