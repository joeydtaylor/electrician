# builder

The builder package is Electrician's public construction layer. It provides ergonomic constructors (NewWire, NewConduit, NewRelay, etc.) and consistent functional options (WireWith..., ConduitWith..., etc.) so you can assemble pipelines without importing internal packages.

## Import boundary

- pkg/builder is the user-facing API
- pkg/internal/* is private implementation (enforced by Go's internal rule)

External consumers should not import pkg/internal.

## What builder does

- Constructs components with defaults
- Applies options consistently
- Keeps internal implementations swappable without API churn

Builder is intentionally thin; it should not add runtime overhead to the hot path.

## Functional options

Example (wire):

```go
w := builder.NewWire[string](
    ctx,
    builder.WireWithConcurrencyControl[string](1024, 8),
    builder.WireWithTransformer[string](func(v string) (string, error) {
        return strings.ToUpper(v), nil
    }),
    builder.WireWithCircuitBreaker[string](cb),
    builder.WireWithSurgeProtector[string](sp),
    builder.WireWithSensor[string](sensor),
)
```

Options are intended to be applied before Start(). Configuration mutation while running is not supported.

## Mapping to internal

The pattern is consistent:

1) builder.NewX(...) delegates to pkg/internal/x.NewX(...)
2) builder.XWithY(...) options call the underlying component methods
3) The result implements the shared types contracts

## Package layout

Builder exposes one file per component family:

- wire.go
- conduit.go
- circuitbreaker.go
- surgeprotector.go
- resister.go
- sensor.go
- meter.go
- logger.go
- codec.go
- adapter helpers (HTTP, Kafka, S3)
- forwardrelay.go / receivingrelay.go
- plug.go / generator.go
- jack.go (HTTP server)

## References

- Root README: ../../README.md
- Internal README: ../internal/README.MD
- Examples: ../../example/

## License

Apache 2.0. See ../../LICENSE.
