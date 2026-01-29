//go:build webtransport

package webtransportrelay

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/receivingrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type feedback struct {
	CustomerID string `json:"customerId"`
	Content    string `json:"content"`
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
		Subject:               pkix.Name{CommonName: "wt-test-ca"},
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

	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("client key: %v", err)
	}
	clientTmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(3),
		Subject:               pkix.Name{CommonName: "client"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
	}
	clientDER, err := x509.CreateCertificate(rand.Reader, clientTmpl, caTmpl, &clientKey.PublicKey, caKey)
	if err != nil {
		t.Fatalf("client cert: %v", err)
	}

	write := func(name string, b []byte) string {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, b, 0o600); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
		return path
	}

	caPath := write("ca.pem", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: caDER}))
	serverCert := write("server.crt", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: serverDER}))
	serverKeyPath := write("server.key", pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(serverKey)}))
	clientCert := write("client.crt", pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: clientDER}))
	clientKeyPath := write("client.key", pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)}))

	server := &types.TLSConfig{UseTLS: true, CertFile: serverCert, KeyFile: serverKeyPath, CAFile: caPath, SubjectAlternativeName: "localhost", MinTLSVersion: tls.VersionTLS13, MaxTLSVersion: tls.VersionTLS13}
	client := &types.TLSConfig{UseTLS: true, CertFile: clientCert, KeyFile: clientKeyPath, CAFile: caPath, SubjectAlternativeName: "localhost", MinTLSVersion: tls.VersionTLS13, MaxTLSVersion: tls.VersionTLS13}

	return testTLS{server: server, client: client, cleanup: func() { _ = os.RemoveAll(dir) }}
}

func TestWebTransportRelayRoundTrip(t *testing.T) {
	addr := "localhost:0"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tlsFiles := newTestTLS(t)
	defer tlsFiles.cleanup()

	recv := NewReceivingRelay[feedback](ctx,
		func(rr *ReceivingRelay[feedback]) { rr.SetAddress(addr) },
		func(rr *ReceivingRelay[feedback]) { rr.SetPath("/relay") },
		func(rr *ReceivingRelay[feedback]) { rr.SetTLSConfig(tlsFiles.server) },
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	// NOTE: Full E2E requires a live WebTransport client; keep this as a compile-time check.
	_ = recv
}

func TestWebTransportWrapUnwrapEncrypted(t *testing.T) {
	ctx := context.Background()
	key := "0123456789abcdef0123456789abcdef"
	sec := &relay.SecurityOptions{Enabled: true, Suite: relay.EncryptionSuite_ENCRYPTION_AES_GCM}

	fr := NewForwardRelay[feedback](ctx, func(fr *ForwardRelay[feedback]) {
		fr.SetPayloadFormat("json")
		fr.SetSecurityOptions(sec, key)
		fr.SetOmitPayloadMetadata(false)
	})

	in := feedback{CustomerID: "cust-1", Content: "hello"}
	wp, err := fr.wrapItem(ctx, in)
	if err != nil {
		t.Fatalf("wrap: %v", err)
	}

	var out feedback
	if err := receivingrelay.UnwrapPayload(wp, key, &out); err != nil {
		t.Fatalf("unwrap: %v", err)
	}
	if out.CustomerID != in.CustomerID || out.Content != in.Content {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}
