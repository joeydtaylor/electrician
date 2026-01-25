# surgeprotector

The surgeprotector package applies rate limiting and overload behavior to submissions. It can drop, delay, or reroute items based on configurable policies and integrates with a resister queue for deferred processing.

## Responsibilities

- Enforce rate limits and burst control.
- Trip on overload and reset when conditions recover.
- Optionally route overflow to a backup wire or resister queue.

## Key types and functions

- SurgeProtector[T]: main type.
- Trip() / Reset(): manual control for blackout modes.

## Configuration

Surge protectors are configured with:

- Rate limits (tokens, windows, burst)
- Backup wire for overflow
- Resister for queueing
- Optional sensor and logger

Configuration must be finalized before Start().

## Behavior

- Rate limiting can be strict or burst-based.
- Blackout and trip modes allow explicit outage control.
- Backup wiring keeps data flowing to an alternate path when overloaded.

## Observability

Sensors emit events for trips, resets, and overflow submissions. Meters can track rates and backup routing percentages.

## Usage

```go
sp := builder.NewSurgeProtector(
    ctx,
    builder.SurgeProtectorWithRateLimit[Item](1, 5*time.Second, 10),
    builder.SurgeProtectorWithSensor(sensor),
)
```

## References

- examples: example/surge_protector_example/
- builder: pkg/builder/surgeprotector.go
- internal contracts: pkg/internal/types/surgeprotector.go
