# pkg/internal

This directory contains Electrician's private implementation. It is not part of the public API. Go enforces this with the internal import rule:

- Packages under pkg/ can import pkg/internal/...
- External consumers cannot

In this repository:

- pkg/ is the public surface area
- pkg/internal/ is the private engine
- pkg/builder is the user-facing construction layer

## Why it exists

Electrician keeps implementation details private so the public API can remain stable while the internals evolve. This also allows aggressive refactors and performance work without breaking downstream users.

## Package map

| Package | Purpose |
| --- | --- |
| adapter/ | Integration adapters (Kafka, S3, HTTP, auth) |
| jack/httpserver/ | Inbound HTTP/HTTPS server |
| circuitbreaker/ | Failure detection and trip/reset behavior |
| codec/ | Encoding and compression helpers |
| conduit/ | Multi-stage composition |
| forwardrelay/ | Outbound gRPC relay |
| receivingrelay/ | Inbound gRPC relay |
| relay/ | Protobuf and gRPC contract code |
| generator/ | Source components and pollers |
| internallogger/ | Structured logging backend |
| meter/ | Counters, rates, and progress reporting |
| plug/ | Pluggable input abstraction |
| resister/ | Queue-based smoothing and deferred delivery |
| sensor/ | Telemetry callbacks for pipeline events |
| surgeprotector/ | Rate limiting and overload behavior |
| types/ | Shared interfaces and constants |
| utils/ | Small internal helpers |
| wire/ | Core pipeline primitive |

## File conventions

Packages generally follow a predictable file layout where it makes sense:

- config.go: configuration state and defaults
- connect.go: wiring and dependency attachment
- options.go: functional options (With*)
- lifecycle.go: start/stop and runtime behavior
- accessors.go: read-only getters
- notify.go or telemetry.go: logging/metrics hooks
- immutability.go: guards against runtime mutation
- errors.go: package-specific errors

Not every package uses every file, but new packages should follow the same shape when applicable.

## Design contract

Configure -> Start -> Run -> Stop/Restart

Components are expected to be wired before Start(). Configuration mutation while running is not supported and may race.

## Dependencies

The pipeline core relies on the standard library. Integrations bring additional dependencies (Kafka, AWS SDK, Parquet, gRPC, codecs, host metrics). These are isolated to their packages and only linked when imported.

## Extending internal packages

1) Update contracts in pkg/internal/types when behavior is shared across components.
2) Implement in the target package.
3) Expose configuration through options.
4) Add tests, including performance checks where needed.

## References

- Root README: ../../README.md
- Examples: ../../example/
- Per-package READMEs under pkg/internal/*

## License

Apache 2.0. See ../../LICENSE.
