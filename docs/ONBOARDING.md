# Onboarding

Welcome to Electrician. This repo is a Go library plus runnable examples.

## Requirements
- Go: `1.24.3` (see `.tool-versions` and `go.mod`)
- Optional: Docker (for LocalStack + Redpanda integration tests)

## Quick start

Run unit tests:

```bash
go test ./... -race
```

Run a secure relay example (gRPC-Web + JWKS + AES-GCM):

```bash
go run ./example/relay_example/secure_advanced_relay_b_oauth_offline_jwks_mtls_aes_grpcweb
```

Run the mock OAuth server for local tokens:

```bash
go run ./example/relay_example/mock_oauth_server
```

## Common docs
- Relay gRPC-Web / Connect: `docs/relay-grpc-connect.md`
- Relay CORS: `docs/relay-grpcweb-cors.md`
- JWT issuer mismatch: `docs/relay-issuer-mismatch.md`
- Relay examples index: `example/relay_example/README.md`

## Integration tests (LocalStack + Redpanda)

```bash
docker network create steeze-edge || true
docker compose -f local-stack/docker-compose.yml up -d
go test ./pkg/... -tags=integration -count=1
```
