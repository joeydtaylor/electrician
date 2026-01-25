# adapter/httpclient

The httpclient adapter provides a configurable HTTP client for fetching remote data and optionally submitting results into a pipeline. It supports TLS configuration, OAuth flows, retry policies, and structured telemetry.

## Responsibilities

- Build and configure HTTP clients.
- Perform fetch operations with configurable request settings.
- Support OAuth2 flows and TLS pinning where configured.
- Emit telemetry for request outcomes and retries.

## Key types and functions

- HTTPClientAdapter[T]: main adapter.
- Fetch(): perform a request and return a response payload.
- Serve(): fetch and submit into a pipeline (when used with a plug/generator).

## Configuration

Common options include:

- Request method, URL, headers, body
- Timeout, retry policy, and backoff
- TLS configuration (including pinning)
- OAuth2 client settings
- Sensor and logger

## Observability

Sensors emit metrics for request counts, errors, retries, and success rates. Loggers capture structured request events.

## Usage

```go
adapter := builder.NewHTTPClientAdapter(
    ctx,
    builder.HTTPClientAdapterWithRequestConfig[Item]("GET", "https://example.com", nil),
    builder.HTTPClientAdapterWithTimeout[Item](5*time.Second),
    builder.HTTPClientAdapterWithSensor[Item](sensor),
)
```

## References

- examples: example/plug_example/httpadapter/
- builder: pkg/builder/httpclient_adapter.go
- internal contracts: pkg/internal/types/httpclient_adapter.go
