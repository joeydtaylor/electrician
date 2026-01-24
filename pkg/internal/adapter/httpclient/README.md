# HTTP Client Adapter

The HTTP Client Adapter executes HTTP requests and decodes responses into `T` for use in Electrician pipelines. It is typically wired through a `Plug` and driven by a `Generator`.

## Responsibilities

- Build and execute HTTP requests using `net/http`.
- Decode response bodies into `T` based on content type.
- Emit telemetry callbacks for request lifecycle events.
- Provide optional polling behavior through `Serve`.

## Pipeline Composition

A common arrangement is:

```
HTTP Client Adapter -> Plug -> Generator -> Wire
```

The adapter owns transport and decoding. Business logic should live in wire transformers.

## Configuration

Core configuration options include:

- Request configuration (method, endpoint, body).
- Headers (static values including `Authorization`).
- Polling interval and retry limits for `Serve`.
- Request timeouts.
- Optional OAuth2 client credentials flow.
- Optional TLS certificate pinning.

## Response Decoding

The adapter decodes based on the response `Content-Type`:

- `application/json` uses the JSON decoder.
- `text/xml` and `application/xml` use the XML decoder.
- `application/octet-stream`, `text/plain`, and `text/html` are returned as raw bytes in the wrapped response.

## Telemetry

Sensors can subscribe to:

- Request start
- Response received
- Request complete
- Error

Loggers receive formatted messages from adapter operations.

## Package Layout

- `httpclient.go`: core types and constructor
- `config.go`: configuration setters
- `connect.go`: logger and sensor wiring
- `fetch.go`: request execution and decoding
- `serve.go`: polling and retry loop
- `oauth.go`: OAuth2 helpers
- `tls.go`: certificate pinning
- `notify.go`: telemetry hooks
- `options.go`: functional options
- `*_test.go`: tests

## References

- Root project README: `README.md`
- HTTP adapter examples: `example/plug_example/httpadapter/`
