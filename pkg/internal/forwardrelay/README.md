# forwardrelay

The forwardrelay package streams payloads to downstream services using gRPC. It handles payload wrapping (compression and encryption), authentication, and TLS configuration.

## Responsibilities

- Connect to remote relay targets.
- Wrap payloads with compression and encryption as configured.
- Stream payloads over gRPC.
- Emit telemetry for connections and submissions.

## Key types and functions

- ForwardRelay[T]: main type.
- Submit(ctx, item): wrap and stream an item.
- Start(ctx) / Stop(): lifecycle management.

## Configuration

Common options include:

- Target endpoints and connection settings
- TLS configuration and client certificates
- OAuth2 or token-based authentication
- Compression and encryption options
- Sensor and logger

Configuration must be finalized before Start().

## Error handling

Errors during wrapping or transmission are surfaced to callers and reported via sensors/loggers.

## Observability

Sensors emit metrics for submissions, payload wrapping, and errors. Loggers capture connection and runtime events.

## References

- examples: example/relay_example/
- builder: pkg/builder/forwardrelay.go
- internal contracts: pkg/internal/types/forwardrelay.go
