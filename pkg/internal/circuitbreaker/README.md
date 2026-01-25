# circuitbreaker

The circuitbreaker package provides trip/reset behavior for pipelines. It tracks error counts and trips when thresholds are exceeded, allowing downstream logic to divert or drop items until the breaker resets.

## Responsibilities

- Track error rates or counts.
- Trip when thresholds are exceeded.
- Reset after a cooldown period.
- Optionally route items to a neutral wire.

## Key types and functions

- CircuitBreaker[T]: main type.
- Trip() / Reset(): manual control.
- Start(ctx) / Stop(): lifecycle and timer management.

## Configuration

Circuit breakers are configured with:

- Trip threshold and reset duration
- Optional neutral wire for diversion
- Optional sensor and logger

Configuration must be finalized before Start().

## Behavior

- When tripped, submissions can be diverted or blocked depending on wiring.
- Reset timers control when the breaker returns to normal.

## Observability

Sensors emit trip/reset events and error counts. Meters can track trip frequency and error ratios.

## Usage

```go
cb := builder.NewCircuitBreaker(
    ctx,
    3,
    5*time.Second,
    builder.CircuitBreakerWithSensor(sensor),
)
```

## References

- examples: example/circuit_breaker_example/
- builder: pkg/builder/circuitbreaker.go
- internal contracts: pkg/internal/types/circuitbreaker.go
