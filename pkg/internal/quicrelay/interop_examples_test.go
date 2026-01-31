package quicrelay

import (
	"context"
	"crypto/tls"
	"encoding/hex"
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

func exampleTLSConfigs(t *testing.T) (*types.TLSConfig, *types.TLSConfig) {
	t.Helper()
	base := filepath.Join("..", "..", "..", "example", "relay_example", "tls")
	serverCert := filepath.Join(base, "server.crt")
	serverKey := filepath.Join(base, "server.key")
	clientCert := filepath.Join(base, "client.crt")
	clientKey := filepath.Join(base, "client.key")
	ca := filepath.Join(base, "ca.crt")

	if _, err := os.Stat(serverCert); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(serverKey); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(clientCert); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(clientKey); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(ca); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}

	serverTLS := &types.TLSConfig{
		UseTLS:                 true,
		CertFile:               serverCert,
		KeyFile:                serverKey,
		CAFile:                 ca,
		SubjectAlternativeName: "localhost",
		MinTLSVersion:          tls.VersionTLS13,
		MaxTLSVersion:          tls.VersionTLS13,
	}
	clientTLS := &types.TLSConfig{
		UseTLS:                 true,
		CertFile:               clientCert,
		KeyFile:                clientKey,
		CAFile:                 ca,
		SubjectAlternativeName: "localhost",
		MinTLSVersion:          tls.VersionTLS13,
		MaxTLSVersion:          tls.VersionTLS13,
	}
	return serverTLS, clientTLS
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

func TestSecureQUICExampleInterop(t *testing.T) {
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

	serverTLS, clientTLS := exampleTLSConfigs(t)
	addr := freeUDPAddr(t)

	jwtOpts := &relay.OAuth2Options{
		AcceptJwt:        true,
		Issuer:           exampleIssuer,
		JwksUri:          authSrv.JWKSURL,
		RequiredAudience: []string{exampleAudience},
		RequiredScopes:   []string{exampleScope},
		JwksCacheSeconds: 300,
	}
	recvAuth := &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_OAUTH2,
		Oauth2:  jwtOpts,
	}

	recv := NewReceivingRelay[exampleFeedback](
		ctx,
		ReceivingRelayWithAddress[exampleFeedback](addr),
		ReceivingRelayWithTLSConfig[exampleFeedback](serverTLS),
		ReceivingRelayWithDecryptionKey[exampleFeedback](mustKeyFromHex(t)),
		ReceivingRelayWithAuthenticationOptions[exampleFeedback](recvAuth),
		ReceivingRelayWithStaticHeaders[exampleFeedback](map[string]string{"x-tenant": "local"}),
		ReceivingRelayWithAuthRequired[exampleFeedback](true),
	)
	if err := recv.Start(ctx); err != nil {
		t.Fatalf("receiver start: %v", err)
	}
	defer recv.Stop()

	perf := &relay.PerformanceOptions{
		UseCompression:       true,
		CompressionAlgorithm: forwardrelay.COMPRESS_SNAPPY,
	}
	sec := &relay.SecurityOptions{
		Enabled: true,
		Suite:   forwardrelay.ENCRYPT_AES_GCM,
	}
	fwdAuth := &relay.AuthenticationOptions{
		Enabled: true,
		Mode:    relay.AuthMode_AUTH_OAUTH2,
		Oauth2:  jwtOpts,
	}
	tokenSource := forwardrelay.NewRefreshingClientCredentialsSource(
		authSrv.URL,
		exampleClientID,
		exampleClientSecret,
		[]string{exampleScope},
		20*time.Second,
		oauthHTTPClient(),
	)

	fr := NewForwardRelay[exampleFeedback](
		ctx,
		ForwardRelayWithTarget[exampleFeedback](addr),
		ForwardRelayWithTLSConfig[exampleFeedback](clientTLS),
		ForwardRelayWithPerformanceOptions[exampleFeedback](perf),
		ForwardRelayWithSecurityOptions[exampleFeedback](sec, mustKeyFromHex(t)),
		ForwardRelayWithAuthenticationOptions[exampleFeedback](fwdAuth),
		ForwardRelayWithOAuth2[exampleFeedback](tokenSource),
		ForwardRelayWithStaticHeaders[exampleFeedback](map[string]string{"x-tenant": "local"}),
		ForwardRelayWithAuthRequired[exampleFeedback](true),
	)

	clientTLSCfg, err := fr.buildClientTLSConfig()
	if err != nil {
		t.Fatalf("client tls: %v", err)
	}
	waitForReady(t, addr, clientTLSCfg)

	in := exampleFeedback{
		CustomerID: "cust-001",
		Content:    "quic secure message",
		Category:   "feedback",
		IsNegative: false,
		Tags:       []string{"quic", "secure"},
	}
	if err := fr.Submit(ctx, in); err != nil {
		t.Fatalf("submit: %v", err)
	}

	out := waitForMessage(t, recv.DataCh)
	if out.CustomerID != in.CustomerID || out.Content != in.Content || out.IsNegative != in.IsNegative {
		t.Fatalf("mismatch: got %+v want %+v", out, in)
	}
}
