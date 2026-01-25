package receivingrelay

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/forwardrelay"
	"github.com/joeydtaylor/electrician/pkg/internal/relay"
	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

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

func encryptAESGCM(t *testing.T, key string, plaintext []byte) []byte {
	t.Helper()

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		t.Fatalf("aes.NewCipher error: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("cipher.NewGCM error: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		t.Fatalf("nonce read error: %v", err)
	}
	return gcm.Seal(nonce, nonce, plaintext, nil)
}

type stubSubmitter[T any] struct {
	ch      chan T
	started int32
	meta    types.ComponentMetadata
}

func newStubSubmitter[T any]() *stubSubmitter[T] {
	return &stubSubmitter[T]{
		ch: make(chan T, 4),
		meta: types.ComponentMetadata{
			ID:   "stub",
			Type: "STUB_SUBMITTER",
		},
	}
}

func (s *stubSubmitter[T]) Submit(_ context.Context, elem T) error {
	s.ch <- elem
	return nil
}
func (s *stubSubmitter[T]) ConnectGenerator(...types.Generator[T]) {}
func (s *stubSubmitter[T]) GetGenerators() []types.Generator[T]     { return nil }
func (s *stubSubmitter[T]) ConnectLogger(...types.Logger)           {}
func (s *stubSubmitter[T]) Restart(context.Context) error           { return nil }
func (s *stubSubmitter[T]) NotifyLoggers(types.LogLevel, string, ...interface{}) {
}
func (s *stubSubmitter[T]) GetComponentMetadata() types.ComponentMetadata { return s.meta }
func (s *stubSubmitter[T]) SetComponentMetadata(name string, id string) {
	s.meta.Name = name
	s.meta.ID = id
}
func (s *stubSubmitter[T]) Stop() error { return nil }
func (s *stubSubmitter[T]) IsStarted() bool {
	return atomic.LoadInt32(&s.started) == 1
}
func (s *stubSubmitter[T]) Start(context.Context) error {
	atomic.StoreInt32(&s.started, 1)
	return nil
}

type stubStream struct {
	ctx context.Context
}

func (s *stubStream) Context() context.Context                        { return s.ctx }
func (s *stubStream) SendHeader(metadata.MD) error                    { return nil }
func (s *stubStream) SetHeader(metadata.MD) error                     { return nil }
func (s *stubStream) SetTrailer(metadata.MD)                          {}
func (s *stubStream) SendMsg(interface{}) error                       { return nil }
func (s *stubStream) RecvMsg(interface{}) error                       { return nil }

func TestUnwrapPayloadNil(t *testing.T) {
	var out string
	if err := UnwrapPayload(nil, "", &out); err == nil {
		t.Fatalf("expected error for nil wrapped payload")
	}
}

func TestUnwrapPayloadUnsupportedEncoding(t *testing.T) {
	wp := &relay.WrappedPayload{
		Payload:         []byte("ignored"),
		PayloadEncoding: relay.PayloadEncoding(99),
	}
	var out string
	if err := UnwrapPayload(wp, "", &out); err == nil {
		t.Fatalf("expected error for unsupported payload encoding")
	}
}

func TestUnwrapPayloadGobRoundTrip(t *testing.T) {
	in := map[string]int{"a": 1}
	wp, err := forwardrelay.WrapPayload(in, nil, nil, "")
	if err != nil {
		t.Fatalf("WrapPayload error: %v", err)
	}

	var out map[string]int
	if err := UnwrapPayload(wp, "", &out); err != nil {
		t.Fatalf("UnwrapPayload error: %v", err)
	}
	if out["a"] != 1 {
		t.Fatalf("unexpected gob decode result: %+v", out)
	}
}

func TestUnwrapPayloadProtoRoundTrip(t *testing.T) {
	ack := &relay.StreamAcknowledgment{
		Success:  true,
		Message:  "ok",
		StreamId: "s1",
		Seq:      9,
	}
	b, err := proto.Marshal(ack)
	if err != nil {
		t.Fatalf("proto.Marshal error: %v", err)
	}

	wp := &relay.WrappedPayload{
		Payload:         b,
		PayloadEncoding: relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO,
		PayloadType:     string(ack.ProtoReflect().Descriptor().FullName()),
	}

	var out relay.StreamAcknowledgment
	if err := UnwrapPayload(wp, "", &out); err != nil {
		t.Fatalf("UnwrapPayload error: %v", err)
	}
	if out.GetStreamId() != "s1" || !out.GetSuccess() {
		t.Fatalf("unexpected proto decode result: %+v", out)
	}
}

