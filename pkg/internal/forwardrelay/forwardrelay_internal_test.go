package forwardrelay

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc/metadata"
)

type tokenSourceFunc func(context.Context) (string, error)

func (f tokenSourceFunc) AccessToken(ctx context.Context) (string, error) { return f(ctx) }

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func jsonResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

func tlsFixture(t *testing.T) *types.TLSConfig {
	t.Helper()

	base := filepath.Join("..", "..", "..", "example", "relay_example", "tls")
	cert := filepath.Join(base, "server.crt")
	key := filepath.Join(base, "server.key")
	ca := filepath.Join(base, "ca.crt")

	if _, err := os.Stat(cert); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(key); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}
	if _, err := os.Stat(ca); err != nil {
		t.Skipf("tls fixture not available: %v", err)
	}

	return &types.TLSConfig{
		UseTLS:                 true,
		CertFile:               cert,
		KeyFile:                key,
		CAFile:                 ca,
		SubjectAlternativeName: "localhost",
	}
}

func TestPassthroughPayloadValue(t *testing.T) {
	fr := &ForwardRelay[relay.WrappedPayload]{}
	wp := relay.WrappedPayload{Id: "id-1"}

	got, err := fr.asPassthroughPayload(wp)
	if err != nil {
		t.Fatalf("asPassthroughPayload error: %v", err)
	}
	if got.GetId() != "id-1" {
		t.Fatalf("unexpected payload id: %s", got.GetId())
	}
}

func TestPassthroughPayloadPointer(t *testing.T) {
	fr := &ForwardRelay[*relay.WrappedPayload]{}
	wp := &relay.WrappedPayload{Id: "id-2"}

	got, err := fr.asPassthroughPayload(wp)
	if err != nil {
		t.Fatalf("asPassthroughPayload error: %v", err)
	}
	if got != wp {
		t.Fatalf("expected same pointer")
	}
}

func TestPassthroughPayloadInvalidType(t *testing.T) {
	fr := &ForwardRelay[string]{}
	_, err := fr.asPassthroughPayload("nope")
	if err == nil {
		t.Fatalf("expected passthrough type error")
	}
}

