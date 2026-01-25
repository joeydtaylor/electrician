# receivingrelay

The receivingrelay package implements the inbound side of the relay system. It accepts gRPC streams, unwraps payloads (decompression and decryption), and submits them to downstream components.

## Responsibilities

- Listen for gRPC relay streams.
- Validate authentication and TLS settings.
- Unwrap payloads with compression and encryption settings.
- Optional passthrough mode for raw WrappedPayload forwarding.
- Submit decoded items to downstream handlers.

## Key types and functions

- ReceivingRelay[T]: main type.
- Listen(listenForever, retryInSeconds): start the gRPC listener.
- ListenGRPCWeb(listenForever, retryInSeconds): start the gRPC-Web listener.
- Start(ctx): start the relay with a gRPC listener.
- StartGRPCWeb(ctx): start the relay with a gRPC-Web listener.
- Receive(ctx, payload): receive and submit a payload.

## Configuration

Common options include:

- Listener address and TLS configuration
- OAuth2 or token introspection settings
- Compression and encryption options
- Sensor and logger

Configuration must be finalized before Start()/Listen() or StartGRPCWeb()/ListenGRPCWeb().

Passthrough mode expects the relay type parameter to be `relay.WrappedPayload` (or pointer) so the
raw payload and metadata can be forwarded without decoding.

## Error handling

Unwrap and decode errors are surfaced to callers and reported via telemetry. Invalid auth and TLS failures are rejected at the connection level.

## Observability

Sensors emit metrics for received payloads, unwrap successes/failures, and relay lifecycle events. Loggers capture connection and runtime events.

## References

- examples: example/relay_example/
- builder: pkg/builder/receivingrelay.go
- internal contracts: pkg/internal/types/receivingrelay.go
