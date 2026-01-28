# Relay gRPC-Web CORS Requirements

Browsers issue a CORS preflight (OPTIONS) before the gRPC-web POST. The relay
server (or TLS proxy) must respond with the required headers or the request
will be blocked by the browser.

## Required headers for OPTIONS and POST
- `Access-Control-Allow-Origin: https://localhost:3001`
  - If multiple origins are allowed, echo the incoming Origin.
- `Access-Control-Allow-Methods: POST, OPTIONS`
- `Access-Control-Allow-Headers:`
  - allow all, or at minimum:
    - `content-type`
    - `connect-protocol-version`
    - `x-grpc-web`
    - `x-user-agent`
    - `grpc-timeout`
    - `grpc-encoding`
    - `grpc-accept-encoding`
    - `authorization`
    - `x-tenant`
- `Access-Control-Expose-Headers:`
  - `grpc-status, grpc-message, grpc-status-details-bin`

If cookies are required, also set:
- `Access-Control-Allow-Credentials: true`

> When `Allow-Credentials` is true, `Allow-Origin` must be a specific origin
> (not `*`).

## Server-side config (Go)
Electrician supports a gRPC-web config hook:

```go
grpcWebCfg := &types.GRPCWebConfig{
  AllowedOrigins: []string{"https://localhost:3001"},
  AllowedHeaders: []string{
    "content-type",
    "connect-protocol-version",
    "x-grpc-web",
    "x-user-agent",
    "grpc-timeout",
    "grpc-encoding",
    "grpc-accept-encoding",
    "authorization",
    "x-tenant",
  },
  AllowCredentials: true,
}

recv := builder.NewReceivingRelay[Feedback](
  ctx,
  builder.ReceivingRelayWithGRPCWebConfig[Feedback](grpcWebCfg),
)
```

For dev-only use, you can allow all origins:

```go
grpcWebCfg := builder.NewGRPCWebConfigAllowAllOrigins()
```

## Validate
- DevTools → Network → OPTIONS should return the headers above.
- The POST should then succeed and reach the relay handler.