func TestUnwrapPayloadProtoPointerRoundTrip(t *testing.T) {
	ack := &relay.StreamAcknowledgment{Success: true, Message: "ok"}
	b, err := proto.Marshal(ack)
	if err != nil {
		t.Fatalf("proto.Marshal error: %v", err)
	}

	wp := &relay.WrappedPayload{
		Payload:         b,
		PayloadEncoding: relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO,
		PayloadType:     string(ack.ProtoReflect().Descriptor().FullName()),
	}

	var out *relay.StreamAcknowledgment
	if err := UnwrapPayload(wp, "", &out); err != nil {
		t.Fatalf("UnwrapPayload error: %v", err)
	}
	if out == nil || !out.GetSuccess() {
		t.Fatalf("unexpected proto decode result: %+v", out)
	}
}

func TestUnwrapPayloadProtoTypeMismatch(t *testing.T) {
	ack := &relay.StreamAcknowledgment{Success: true}
	b, err := proto.Marshal(ack)
	if err != nil {
		t.Fatalf("proto.Marshal error: %v", err)
	}

	wp := &relay.WrappedPayload{
		Payload:         b,
		PayloadEncoding: relay.PayloadEncoding_PAYLOAD_ENCODING_PROTO,
		PayloadType:     "wrong.Type",
	}

	var out relay.StreamAcknowledgment
	if err := UnwrapPayload(wp, "", &out); err == nil {
		t.Fatalf("expected payload_type mismatch error")
	}
}

func TestUnwrapPayloadJSONSuffix(t *testing.T) {
	wp := &relay.WrappedPayload{
		Payload: []byte(`{"name":"delta"}`),
		Metadata: &relay.MessageMetadata{
			ContentType: "application/vnd.test+json",
		},
	}

	var out struct {
		Name string `json:"name"`
	}
	if err := UnwrapPayload(wp, "", &out); err != nil {
		t.Fatalf("UnwrapPayload error: %v", err)
	}
	if out.Name != "delta" {
		t.Fatalf("unexpected json decode result: %+v", out)
	}
}

func TestDecryptDataRoundTrip(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	plaintext := []byte("hello")
	ciphertext := encryptAESGCM(t, key, plaintext)

	out, err := decryptData(ciphertext, &relay.SecurityOptions{Enabled: true, Suite: ENCRYPTION_AES_GCM}, key)
	if err != nil {
		t.Fatalf("decryptData error: %v", err)
	}
	if !bytes.Equal(out, plaintext) {
		t.Fatalf("decryptData mismatch: got %q", out)
	}
}

func TestDecryptDataInvalidKey(t *testing.T) {
	_, err := decryptData([]byte("cipher"), &relay.SecurityOptions{Enabled: true, Suite: ENCRYPTION_AES_GCM}, "short")
	if err == nil {
		t.Fatalf("expected invalid key error")
	}
}

func TestDecryptDataShortCiphertext(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	_, err := decryptData([]byte("short"), &relay.SecurityOptions{Enabled: true, Suite: ENCRYPTION_AES_GCM}, key)
	if err == nil {
		t.Fatalf("expected ciphertext length error")
	}
}

func TestDecompressDataInvalidGzip(t *testing.T) {
	if _, err := decompressData([]byte("nope"), COMPRESS_DEFLATE); err == nil {
		t.Fatalf("expected decompression error")
	}
}

func TestTLSCredentials(t *testing.T) {
	rr := &ReceivingRelay[string]{}
	if _, err := rr.loadTLSCredentials(&types.TLSConfig{UseTLS: false}); err == nil {
		t.Fatalf("expected error when TLS disabled")
	}

	cfg := tlsFixture(t)
	if _, err := rr.loadTLSCredentials(cfg); err != nil {
		t.Fatalf("loadTLSCredentials error: %v", err)
	}
}

