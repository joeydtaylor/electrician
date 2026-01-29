package quicrelay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"github.com/quic-go/quic-go"
)

type feedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

type testTLS struct {
	server  *types.TLSConfig
	client  *types.TLSConfig
	cleanup func()
}

func newTestTLS(t *testing.T) testTLS {
	t.Helper()
	dir := t.TempDir()

	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("ca key: %v", err)
	}
	caTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "quic-test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	caDER, err := x509.CreateCertificate(rand.Reader, caTmpl, caTmpl, &caKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("ca cert: %v", err)
	}

	serverKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("server key: %v", err)
	}
	serverTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(2),
		Subject:               pkix.Name{CommonName: "localhost"},
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	serverDER, err := x509.CreateCertificate(rand.Reader, serverTmpl, caTmpl, &serverKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("server cert: %v", err)
	}

	caPath := filepath.Join(dir, "ca.crt")
	serverCertPath := filepath.Join(dir, "server.crt")
	serverKeyPath := filepath.Join(dir, "server.key")

	if err := writePEM(caPath, "CERTIFICATE", caDER); err != nil {
		t.Fatalf("write ca: %v", err)
	}
	if err := writePEM(serverCertPath, "CERTIFICATE", serverDER); err != nil {
		t.Fatalf("write server cert: %v", err)
	}
	if err := writePEM(serverKeyPath, "RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(serverKey)); err != nil {
		t.Fatalf("write server key: %v", err)
	}

	serverCfg := &types.TLSConfig{
		UseTLS:                 true,
		CertFile:               serverCertPath,
		KeyFile:                serverKeyPath,
		CAFile:                 caPath,
		SubjectAlternativeName: "localhost",
		MinTLSVersion:          tls.VersionTLS13,
		MaxTLSVersion:          tls.VersionTLS13,
	}
	clientCfg := &types.TLSConfig{
		UseTLS:                 true,
		CAFile:                 caPath,
		SubjectAlternativeName: "localhost",
		MinTLSVersion:          tls.VersionTLS13,
		MaxTLSVersion:          tls.VersionTLS13,
	}

	cleanup := func() {
		_ = os.Remove(caPath)
		_ = os.Remove(serverCertPath)
		_ = os.Remove(serverKeyPath)
	}

	return testTLS{server: serverCfg, client: clientCfg, cleanup: cleanup}
}

func writePEM(path, typ string, der []byte) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, &pem.Block{Type: typ, Bytes: der})
}

func freeUDPAddr(t *testing.T) string {
	t.Helper()
	l, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) || strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("udp listen not permitted: %v", err)
		}
		t.Fatalf("listen udp: %v", err)
	}
	addr := l.LocalAddr().(*net.UDPAddr)
	_ = l.Close()
	return fmt.Sprintf("127.0.0.1:%d", addr.Port)
}

func waitForReady(t *testing.T, addr string, tlsCfg *tls.Config) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		conn, err := quic.DialAddr(ctx, addr, tlsCfg, &quic.Config{})
		cancel()
		if err == nil {
			_ = conn.CloseWithError(0, "ready")
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for QUIC listener: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func waitForMessage[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for message")
	}
	var zero T
	return zero
}

