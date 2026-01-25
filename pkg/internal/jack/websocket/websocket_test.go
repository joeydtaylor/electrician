package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"nhooyr.io/websocket"
)

func newTestServer[T any](t *testing.T, srv *serverAdapter[T], submit func(context.Context, T) error) (*httptest.Server, string) {
	t.Helper()
	cfg := srv.snapshotConfig()
	h := srv.buildHandler(context.Background(), cfg, submit)
	ts := httptest.NewServer(h)
	wsURL := "ws" + strings.TrimPrefix(ts.URL, "http") + cfg.endpoint
	return ts, wsURL
}

func TestWebSocketServer_SubmitJSON(t *testing.T) {
	type payload struct {
		ID int `json:"id"`
	}

	srv := NewWebSocketServer[payload](context.Background(), WithEndpoint[payload]("/ws")).(*serverAdapter[payload])
	gotCh := make(chan payload, 1)
	submit := func(ctx context.Context, msg payload) error {
		gotCh <- msg
		return nil
	}

	ts, url := newTestServer(t, srv, submit)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	payloadBytes, err := json.Marshal(payload{ID: 42})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if err := conn.Write(ctx, websocket.MessageText, payloadBytes); err != nil {
		t.Fatalf("write: %v", err)
	}

	select {
	case got := <-gotCh:
		if got.ID != 42 {
			t.Fatalf("unexpected payload: %+v", got)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for submit")
	}
}

func TestWebSocketServer_BroadcastText(t *testing.T) {
	srv := NewWebSocketServer[string](
		context.Background(),
		WithEndpoint[string]("/ws"),
		WithMessageFormat[string](types.WebSocketFormatText),
	).(*serverAdapter[string])

	ts, url := newTestServer(t, srv, func(context.Context, string) error { return nil })
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if err := srv.Broadcast(ctx, "hello"); err != nil {
		t.Fatalf("broadcast: %v", err)
	}

	_, payload, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(payload) != "hello" {
		t.Fatalf("unexpected payload: %s", payload)
	}
}

func TestWebSocketServer_AuthStaticHeaders(t *testing.T) {
	srv := NewWebSocketServer[string](context.Background(), WithEndpoint[string]("/ws")).(*serverAdapter[string])
	srv.SetStaticHeaders(map[string]string{"x-auth": "ok"})

	ts, url := newTestServer(t, srv, func(context.Context, string) error { return nil })
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, resp, err := websocket.Dial(ctx, url, nil)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatal("expected auth error")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %v", resp)
	}
}

func TestWebSocketServer_TokenQueryParamAuth(t *testing.T) {
	srv := NewWebSocketServer[string](context.Background(), WithEndpoint[string]("/ws")).(*serverAdapter[string])
	srv.SetTokenQueryParam("token")
	srv.SetDynamicAuthValidator(func(ctx context.Context, headers map[string]string) error {
		if headers["authorization"] != "Bearer good" {
			return errors.New("bad token")
		}
		return nil
	})

	ts, url := newTestServer(t, srv, func(context.Context, string) error { return nil })
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url+"?token=good", nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")
}

func TestWebSocketServer_BinaryFormat(t *testing.T) {
	srv := NewWebSocketServer[[]byte](
		context.Background(),
		WithEndpoint[[]byte]("/ws"),
		WithMessageFormat[[]byte](types.WebSocketFormatBinary),
	).(*serverAdapter[[]byte])

	gotCh := make(chan []byte, 1)
	submit := func(ctx context.Context, msg []byte) error {
		gotCh <- msg
		return nil
	}

	ts, url := newTestServer(t, srv, submit)
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close(websocket.StatusNormalClosure, "test done")

	if err := conn.Write(ctx, websocket.MessageBinary, []byte("bin")); err != nil {
		t.Fatalf("write: %v", err)
	}

	select {
	case got := <-gotCh:
		if string(got) != "bin" {
			t.Fatalf("unexpected payload: %s", got)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for submit")
	}
}

func TestWebSocketServer_InvalidFormat(t *testing.T) {
	srv := NewWebSocketServer[string](
		context.Background(),
		WithEndpoint[string]("/ws"),
		WithMessageFormat[string]("nope"),
	).(*serverAdapter[string])

	err := srv.Serve(context.Background(), func(context.Context, string) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "unsupported websocket format") {
		t.Fatalf("expected format error, got %v", err)
	}
}
