package receivingrelay

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/testoauth"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
)

type exampleFeedback struct {
	CustomerID string   `json:"customerId"`
	Content    string   `json:"content"`
	Category   string   `json:"category,omitempty"`
	IsNegative bool     `json:"isNegative"`
	Tags       []string `json:"tags,omitempty"`
}

const (
	exampleIssuer       = "auth-service"
	exampleAudience     = "your-api"
	exampleScope        = "write:data"
	exampleClientID     = "steeze-local-cli"
	exampleClientSecret = "local-secret"
	exampleAESKeyHex    = "ea8ccb51eefcdd058b0110c4adebaf351acbf43db2ad250fdc0d4131c959dfec"
)

func mustKeyFromHex(t *testing.T) string {
	t.Helper()
	raw, err := hex.DecodeString(exampleAESKeyHex)
	if err != nil {
		t.Fatalf("bad AES key hex: %v", err)
	}
	if len(raw) != 32 {
		t.Fatalf("unexpected AES key length: %d", len(raw))
	}
	return string(raw)
}

func tlsClientFixture(t *testing.T) *types.TLSConfig {
	t.Helper()

	base := filepath.Join("..", "..", "..", "example", "relay_example", "tls")
	cert := filepath.Join(base, "client.crt")
	key := filepath.Join(base, "client.key")
	ca := filepath.Join(base, "ca.crt")

	if _, err := os.Stat(cert); err != nil {
		t.Skipf("tls client fixture not available: %v", err)
	}
	if _, err := os.Stat(key); err != nil {
		t.Skipf("tls client fixture not available: %v", err)
	}
	if _, err := os.Stat(ca); err != nil {
		t.Skipf("tls client fixture not available: %v", err)
	}

	return &types.TLSConfig{
		UseTLS:                 true,
		CertFile:               cert,
		KeyFile:                key,
		CAFile:                 ca,
		SubjectAlternativeName: "localhost",
		MinTLSVersion:          tls.VersionTLS13,
		MaxTLSVersion:          tls.VersionTLS13,
	}
}

func freeTCPAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}

func waitForTCP(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, 150*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("timeout waiting for listener: %v", err)
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func oauthHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 6 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				MinVersion:         tls.VersionTLS13,
				MaxVersion:         tls.VersionTLS13,
				InsecureSkipVerify: true,
			},
		},
	}
}

func TestSecureGRPCExampleInterop(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
	defer cancel()

	authSrv, err := testoauth.NewTLSServer(testoauth.Config{
		Issuer:       exampleIssuer,
		Audience:     exampleAudience,
		Scope:        exampleScope,
		ClientID:     exampleClientID,
		ClientSecret: exampleClientSecret,
	})
	if err != nil {
		t.Fatalf("oauth server: %v", err)
	}
	defer authSrv.Close()

	serverTLS := tlsFixture(t)
	clientTLS := tlsClientFixture(t)
	addr := freeTCPAddr(t)

	out := newStubSubmitter[exampleFeedback]()
	jwtOpts := NewOAuth2JWTOptions(
		exampleIssuer,
		authSrv.JWKSURL,
		[]string{exampleAudience},
		[]string{exampleScope},
		300,
	)
	recvAuth := NewAuthenticationOptionsOAuth2(jwtOpts)

	recv := NewReceivingRelay[exampleFeedback](
		ctx,
		WithAddress[exampleFeedback](addr),
		WithTLSConfig[exampleFeedback](serverTLS),
		WithDecryptionKey[exampleFeedback](mustKeyFromHex(t)),
		WithAuthenticationOptions[exampleFeedback](recvAuth),
		WithStaticHeaders[exampleFeedback](map[string]string{"x-tenant": "local"}),
		WithAuthRequired[exampleFeedback](true),
		WithOutput[exampleFeedback](out),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()
	waitForTCP(t, addr)

	perf := &relay.PerformanceOptions{
		UseCompression:       true,
		CompressionAlgorithm: forwardrelay.COMPRESS_SNAPPY,
	}
	sec := &relay.SecurityOptions{
		Enabled: true,
		Suite:   forwardrelay.ENCRYPT_AES_GCM,
	}
	fwdJWT := forwardrelay.NewOAuth2JWTOptions(
		exampleIssuer,
		authSrv.JWKSURL,
		[]string{exampleAudience},
		[]string{exampleScope},
		300,
	)
	fwdAuth := forwardrelay.NewAuthenticationOptionsOAuth2(fwdJWT)
	tokenSource := forwardrelay.NewRefreshingClientCredentialsSource(
		authSrv.URL,
		exampleClientID,
		exampleClientSecret,
		[]string{exampleScope},
		20*time.Second,
		oauthHTTPClient(),
	)

	fr := forwardrelay.NewForwardRelay[exampleFeedback](
		ctx,
		forwardrelay.WithTarget[exampleFeedback](addr),
		forwardrelay.WithTLSConfig[exampleFeedback](clientTLS),
		forwardrelay.WithPerformanceOptions[exampleFeedback](perf),
		forwardrelay.WithSecurityOptions[exampleFeedback](sec, mustKeyFromHex(t)),
		forwardrelay.WithAuthenticationOptions[exampleFeedback](fwdAuth),
		forwardrelay.WithOAuth2[exampleFeedback](tokenSource),
		forwardrelay.WithStaticHeaders[exampleFeedback](map[string]string{"x-tenant": "local"}),
		forwardrelay.WithAuthRequired[exampleFeedback](true),
	)

	in := exampleFeedback{
		CustomerID: "cust-001",
		Content:    "grpc secure message",
		Category:   "feedback",
		IsNegative: false,
		Tags:       []string{"grpc", "secure"},
	}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}

	select {
	case got := <-out.ch:
		if got.CustomerID != in.CustomerID || got.Content != in.Content || got.IsNegative != in.IsNegative {
			t.Fatalf("mismatch: got %+v want %+v", got, in)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for relay payload")
	}
}
