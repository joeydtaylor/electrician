# adapter/websocketclient

The websocketclient adapter connects to a WebSocket endpoint and streams typed messages in and out. It supports json, text, binary, and proto payloads with configurable timeouts and TLS settings.

## Responsibilities

- Dial WebSocket endpoints (ws/wss).
- Decode inbound frames into typed messages.
- Encode outbound messages from a channel.
- Run duplex read/write loops on a single connection.

## Key types and functions

- WebSocketClientAdapter[T]: main adapter type.
- Serve(ctx, submit): read-only loop.
- ServeWriter(ctx, in): write-only loop.
- ServeDuplex(ctx, in, submit): full duplex loop.

## Configuration

Common options include:

- URL and headers
- Message format (json, text, binary, proto)
- Read limit, write timeout, idle timeout
- TLS configuration
- Logger and sensor

## WebAssembly

The adapter uses nhooyr.io/websocket, which supports js/wasm clients. TLS customization is ignored by browsers, but the same adapter code compiles for WASM targets.

## Usage

```go
adapter := builder.NewWebSocketClientAdapter[Event](
    ctx,
    builder.WebSocketClientAdapterWithURL[Event]("ws://localhost:8080/ws"),
    builder.WebSocketClientAdapterWithMessageFormat[Event]("json"),
)

err := adapter.Serve(ctx, wire.Submit)
```

## References

- builder: pkg/builder/websocket_adapter.go
- internal contracts: pkg/internal/types/websocket_adapter.go
