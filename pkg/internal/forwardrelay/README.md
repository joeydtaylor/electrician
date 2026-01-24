# Forward Relay

The forward relay is Electrician's gRPC client for sending relay envelopes to a receiving relay. It wraps items into `WrappedPayload` messages and streams them to configured targets.

## Responsibilities

- Wrap items into relay envelopes with metadata, timestamps, and sequencing.
- Apply optional compression and payload encryption.
- Attach optional auth metadata and custom headers per RPC.
- Maintain a streaming connection and handle acknowledgments.

## Non-goals

- Durable delivery or replay guarantees.
- Exactly-once processing.
- Queue semantics or persistence.

## Pipeline fit

Typical flow:

`Wire -> ForwardRelay -> ReceivingRelay`

The forward relay is an egress adapter. It should not implement business logic.

## Message model

The relay protocol is defined in protobuf:

- `WrappedPayload` is the core envelope.
- `RelayEnvelope` wraps streaming messages (`StreamOpen`, payloads, `StreamClose`).

Important fields:

- `payload` contains serialized bytes (gob by default).
- `metadata.content_type` describes the payload bytes.
- `metadata.headers` carries optional context.
- `trace_id` is used for correlation.

## Compression, encryption, auth

- Compression and encryption are opt-in via `PerformanceOptions` and `SecurityOptions`.
- TLS/mTLS protects the transport; payload encryption is separate and explicit.
- Auth options and metadata are hints; enforcement is handled by the receiver.

## Lifecycle

Forward relays follow the standard component lifecycle:

1. Configure options.
2. Call `Start(ctx)`.
3. Submit items or connect inputs.
4. Call `Stop()` to shut down.

Configuration is immutable after `Start()`.

## Package layout

- `forwardrelay.go`: type definition and constructor
- `connect.go`: inputs and loggers
- `config.go`: configuration setters
- `lifecycle.go`: start/stop
- `payload.go`: wrapping, compression, encryption
- `stream.go`: gRPC streaming and acks
- `options.go`: functional options
- `*_test.go`: tests

## Extending

- Protocol changes: update `.proto`, regenerate, then update forward/receiving relays.
- Cross-component behavior: update `pkg/internal/types/forwardrelay.go` first.
- User-facing knobs: expose via `pkg/builder`.

## References

- `README.md`
- `proto/README.md`
- `example/relay_example/`