func TestQuicRelayRoundTripGob(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[feedback](
		ctx,
		ReceivingRelayWithAddress[feedback](addr),
		ReceivingRelayWithTLSConfig[feedback](tlsFiles.server),
	)

	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	fr := NewForwardRelay[feedback](
		ctx,
		ForwardRelayWithTarget[feedback](addr),
		ForwardRelayWithTLSConfig[feedback](tlsFiles.client),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	in := feedback{CustomerID: "cust-1", Content: "hello", Category: "feedback", IsNegative: false, Tags: []string{"quic"}}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}

	out := waitForMessage(t, recv.DataCh)
	if out.CustomerID != in.CustomerID || out.Content != in.Content || out.IsNegative != in.IsNegative {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}

func TestQuicRelayRoundTripJSONPassthrough(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[feedback](
		ctx,
		ReceivingRelayWithAddress[feedback](addr),
		ReceivingRelayWithTLSConfig[feedback](tlsFiles.server),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	fr := NewForwardRelay[*relay.WrappedPayload](
		ctx,
		ForwardRelayWithTarget[*relay.WrappedPayload](addr),
		ForwardRelayWithTLSConfig[*relay.WrappedPayload](tlsFiles.client),
		ForwardRelayWithPassthrough[*relay.WrappedPayload](true),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	payload := []byte(`{"customerId":"cust-json","content":"json","category":"feedback","isNegative":true,"tags":["json"]}`)
	wp := &relay.WrappedPayload{
		Id:      "json-1",
		Payload: payload,
		Metadata: &relay.MessageMetadata{
			ContentType: "application/json",
		},
	}

	if err := fr.Submit(ctx, wp); err != nil {
		t.Fatalf("submit: %v", err)
	}

	out := waitForMessage(t, recv.DataCh)
	if out.CustomerID != "cust-json" || !out.IsNegative {
		t.Fatalf("unexpected output: %+v", out)
	}
}

func TestQuicRelayPreservesTraceIDDefaults(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[*relay.WrappedPayload](
		ctx,
		ReceivingRelayWithAddress[*relay.WrappedPayload](addr),
		ReceivingRelayWithTLSConfig[*relay.WrappedPayload](tlsFiles.server),
		ReceivingRelayWithPassthrough[*relay.WrappedPayload](true),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	const traceID = "trace-quic-default"
	fr := NewForwardRelay[*relay.WrappedPayload](
		ctx,
		ForwardRelayWithTarget[*relay.WrappedPayload](addr),
		ForwardRelayWithTLSConfig[*relay.WrappedPayload](tlsFiles.client),
		ForwardRelayWithPassthrough[*relay.WrappedPayload](true),
		ForwardRelayWithStaticHeaders[*relay.WrappedPayload](map[string]string{"trace-id": traceID}),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	wp := &relay.WrappedPayload{
		Id:      "trace-1",
		Payload: []byte("trace"),
	}
	if err := fr.Submit(ctx, wp); err != nil {
		t.Fatalf("submit: %v", err)
	}

	out := waitForMessage(t, recv.DataCh)
	if out.GetMetadata() == nil {
		t.Fatalf("expected metadata from defaults")
	}
	if got := out.GetMetadata().GetHeaders()["trace-id"]; got != traceID {
		t.Fatalf("expected trace-id %q, got %q", traceID, got)
	}
}

func TestQuicRelayEncryptionAESGCM(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	key := "0123456789abcdef0123456789abcdef"
	sec := &relay.SecurityOptions{Enabled: true, Suite: ENCRYPTION_AES_GCM}

	recv := NewReceivingRelay[feedback](
		ctx,
		ReceivingRelayWithAddress[feedback](addr),
		ReceivingRelayWithTLSConfig[feedback](tlsFiles.server),
		ReceivingRelayWithDecryptionKey[feedback](key),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	fr := NewForwardRelay[feedback](
		ctx,
		ForwardRelayWithTarget[feedback](addr),
		ForwardRelayWithTLSConfig[feedback](tlsFiles.client),
		ForwardRelayWithSecurityOptions[feedback](sec, key),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	in := feedback{CustomerID: "cust-enc", Content: "secure", IsNegative: true}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}

	out := waitForMessage(t, recv.DataCh)
	if out.CustomerID != in.CustomerID || out.Content != in.Content || out.IsNegative != in.IsNegative {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}

func TestQuicRelayStaticHeadersRequired(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[feedback](
		ctx,
		ReceivingRelayWithAddress[feedback](addr),
		ReceivingRelayWithTLSConfig[feedback](tlsFiles.server),
		ReceivingRelayWithStaticHeaders[feedback](map[string]string{"x-tenant": "local"}),
		ReceivingRelayWithAuthRequired[feedback](true),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	fr := NewForwardRelay[feedback](
		ctx,
		ForwardRelayWithTarget[feedback](addr),
		ForwardRelayWithTLSConfig[feedback](tlsFiles.client),
		ForwardRelayWithStaticHeaders[feedback](map[string]string{"x-tenant": "local"}),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	in := feedback{CustomerID: "cust-1", Content: "ok"}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}
	_ = waitForMessage(t, recv.DataCh)
}

func TestQuicRelayStaticHeadersSoftFail(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[feedback](
		ctx,
		ReceivingRelayWithAddress[feedback](addr),
		ReceivingRelayWithTLSConfig[feedback](tlsFiles.server),
		ReceivingRelayWithStaticHeaders[feedback](map[string]string{"x-tenant": "local"}),
		ReceivingRelayWithAuthRequired[feedback](false),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	fr := NewForwardRelay[feedback](
		ctx,
		ForwardRelayWithTarget[feedback](addr),
		ForwardRelayWithTLSConfig[feedback](tlsFiles.client),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	in := feedback{CustomerID: "cust-soft", Content: "soft"}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}
	_ = waitForMessage(t, recv.DataCh)
}

func TestQuicRelayAckModes(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[feedback](
		ctx,
		ReceivingRelayWithAddress[feedback](addr),
		ReceivingRelayWithTLSConfig[feedback](tlsFiles.server),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	fr := NewForwardRelay[*relay.WrappedPayload](
		ctx,
		ForwardRelayWithTarget[*relay.WrappedPayload](addr),
		ForwardRelayWithTLSConfig[*relay.WrappedPayload](tlsFiles.client),
		ForwardRelayWithPassthrough[*relay.WrappedPayload](true),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	conn, err := quic.DialAddr(ctx, addr, clientTLS, &quic.Config{})
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.CloseWithError(0, "done")

	stream, err := conn.OpenStreamSync(ctx)
	if err != nil {
		t.Fatalf("open stream: %v", err)
	}

	open := &relay.StreamOpen{StreamId: "ack-test", AckMode: relay.AckMode_ACK_BATCH, AckEveryN: 2, MaxInFlight: 10}
	if err := writeProtoFrame(stream, &relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Open{Open: open}}); err != nil {
		t.Fatalf("send open: %v", err)
	}

	wp1, err := forwardrelay.WrapPayload(feedback{CustomerID: "a"}, nil, nil, "")
	if err != nil {
		t.Fatalf("wrap1: %v", err)
	}
	wp2, err := forwardrelay.WrapPayload(feedback{CustomerID: "b"}, nil, nil, "")
	if err != nil {
		t.Fatalf("wrap2: %v", err)
	}
	wp1.Seq = 1
	wp2.Seq = 2

	if err := writeProtoFrame(stream, &relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Payload{Payload: wp1}}); err != nil {
		t.Fatalf("send payload1: %v", err)
	}
	if err := writeProtoFrame(stream, &relay.RelayEnvelope{Msg: &relay.RelayEnvelope_Payload{Payload: wp2}}); err != nil {
		t.Fatalf("send payload2: %v", err)
	}

	ackOpen := &relay.StreamAcknowledgment{}
	if err := readProtoFrame(stream, ackOpen, defaultMaxFrameBytes); err != nil {
		t.Fatalf("read open ack: %v", err)
	}
	if !ackOpen.Success {
		t.Fatalf("expected open ack success")
	}

	ackBatch := &relay.StreamAcknowledgment{}
	if err := readProtoFrame(stream, ackBatch, defaultMaxFrameBytes); err != nil {
		t.Fatalf("read batch ack: %v", err)
	}
	if ackBatch.GetOkCount() != 2 {
		t.Fatalf("expected ok_count=2, got %d", ackBatch.GetOkCount())
	}
}

func TestQuicRelayConcurrentSend(t *testing.T) {
	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	addr := freeUDPAddr(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recv := NewReceivingRelay[feedback](
		ctx,
		ReceivingRelayWithAddress[feedback](addr),
		ReceivingRelayWithTLSConfig[feedback](tlsFiles.server),
		ReceivingRelayWithBufferSize[feedback](256),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	fr := NewForwardRelay[feedback](
		ctx,
		ForwardRelayWithTarget[feedback](addr),
		ForwardRelayWithTLSConfig[feedback](tlsFiles.client),
	)

	clientTLS, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLS)

	const total = 50
	var wg sync.WaitGroup
	wg.Add(total)
	for i := 0; i < total; i++ {
		i := i
		go func() {
			defer wg.Done()
			_ = fr.Submit(ctx, feedback{CustomerID: fmt.Sprintf("c-%d", i), Content: "hi"})
		}()
	}
	wg.Wait()

	seen := make(map[string]struct{})
	deadline := time.After(2 * time.Second)
	for len(seen) < total {
		select {
		case v := <-recv.DataCh:
			seen[v.CustomerID] = struct{}{}
		case <-deadline:
			t.Fatalf("timeout waiting for messages: got %d", len(seen))
		}
	}
}
