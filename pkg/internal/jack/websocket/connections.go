package websocket

import (
	"context"
	"errors"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"nhooyr.io/websocket"
)

type outboundMessage struct {
	messageType websocket.MessageType
	payload     []byte
}

type wsConn struct {
	conn *websocket.Conn
	send chan outboundMessage
	done chan struct{}

	mu        sync.Mutex
	closed    bool
	closeOnce sync.Once
}

func newWSConn(conn *websocket.Conn, buffer int) *wsConn {
	if buffer <= 0 {
		buffer = 128
	}
	return &wsConn{
		conn: conn,
		send: make(chan outboundMessage, buffer),
		done: make(chan struct{}),
	}
}

func (c *wsConn) enqueue(msg outboundMessage) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return false
	}
	select {
	case c.send <- msg:
		return true
	default:
		return false
	}
}

func (c *wsConn) close(code websocket.StatusCode, reason string) {
	c.closeOnce.Do(func() {
		c.mu.Lock()
		c.closed = true
		close(c.done)
		close(c.send)
		c.mu.Unlock()
		_ = c.conn.Close(code, reason)
	})
}

func (s *serverAdapter[T]) addConn(conn *websocket.Conn, cfg serverConfig) (*wsConn, error) {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()

	if s.conns == nil {
		s.conns = make(map[*websocket.Conn]*wsConn)
	}
	if cfg.maxConnections > 0 && len(s.conns) >= cfg.maxConnections {
		return nil, errors.New("max connections reached")
	}

	wc := newWSConn(conn, cfg.sendBuffer)
	s.conns[conn] = wc
	return wc, nil
}

func (s *serverAdapter[T]) dropConn(conn *wsConn) {
	s.connsMu.Lock()
	for k, v := range s.conns {
		if v == conn {
			delete(s.conns, k)
			break
		}
	}
	s.connsMu.Unlock()
}

func (s *serverAdapter[T]) snapshotConns() []*wsConn {
	s.connsMu.Lock()
	defer s.connsMu.Unlock()

	if len(s.conns) == 0 {
		return nil
	}

	out := make([]*wsConn, 0, len(s.conns))
	for _, conn := range s.conns {
		out = append(out, conn)
	}
	return out
}

func (s *serverAdapter[T]) connectionCount() int {
	s.connsMu.Lock()
	count := len(s.conns)
	s.connsMu.Unlock()
	return count
}

func (s *serverAdapter[T]) closeAllConnections(reason string) {
	for _, conn := range s.snapshotConns() {
		if conn == nil {
			continue
		}
		conn.close(websocket.StatusNormalClosure, reason)
	}
}

func (s *serverAdapter[T]) startOutputFanout(ctx context.Context) {
	wires := s.snapshotOutputWires()
	if len(wires) == 0 {
		return
	}
	for _, w := range wires {
		if w == nil {
			continue
		}
		out := w.GetOutputChannel()
		go func(ch <-chan T) {
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-ch:
					if !ok {
						return
					}
					if err := s.Broadcast(ctx, msg); err != nil {
						s.NotifyLoggers(
							types.WarnLevel,
							"Broadcast: output wire error",
							"component", s.componentMetadata,
							"event", "Broadcast",
							"error", err,
						)
					}
				}
			}
		}(out)
	}
}

func (s *serverAdapter[T]) snapshotOutputWires() []types.Wire[T] {
	s.configLock.Lock()
	defer s.configLock.Unlock()

	if len(s.outputWires) == 0 {
		return nil
	}

	out := make([]types.Wire[T], len(s.outputWires))
	copy(out, s.outputWires)
	return out
}
