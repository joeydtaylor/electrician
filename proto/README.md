# proto

This directory contains the protobuf definitions for Electrician relays. The generated Go code lives under pkg/internal/relay and is used by the forward and receiving relay packages.

## Contents

- electrician_relay.proto: relay service definitions and message types

## Generation

Regenerate stubs with `protoc` when modifying the proto definition. The generated files are committed to keep consumers and internal packages in sync.

## TL;DR (copy/paste)

```bash
PROTOC_INCLUDE="$(dirname "$(which protoc)")/../include"

protoc \
  -I=proto \
  -I="$PROTOC_INCLUDE" \
  --go_out=pkg/internal/relay \
  --go_opt=paths=source_relative \
  --go-grpc_out=pkg/internal/relay \
  --go-grpc_opt=paths=source_relative \
  proto/electrician_relay.proto
```

If you get `protoc: command not found`, run the prerequisites below first.

### 1) Prereqs (one-time)

You need:
- `protoc` (the protobuf compiler)
- `protoc-gen-go` and `protoc-gen-go-grpc` on your PATH

Quick install (macOS/Homebrew):

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Make sure your Go bin is on PATH (only needed once):

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

### 2) Copy/paste command (from repo root)

The proto imports `google/protobuf/timestamp.proto` and `google/protobuf/duration.proto`, so we must include the well-known types path.

```bash
PROTOC_INCLUDE="$(dirname "$(which protoc)")/../include"

protoc \
  -I=proto \
  -I="$PROTOC_INCLUDE" \
  --go_out=pkg/internal/relay \
  --go_opt=paths=source_relative \
  --go-grpc_out=pkg/internal/relay \
  --go-grpc_opt=paths=source_relative \
  proto/electrician_relay.proto
```

### 3) Sanity check

You should see these updated files:
- `pkg/internal/relay/electrician_relay.pb.go`
- `pkg/internal/relay/electrician_relay_grpc.pb.go`

If `protoc` complains about missing `google/protobuf/*.proto`, your include path is wrong.

### Optional: Homebrew include path

If you prefer, this is equivalent on macOS/Homebrew:

```bash
protoc \
  -I=proto \
  -I="$(brew --prefix protobuf)/include" \
  --go_out=pkg/internal/relay \
  --go_opt=paths=source_relative \
  --go-grpc_out=pkg/internal/relay \
  --go-grpc_opt=paths=source_relative \
  proto/electrician_relay.proto
```

## Usage

The forwardrelay and receivingrelay packages depend on these definitions to stream payloads across services.

## License

Apache 2.0. See ../LICENSE.