func TestMergeOAuth2Options(t *testing.T) {
	dst := &relay.OAuth2Options{
		Issuer:             "old",
		ForwardBearerToken: true,
		ForwardMetadataKey: "old",
		RequiredAudience:   []string{"old"},
	}
	src := &relay.OAuth2Options{
		AcceptJwt:                true,
		AcceptIntrospection:      true,
		Issuer:                   "new",
		JwksUri:                  "jwks",
		RequiredAudience:         []string{"new"},
		RequiredScopes:           []string{"scope"},
		IntrospectionUrl:         "https://introspect",
		IntrospectionAuthType:    "basic",
		IntrospectionClientId:    "client",
		IntrospectionClientSecret: "secret",
		ForwardMetadataKey:       "auth",
		JwksCacheSeconds:         30,
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

func TestAuthOptionsClone(t *testing.T) {
	principals := []string{"a", "b"}
	opts := NewAuthenticationOptionsMTLS(principals, "trust")
	principals[0] = "mutated"
	if got := opts.GetMtls().AllowedPrincipals[0]; got != "a" {
		t.Fatalf("expected principals to be cloned, got %q", got)
	}
}

func TestOAuth2OptionsClone(t *testing.T) {
	aud := []string{"aud"}
	scopes := []string{"scope"}
	opts := NewOAuth2JWTOptions("issuer", "jwks", aud, scopes, 10)
	aud[0] = "mutated"
	scopes[0] = "mutated"
	if got := opts.GetRequiredAudience()[0]; got != "aud" {
		t.Fatalf("expected audience to be cloned, got %q", got)
	}
	if got := opts.GetRequiredScopes()[0]; got != "scope" {
		t.Fatalf("expected scopes to be cloned, got %q", got)
	}
}

func TestCollectIncomingMDLowercase(t *testing.T) {
	rr := &ReceivingRelay[string]{}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("X-Token", "abc", "trace-id", "tid"))
	md := rr.collectIncomingMD(ctx)
	if md["x-token"] != "abc" || md["trace-id"] != "tid" {
		t.Fatalf("unexpected md map: %+v", md)
	}
}

func TestStaticHeadersCopy(t *testing.T) {
	rr := &ReceivingRelay[string]{}
	headers := map[string]string{"X-Key": "a"}
	rr.SetStaticHeaders(headers)
	headers["X-Key"] = "b"

	if err := rr.checkStaticHeaders(map[string]string{"x-key": "a"}); err != nil {
		t.Fatalf("expected copied header to remain, got %v", err)
	}
}

func TestPolicyInterceptorStaticHeaders(t *testing.T) {
	rr := &ReceivingRelay[string]{authRequired: true}
	rr.staticHeaders = map[string]string{"x-key": "ok"}
	interceptor := rr.buildUnaryPolicyInterceptor()
	if interceptor == nil {
		t.Fatalf("expected interceptor to be built")
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return "ok", nil }
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-key", "ok"))
	if _, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctxBad := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-key", "bad"))
	if _, err := interceptor(ctxBad, nil, &grpc.UnaryServerInfo{}, handler); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}

	rr.authRequired = false
	if _, err := interceptor(ctxBad, nil, &grpc.UnaryServerInfo{}, handler); err != nil {
		t.Fatalf("expected soft-fail, got %v", err)
	}
}

func TestPolicyInterceptorDynamicValidator(t *testing.T) {
	rr := &ReceivingRelay[string]{authRequired: true}
	rr.dynamicAuthValidator = func(context.Context, map[string]string) error {
		return errors.New("boom")
	}
	interceptor := rr.buildUnaryPolicyInterceptor()

	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return "ok", nil }
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-key", "ok"))
	if _, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}

	rr.authRequired = false
	if _, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{}, handler); err != nil {
		t.Fatalf("expected soft-fail, got %v", err)
	}
}

