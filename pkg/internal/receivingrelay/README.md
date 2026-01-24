# Receiving Relay

The receiving relay is Electrician's gRPC server for ingesting relay envelopes. It implements the relay service and forwards decoded items to downstream submitters.

## Responsibilities

- Accept unary and streaming relay RPCs.
- Decode `WrappedPayload` messages into application types.
- Enforce optional ingress policies (TLS, auth, metadata checks).
- Forward decoded items to configured outputs.

## Non-goals

- Durable delivery or replay guarantees.
- Exactly-once semantics.
- Queueing or persistence.

## Message handling

- `WrappedPayload` is the core envelope.
- `RelayEnvelope` wraps streaming messages (`StreamOpen`, payloads, `StreamClose`).
- Payload decoding defaults to gob unless payload encoding or content type indicates otherwise.

## Compression, encryption, auth

- Compression and payload encryption are opt-in via metadata.
- TLS/mTLS protects the transport; payload encryption is separate.
- Auth options are hints; enforcement is handled by configured policy and interceptors.

## Lifecycle

Receiving relays follow the standard component lifecycle:

1. Configure options.
2. Call `Start(ctx)` or `Listen(...)`.
3. Relay messages to configured outputs.
4. Call `Stop()` to shut down.

Configuration is immutable after `Start()` or `Listen()`.

## Package layout

- `receivingrelay.go`: type definition and constructor
- `listen.go`: gRPC server startup
- `receive.go`: unary handler
- `stream.go`: streaming handler
- `payload.go`: unwrap/decode
- `auth*.go`: auth policy helpers
- `options.go`: functional options
- `*_test.go`: tests

## Extending

- Protocol changes: update `.proto`, regenerate, then update forward/receiving relays.
- Cross-component behavior: update `pkg/internal/types/receivingrelay.go` first.
- User-facing knobs: expose via `pkg/builder`.

## References

- `README.md`
- `proto/README.md`
- `example/relay_example/`