func TestBuildPerRPCContextHeaders(t *testing.T) {
	fr := &ForwardRelay[string]{authRequired: true}
	fr.SetStaticHeaders(map[string]string{
		"x-static":   "static",
		"x-override": "static",
		"":           "skip",
	})
	fr.SetDynamicHeaders(func(context.Context) map[string]string {
		return map[string]string{
			"x-dynamic":  "dynamic",
			"x-override": "dynamic",
			"x-empty":    "",
		}
	})

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("trace-id", "existing", "x-pre", "pre"))
	out, err := fr.buildPerRPCContext(ctx)
	if err != nil {
		t.Fatalf("buildPerRPCContext error: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(out)
	if !ok {
		t.Fatalf("expected metadata in outgoing context")
	}
	if got := md.Get("x-pre"); len(got) != 1 || got[0] != "pre" {
		t.Fatalf("expected x-pre to be preserved, got %v", got)
	}
	if got := md.Get("x-static"); len(got) != 1 || got[0] != "static" {
		t.Fatalf("expected x-static header, got %v", got)
	}
	if got := md.Get("x-dynamic"); len(got) != 1 || got[0] != "dynamic" {
		t.Fatalf("expected x-dynamic header, got %v", got)
	}
	if got := md.Get("x-override"); len(got) != 1 || got[0] != "dynamic" {
		t.Fatalf("expected x-override to be dynamic, got %v", got)
	}
	if got := md.Get("trace-id"); len(got) != 1 || got[0] != "existing" {
		t.Fatalf("expected trace-id to be preserved, got %v", got)
	}
	if got := md.Get("x-empty"); len(got) != 0 {
		t.Fatalf("expected empty header to be skipped, got %v", got)
	}
	if _, ok := md[""]; ok {
		t.Fatalf("expected empty key to be skipped")
	}
}

func TestBuildPerRPCContextAddsTraceID(t *testing.T) {
	fr := &ForwardRelay[string]{authRequired: true}

	out, err := fr.buildPerRPCContext(context.Background())
	if err != nil {
		t.Fatalf("buildPerRPCContext error: %v", err)
	}

	md, ok := metadata.FromOutgoingContext(out)
	if !ok {
		t.Fatalf("expected metadata in outgoing context")
	}
	if got := md.Get("trace-id"); len(got) != 1 || got[0] == "" {
		t.Fatalf("expected trace-id to be populated, got %v", got)
	}
}

func TestBuildPerRPCContextTokenRequired(t *testing.T) {
	fr := &ForwardRelay[string]{
		authRequired: true,
		tokenSource: tokenSourceFunc(func(context.Context) (string, error) {
			return "", nil
		}),
	}

	if _, err := fr.buildPerRPCContext(context.Background()); err == nil {
		t.Fatalf("expected error for empty oauth token when required")
	}
}

func TestBuildPerRPCContextTokenOptional(t *testing.T) {
	fr := &ForwardRelay[string]{
		authRequired: false,
		tokenSource: tokenSourceFunc(func(context.Context) (string, error) {
			return "", errors.New("boom")
		}),
	}

	out, err := fr.buildPerRPCContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md, _ := metadata.FromOutgoingContext(out)
	if got := md.Get("authorization"); len(got) != 0 {
		t.Fatalf("expected no authorization header, got %v", got)
	}
}

func TestBuildPerRPCContextTokenSuccess(t *testing.T) {
	fr := &ForwardRelay[string]{
		authRequired: true,
		tokenSource: tokenSourceFunc(func(context.Context) (string, error) {
			return "token-123", nil
		}),
	}

	out, err := fr.buildPerRPCContext(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	md, _ := metadata.FromOutgoingContext(out)
	if got := md.Get("authorization"); len(got) != 1 || got[0] != "Bearer token-123" {
		t.Fatalf("expected authorization header, got %v", got)
	}
}

func TestMakeDialOptionsTLS(t *testing.T) {
	fr := &ForwardRelay[string]{TlsConfig: tlsFixture(t)}

	creds, insecure, err := fr.makeDialOptions()
	if err != nil {
		t.Fatalf("makeDialOptions error: %v", err)
	}
	if insecure {
		t.Fatalf("expected secure dial options")
	}
	if len(creds) != 1 {
		t.Fatalf("expected 1 credential, got %d", len(creds))
	}
}

func TestMakeDialOptionsOAuthRequiresTLS(t *testing.T) {
	fr := &ForwardRelay[string]{
		authRequired: true,
		tokenSource:  staticTokenSource{tok: "token"},
	}

	if _, _, err := fr.makeDialOptions(); err == nil {
		t.Fatalf("expected error when oauth is required without TLS")
	}
}

func TestMakeDialOptionsInsecureOK(t *testing.T) {
	fr := &ForwardRelay[string]{authRequired: false}

	creds, insecure, err := fr.makeDialOptions()
	if err != nil {
		t.Fatalf("makeDialOptions error: %v", err)
	}
	if !insecure {
		t.Fatalf("expected insecure dial options")
	}
	if len(creds) != 0 {
		t.Fatalf("expected no credentials, got %d", len(creds))
	}
}

func TestNormalizeAESKey(t *testing.T) {
	raw := []byte("0123456789abcdef")

	got, err := normalizeAESKey(string(raw))
	if err != nil {
		t.Fatalf("normalizeAESKey error: %v", err)
	}
	if string(got) != string(raw) {
		t.Fatalf("expected raw key, got %x", got)
	}

	rawHex := []byte("0123456789abcdef01234567")
	hexKey := hex.EncodeToString(rawHex)
	got, err = normalizeAESKey(hexKey)
	if err != nil {
		t.Fatalf("normalizeAESKey hex error: %v", err)
	}
	if string(got) != string(rawHex) {
		t.Fatalf("expected hex decoded key, got %x", got)
	}

	rawB64 := []byte("0123456789abcdef0123456789abcdef")
	b64Key := base64.StdEncoding.EncodeToString(rawB64)
	got, err = normalizeAESKey(b64Key)
	if err != nil {
		t.Fatalf("normalizeAESKey base64 error: %v", err)
	}
	if string(got) != string(rawB64) {
		t.Fatalf("expected base64 decoded key, got %x", got)
	}

	if _, err := normalizeAESKey("short"); err == nil {
		t.Fatalf("expected error for invalid key length")
	}
}

func TestMergeOAuth2Options(t *testing.T) {
	dst := &relay.OAuth2Options{
		Issuer:                "old",
		ForwardBearerToken:    true,
		ForwardMetadataKey:    "old",
		RequiredAudience:      []string{"old"},
		IntrospectionAuthType: "basic",
	}
	src := &relay.OAuth2Options{
		AcceptJwt:                 true,
		AcceptIntrospection:       true,
		Issuer:                    "new",
		JwksUri:                   "jwks",
		RequiredAudience:          []string{"new"},
		RequiredScopes:            []string{"scope"},
		IntrospectionUrl:          "https://introspect",
		IntrospectionClientId:     "client",
		IntrospectionClientSecret: "secret",
		ForwardMetadataKey:        "auth",
		JwksCacheSeconds:          30,
	}

	out := MergeOAuth2Options(dst, src)
	if out != dst {
		t.Fatalf("expected dst to be returned")
	}
	if dst.Issuer != "new" || dst.JwksUri != "jwks" || !dst.AcceptJwt || !dst.AcceptIntrospection {
		t.Fatalf("unexpected merge result: %+v", dst)
	}
	if dst.ForwardMetadataKey != "auth" || dst.ForwardBearerToken {
		t.Fatalf("expected forwarding settings to be overwritten, got %+v", dst)
	}
	src.RequiredAudience[0] = "mutated"
	if dst.RequiredAudience[0] != "new" {
		t.Fatalf("expected RequiredAudience to be cloned")
	}
	src.RequiredScopes[0] = "mutated"
	if dst.RequiredScopes[0] != "scope" {
		t.Fatalf("expected RequiredScopes to be cloned")
	}
}

func TestMergeOAuth2OptionsNilDst(t *testing.T) {
	src := &relay.OAuth2Options{Issuer: "issuer"}
	out := MergeOAuth2Options(nil, src)
	if out == src {
		t.Fatalf("expected clone when dst is nil")
	}
	src.Issuer = "mutated"
	if out.Issuer != "issuer" {
		t.Fatalf("expected cloned issuer to remain unchanged")
	}
}

func TestAuthenticationOptionsClones(t *testing.T) {
	principals := []string{"a", "b"}
	opts := NewAuthenticationOptionsMTLS(principals, "trust")
	principals[0] = "mutated"

	if got := opts.GetMtls().AllowedPrincipals[0]; got != "a" {
		t.Fatalf("expected principals to be cloned, got %q", got)
	}
}

func TestOAuth2OptionsClones(t *testing.T) {
	aud := []string{"aud"}
	scopes := []string{"scope"}

	opts := NewOAuth2JWTOptions("issuer", "jwks", aud, scopes, 15)
	aud[0] = "mutated"
	scopes[0] = "mutated"

	if got := opts.GetRequiredAudience()[0]; got != "aud" {
		t.Fatalf("expected audience to be cloned, got %q", got)
	}
	if got := opts.GetRequiredScopes()[0]; got != "scope" {
		t.Fatalf("expected scopes to be cloned, got %q", got)
	}
}

func TestTokenSources(t *testing.T) {
	t.Run("static", func(t *testing.T) {
		ts := NewStaticBearerTokenSource("token")
		got, err := ts.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken error: %v", err)
		}
		if got != "token" {
			t.Fatalf("expected token, got %q", got)
		}
	})

	t.Run("env", func(t *testing.T) {
		t.Setenv("TOKEN_ENV", "env-token")
		ts := NewEnvBearerTokenSource("TOKEN_ENV")
		got, err := ts.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken error: %v", err)
		}
		if got != "env-token" {
			t.Fatalf("expected env token, got %q", got)
		}
	})

	t.Run("client-credentials-cache", func(t *testing.T) {
		var hits int32
		errCh := make(chan error, 1)

		client := &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				atomic.AddInt32(&hits, 1)
				if r.URL.Path != "/api/auth/oauth/token" {
					errCh <- fmt.Errorf("unexpected path: %s", r.URL.Path)
					return jsonResponse(http.StatusNotFound, `{"error":"not found"}`), nil
				}
				if r.Method != http.MethodPost {
					errCh <- fmt.Errorf("unexpected method: %s", r.Method)
				}
				body, _ := io.ReadAll(r.Body)
				values, _ := url.ParseQuery(string(body))
				if values.Get("grant_type") != "client_credentials" {
					errCh <- fmt.Errorf("unexpected grant_type: %s", values.Get("grant_type"))
				}
				if values.Get("scope") != "a b" {
					errCh <- fmt.Errorf("unexpected scope: %s", values.Get("scope"))
				}
				user, pass, ok := r.BasicAuth()
				if !ok || user != "id" || pass != "secret" {
					errCh <- fmt.Errorf("unexpected basic auth")
				}
				return jsonResponse(http.StatusOK, `{"access_token":"tok1","expires_in":3600}`), nil
			}),
		}

		ts := NewClientCredentialsBearerTokenSource("https://auth.local", "id", "secret", []string{"a", "b"}, client)
		first, err := ts.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken error: %v", err)
		}
		second, err := ts.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken error: %v", err)
		}
		if first != "tok1" || second != "tok1" {
			t.Fatalf("unexpected token values: %q %q", first, second)
		}
		if got := atomic.LoadInt32(&hits); got != 1 {
			t.Fatalf("expected 1 token request, got %d", got)
		}
		select {
		case err := <-errCh:
			t.Fatal(err)
		default:
		}
	})

	t.Run("client-credentials-refresh", func(t *testing.T) {
		var hits int32
		client := &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				count := atomic.AddInt32(&hits, 1)
				token := fmt.Sprintf("tok-%d", count)
				return jsonResponse(http.StatusOK, `{"access_token":"`+token+`","expires_in":1}`), nil
			}),
		}

		ts := NewRefreshingClientCredentialsSource("https://auth.local", "id", "secret", nil, 2*time.Second, client)
		first, err := ts.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken error: %v", err)
		}
		second, err := ts.AccessToken(context.Background())
		if err != nil {
			t.Fatalf("AccessToken error: %v", err)
		}
		if first == "" || second == "" {
			t.Fatalf("expected tokens to be returned")
		}
		if got := atomic.LoadInt32(&hits); got < 2 {
			t.Fatalf("expected refresh to hit token endpoint twice, got %d", got)
		}
	})

	t.Run("client-credentials-non-2xx", func(t *testing.T) {
		client := &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return jsonResponse(http.StatusInternalServerError, `{"error":"nope"}`), nil
			}),
		}

		ts := NewClientCredentialsBearerTokenSource("https://auth.local", "id", "secret", nil, client)
		if _, err := ts.AccessToken(context.Background()); err == nil {
			t.Fatalf("expected error for non-2xx response")
		}
	})

	t.Run("client-credentials-empty-token", func(t *testing.T) {
		client := &http.Client{
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return jsonResponse(http.StatusOK, `{"access_token":"","expires_in":3600}`), nil
			}),
		}

		ts := NewClientCredentialsBearerTokenSource("https://auth.local", "id", "secret", nil, client)
		if _, err := ts.AccessToken(context.Background()); err == nil {
			t.Fatalf("expected error for empty access_token")
		}
	})
}

func TestSetStaticHeadersCopiesMap(t *testing.T) {
	fr := &ForwardRelay[string]{authRequired: true}
	headers := map[string]string{"x-copy": "a"}
	fr.SetStaticHeaders(headers)
	headers["x-copy"] = "b"

	out, err := fr.buildPerRPCContext(context.Background())
	if err != nil {
		t.Fatalf("buildPerRPCContext error: %v", err)
	}
	md, _ := metadata.FromOutgoingContext(out)
	if got := md.Get("x-copy"); len(got) != 1 || got[0] != "a" {
		t.Fatalf("expected copied header to remain, got %v", got)
	}
}

func TestSetDynamicHeadersNil(t *testing.T) {
	fr := &ForwardRelay[string]{authRequired: true}
	fr.SetDynamicHeaders(nil)

	out, err := fr.buildPerRPCContext(context.Background())
	if err != nil {
		t.Fatalf("buildPerRPCContext error: %v", err)
	}
	md, _ := metadata.FromOutgoingContext(out)
	if len(md) == 0 {
		t.Fatalf("expected trace-id to be present")
	}
}
