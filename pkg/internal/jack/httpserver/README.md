# jack/httpserver

The httpserver package provides an inbound HTTP/HTTPS server adapter. It exposes a typed request handler, handles request decoding, and returns structured responses suitable for pipeline entrypoints.

## Responsibilities

- Bind an HTTP server and route requests.
- Decode request bodies into typed structs.
- Invoke user handlers with context and typed requests.
- Build responses with status codes, headers, and byte payloads.
- Support TLS configuration.

## Key types and functions

- HTTPServer[T]: main server type.
- Serve(ctx, handler): start the server and handle requests.
- HTTPServerResponse: structured response (status, headers, body).

## Configuration

Common options include:

- Address, method, and route
- TLS configuration
- Read/write timeouts
- Default headers
- OAuth2 introspection or custom auth validators
- Logger and sensor

Configuration must be finalized before Serve(). Mutation after start panics to avoid races.

## Error handling

Handler errors can be mapped to HTTP error responses. The server propagates fatal errors via Serve() and respects context cancellation.

## Authentication

OAuth2 introspection is supported through authentication options. You can also enforce static headers or install a custom validation callback. Auth failures return 401 responses when strict enforcement is enabled.

## Observability

Sensors emit request start/finish/error events. Loggers capture structured access and error logs.

## Usage

```go
server := builder.NewHTTPServer[Request](
    ctx,
    builder.HTTPServerWithAddress[Request](":8443"),
    builder.HTTPServerWithServerConfig[Request]("POST", "/hello"),
    builder.HTTPServerWithTLS[Request](tlsConfig),
)

err := server.Serve(ctx, func(ctx context.Context, req Request) (builder.HTTPServerResponse, error) {
    return builder.HTTPServerResponse{StatusCode: 200, Body: []byte("ok")}, nil
})
```

## References

- examples: example/httpserverJack/ (including `oauth_introspection/`)
- builder: pkg/builder/jack.go
- internal contracts: pkg/internal/types/httpserver_adapter.go