func TestStreamPolicyInterceptor(t *testing.T) {
	rr := &ReceivingRelay[string]{authRequired: true}
	rr.staticHeaders = map[string]string{"x-key": "ok"}
	interceptor := rr.buildStreamPolicyInterceptor()

	handler := func(_ interface{}, _ grpc.ServerStream) error { return nil }
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-key", "ok"))
	if err := interceptor(nil, &stubStream{ctx: ctx}, &grpc.StreamServerInfo{}, handler); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctxBad := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-key", "bad"))
	if err := interceptor(nil, &stubStream{ctx: ctxBad}, &grpc.StreamServerInfo{}, handler); status.Code(err) != codes.Unauthenticated {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}

	rr.authRequired = false
	if err := interceptor(nil, &stubStream{ctx: ctxBad}, &grpc.StreamServerInfo{}, handler); err != nil {
		t.Fatalf("expected soft-fail, got %v", err)
	}
}

func TestStartOutputFanout(t *testing.T) {
	rr := &ReceivingRelay[string]{DataCh: make(chan string, 1)}
	sub := newStubSubmitter[string]()
	rr.Outputs = []types.Submitter[string]{sub}

	rr.startOutputFanout()

	rr.DataCh <- "payload"
	select {
	case got := <-sub.ch:
		if got != "payload" {
			t.Fatalf("unexpected payload: %v", got)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("timeout waiting for submit")
	}
	close(rr.DataCh)
}

func TestEnsureDefaultAuthValidator(t *testing.T) {
	rr := &ReceivingRelay[string]{}
	rr.authOptions = NewAuthenticationOptionsOAuth2(&relay.OAuth2Options{
		AcceptIntrospection: true,
		IntrospectionUrl:    "https://auth.local/introspect",
	})

	rr.ensureDefaultAuthValidator()
	if rr.dynamicAuthValidator == nil {
		t.Fatalf("expected default auth validator to be installed")
	}
}

func TestIntrospectionValidatorCache(t *testing.T) {
	opts := &relay.OAuth2Options{
		AcceptIntrospection:   true,
		IntrospectionUrl:      "https://auth.local/introspect",
		IntrospectionAuthType: "basic",
		RequiredScopes:        []string{"scope-a"},
		IntrospectionCacheSeconds: 30,
	}
	v := newCachingIntrospectionValidator(opts)

	var hits int32
	client := &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			atomic.AddInt32(&hits, 1)
			if r.Method != http.MethodPost {
				return nil, fmt.Errorf("unexpected method: %s", r.Method)
			}
			return jsonResponse(http.StatusOK, `{"active":true,"scope":"scope-a scope-b"}`), nil
		}),
	}
	v.hc = client

	if err := v.validate(context.Background(), "token"); err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if err := v.validate(context.Background(), "token"); err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Fatalf("expected cache hit, got %d", got)
	}
}

func TestIntrospectionValidatorScopeError(t *testing.T) {
	opts := &relay.OAuth2Options{
		AcceptIntrospection:   true,
		IntrospectionUrl:      "https://auth.local/introspect",
		IntrospectionAuthType: "basic",
		RequiredScopes:        []string{"scope-a"},
		IntrospectionCacheSeconds: 30,
	}
	v := newCachingIntrospectionValidator(opts)
	v.hc = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusOK, `{"active":true,"scope":"scope-b"}`), nil
		}),
	}

	if err := v.validate(context.Background(), "token"); err == nil {
		t.Fatalf("expected scope error")
	}
}

func TestIntrospectionValidatorBackoff(t *testing.T) {
	opts := &relay.OAuth2Options{
		AcceptIntrospection:   true,
		IntrospectionUrl:      "https://auth.local/introspect",
		IntrospectionAuthType: "basic",
		IntrospectionCacheSeconds: 30,
	}
	v := newCachingIntrospectionValidator(opts)

	v.hc = &http.Client{
		Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
			return jsonResponse(http.StatusTooManyRequests, `{"error":"rate"}`), nil
		}),
	}

	if err := v.validate(context.Background(), "token"); err == nil {
		t.Fatalf("expected 429 error")
	}
	if v.backoffUntil.IsZero() {
		t.Fatalf("expected backoff to be set")
	}

	if err := v.validate(context.Background(), "token"); err == nil {
		t.Fatalf("expected backoff error")
	}
}
