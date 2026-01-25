package websocketclient

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/joeydtaylor/electrician/pkg/internal/types"
	"nhooyr.io/websocket"
)

// Serve reads from the connection and dispatches decoded messages to submit.
func (c *WebSocketClientAdapter[T]) Serve(ctx context.Context, submit func(context.Context, T) error) error {
	if submit == nil {
		return errors.New("submit cannot be nil")
	}
	if ctx == nil {
		ctx = c.ctx
	}

	cfg := c.snapshotConfig()
	conn, err := c.dial(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "client shutdown")

	c.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Dial, url: %s", c.componentMetadata, cfg.url)
	defer c.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Disconnect", c.componentMetadata)

	return c.readLoop(ctx, cfg, conn, submit)
}

// ServeWriter encodes and writes messages from in to the connection.
func (c *WebSocketClientAdapter[T]) ServeWriter(ctx context.Context, in <-chan T) error {
	if in == nil {
		return errors.New("input channel cannot be nil")
	}
	if ctx == nil {
		ctx = c.ctx
	}

	cfg := c.snapshotConfig()
	conn, err := c.dial(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "client shutdown")

	c.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Dial, url: %s", c.componentMetadata, cfg.url)
	defer c.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Disconnect", c.componentMetadata)

	return c.writeLoop(ctx, cfg, conn, in)
}

// ServeDuplex runs both read and write loops on a single connection.
func (c *WebSocketClientAdapter[T]) ServeDuplex(ctx context.Context, in <-chan T, submit func(context.Context, T) error) error {
	if submit == nil {
		return errors.New("submit cannot be nil")
	}
	if ctx == nil {
		ctx = c.ctx
	}

	cfg := c.snapshotConfig()
	conn, err := c.dial(ctx, cfg)
	if err != nil {
		return err
	}
	defer conn.Close(websocket.StatusNormalClosure, "client shutdown")

	c.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Dial, url: %s", c.componentMetadata, cfg.url)
	defer c.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: Disconnect", c.componentMetadata)

	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 2)
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		errCh <- c.readLoop(connCtx, cfg, conn, submit)
	}()
	go func() {
		defer wg.Done()
		if in == nil {
			errCh <- nil
			return
		}
		errCh <- c.writeLoop(connCtx, cfg, conn, in)
	}()

	var firstErr error
	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil && firstErr == nil {
			firstErr = err
			cancel()
		}
	}

	wg.Wait()
	return firstErr
}

func (c *WebSocketClientAdapter[T]) readLoop(ctx context.Context, cfg clientConfig, conn *websocket.Conn, submit func(context.Context, T) error) error {
	for {
		readCtx := ctx
		var cancel context.CancelFunc
		if cfg.idleTimeout > 0 {
			readCtx, cancel = context.WithTimeout(ctx, cfg.idleTimeout)
		}
		_, payload, err := conn.Read(readCtx)
		if cancel != nil {
			cancel()
		}
		if err != nil {
			return c.handleReadError(err)
		}

		msg, err := decodeMessage[T](cfg.format, payload)
		if err != nil {
			c.NotifyLoggers(types.WarnLevel, "ReadLoop: decode error: %v", err)
			continue
		}

		if err := submit(ctx, msg); err != nil {
			c.NotifyLoggers(types.WarnLevel, "ReadLoop: submit error: %v", err)
		}
	}
}

func (c *WebSocketClientAdapter[T]) handleReadError(err error) error {
	if err == nil {
		return nil
	}
	status := websocket.CloseStatus(err)
	if status == websocket.StatusNormalClosure || status == websocket.StatusGoingAway {
		return nil
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return nil
	}
	return err
}

func (c *WebSocketClientAdapter[T]) writeLoop(ctx context.Context, cfg clientConfig, conn *websocket.Conn, in <-chan T) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-in:
			if !ok {
				return nil
			}
			messageType, payload, err := encodeMessage(cfg.format, msg)
			if err != nil {
				return fmt.Errorf("encode message: %w", err)
			}
			writeCtx := ctx
			var cancel context.CancelFunc
			if cfg.writeTimeout > 0 {
				writeCtx, cancel = context.WithTimeout(ctx, cfg.writeTimeout)
			}
			if err := conn.Write(writeCtx, messageType, payload); err != nil {
				if cancel != nil {
					cancel()
				}
				return err
			}
			if cancel != nil {
				cancel()
			}
		}
	}
}
