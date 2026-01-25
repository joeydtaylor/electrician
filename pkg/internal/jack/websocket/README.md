# jack/websocket

The websocket jack provides an inbound WebSocket server that decodes typed messages and feeds them into a pipeline. It can also broadcast outbound messages to connected clients, making it suitable for real-time dashboards, event feeds, and UI streaming.

## Responsibilities

- Accept WebSocket connections and enforce basic policies.
- Decode inbound frames into typed messages (json, text, binary, or proto).
- Invoke submit handlers to push data into a wire or other component.
- Broadcast processed outputs back to all connected clients.
- Support TLS and authentication policies.

## Key types and functions

- WebSocketServer[T]: main server type.
- Serve(ctx, submit): start the server and handle inbound messages.
- Broadcast(ctx, msg): push outbound messages to all clients.

## Configuration

Common options include:

- Address and endpoint path
- Allowed Origin patterns
- Token query parameter (for bearer extraction)
- Message format (json, text, binary, proto)
- Read limit, write timeout, idle timeout
- Max connections and per-connection send buffer
- TLS configuration
- Logger and sensor

Configuration must be finalized before Serve(). Mutation after start panics to avoid races.

## Authentication

The server supports:

- Static headers (exact matches)
- Custom auth validator callbacks
- OAuth2 introspection via AuthenticationOptions

Auth failures return 401 responses when strict enforcement is enabled. Token query parameters can be used for browser-style clients that cannot set headers.

## Observability

Loggers receive lifecycle and connection events. Sensors receive start/stop lifecycle callbacks (and can be extended if you need deeper hooks).

## WebAssembly note

The server is for native Go targets only. For WebAssembly, use the WebSocket client adapter.

## Usage

```go
server := builder.NewWebSocketServer[Event](
    ctx,
    builder.WebSocketServerWithAddress[Event](":8080"),
    builder.WebSocketServerWithEndpoint[Event]("/ws"),
    builder.WebSocketServerWithMessageFormat[Event]("json"),
)

err := server.Serve(ctx, wire.Submit)
```

## References

- examples: example/websocketJack/
- builder: pkg/builder/websocket_jack.go
- internal contracts: pkg/internal/types/websocket_adapter.go
