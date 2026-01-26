# HTTP Server Jack with OAuth2 Introspection

This example shows how to protect an HTTP Jack endpoint with OAuth2 token introspection.
It starts two servers:

- An OAuth2 introspection mock server on `http://localhost:8081/introspect`
- The HTTPS HTTP Jack server on `https://localhost:8443/hello`

## Run

```bash
go run example/httpserverJack/oauth_introspection/main.go
```

Send a request with a valid token:

```bash
curl -k -X POST https://localhost:8443/hello \
  -H 'Authorization: Bearer token-123' \
  -d '{"name":"Joey"}'
```

The introspection server expects HTTP Basic auth using these defaults:

- `INTROSPECT_CLIENT_ID=example-client`
- `INTROSPECT_CLIENT_SECRET=example-secret`
- `INTROSPECT_TOKEN=token-123`
- `INTROSPECT_SCOPE=write:hello`

Override defaults with environment variables:

- `HTTP_ADDR` (default `:8443`)
- `INTROSPECT_ADDR` (default `localhost:8081`)
- `INTROSPECT_CLIENT_ID`
- `INTROSPECT_CLIENT_SECRET`
- `INTROSPECT_TOKEN`
- `INTROSPECT_SCOPE`
- `TLS_CERT`, `TLS_KEY`, `TLS_CA` (defaults use `example/httpserverJack/tls`)

## Notes

- The introspection endpoint is a local mock for demo purposes.
- The HTTP Jack enforces OAuth2 by installing the introspection validator from the auth options.
