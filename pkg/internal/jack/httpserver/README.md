# HTTP Server

The HTTP server adapter exposes a webhook-style endpoint that decodes request bodies into `T` and routes them into your pipeline via a submit function. It is intentionally minimal and focused on ingestion.

## Responsibilities

- Accept inbound HTTP requests for a single method/path.
- Decode request bodies into `T` (JSON by default).
- Invoke a submit function and shape the HTTP response.
- Optionally run with TLS using `types.TLSConfig`.

## Non-goals

- General web framework features.
- Routing beyond a single method/path.
- Durable delivery or retry semantics.

## Lifecycle

Configure the server with options, then call `Serve(ctx, submitFunc)`. Configuration is frozen once `Serve` starts.

## Configuration

- `WithAddress` sets the listen address.
- `WithServerConfig` sets the method and endpoint.
- `WithHeader` sets default response headers.
- `WithTimeout` sets read/write timeouts.
- `WithTLS` enables TLS.

## Error handling

If `submitFunc` returns an error, the server responds with a 500 by default. You can return a `*types.HTTPServerError` to control the status code and message.

## Package layout

- `httpserver.go`: type and constructor
- `serve.go`: server lifecycle and request handling
- `decode.go`: request decoding
- `tls.go`: TLS config loading
- `options.go`: functional options
- `*_test.go`: tests
