package websocketclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"nhooyr.io/websocket"
)

func wsTestURL(serverURL string) string {
	return "ws" + strings.TrimPrefix(serverURL, "http")
}

func newTestServer(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if errors.Is(err, syscall.EPERM) || errors.Is(err, syscall.EACCES) || strings.Contains(err.Error(), "operation not permitted") {
			t.Skipf("network listen not permitted: %v", err)
		}
		t.Fatalf("listen: %v", err)
	}
	ts := httptest.NewUnstartedServer(handler)
	ts.Listener = ln
	ts.Start()
	return ts
}

func TestWebSocketClientAdapter_ServeWriter(t *testing.T) {
	received := make(chan string, 1)
	ts := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		_, payload, err := conn.Read(context.Background())
		if err == nil {
			received <- string(payload)
		}
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	adapter := NewWebSocketClientAdapter[string](ctx,
		WithURL[string](wsTestURL(ts.URL)),
		WithMessageFormat[string](types.WebSocketFormatText),
	)

	in := make(chan string, 1)
	in <- "hello"
	close(in)

	if err := adapter.ServeWriter(ctx, in); err != nil {
		t.Fatalf("ServeWriter: %v", err)
	}

	select {
	case got := <-received:
		if got != "hello" {
			t.Fatalf("unexpected payload: %s", got)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for server")
	}
}

func TestWebSocketClientAdapter_Serve(t *testing.T) {
	ts := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		_ = conn.Write(context.Background(), websocket.MessageText, []byte("ping"))
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	adapter := NewWebSocketClientAdapter[string](ctx,
		WithURL[string](wsTestURL(ts.URL)),
		WithMessageFormat[string](types.WebSocketFormatText),
	)

	gotCh := make(chan string, 1)
	submit := func(_ context.Context, msg string) error {
		gotCh <- msg
		return nil
	}

	if err := adapter.Serve(ctx, submit); err != nil {
		t.Fatalf("Serve: %v", err)
	}

	select {
	case got := <-gotCh:
		if got != "ping" {
			t.Fatalf("unexpected payload: %s", got)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for submit")
	}
}

func TestWebSocketClientAdapter_ServeDuplex(t *testing.T) {
	ts := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		_, payload, err := conn.Read(context.Background())
		if err == nil {
			_ = conn.Write(context.Background(), websocket.MessageText, payload)
		}
		_ = conn.Close(websocket.StatusNormalClosure, "done")
	}))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	adapter := NewWebSocketClientAdapter[string](ctx,
		WithURL[string](wsTestURL(ts.URL)),
		WithMessageFormat[string](types.WebSocketFormatText),
	)

	in := make(chan string, 1)
	in <- "echo"
	close(in)

	gotCh := make(chan string, 1)
	submit := func(_ context.Context, msg string) error {
		gotCh <- msg
		return nil
	}

	if err := adapter.ServeDuplex(ctx, in, submit); err != nil {
		t.Fatalf("ServeDuplex: %v", err)
	}

	select {
	case got := <-gotCh:
		if got != "echo" {
			t.Fatalf("unexpected payload: %s", got)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for submit")
	}
}
