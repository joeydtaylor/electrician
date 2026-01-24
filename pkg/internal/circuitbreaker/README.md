# Circuit Breaker

The circuit breaker protects a pipeline from repeated failures by opening when errors exceed a threshold.
While open, callers should divert or drop work until the breaker resets.

## Responsibilities

- Decide whether work is allowed (`Allow`).
- Record failures (`RecordError`) and trip when the threshold is reached.
- Notify observers through sensors, loggers, and the reset channel.

## Behavior

- `RecordError` increments the error count and can trip the breaker.
- `Allow` returns false while open and auto-resets once the time window elapses.
- `Reset` manually closes the breaker and clears the error count.
- `SetDebouncePeriod` coalesces rapid error bursts (seconds).

## Integration With Wire

Wire uses the breaker to gate submissions and optionally divert items to neutral wires when open.
Neutral wires are best-effort alternatives, not durable queues.

## Configuration Notes

Configure the breaker before using it in a running pipeline. Avoid mutating configuration while
it is being used concurrently.

## Extending

- Contracts: `pkg/internal/types/circuitbreaker.go`
- Implementation: `pkg/internal/circuitbreaker`
- Builder entry points: `pkg/builder/circuitbreaker.go`

## Examples

See `example/circuit_breaker_example/`.

## License

Apache 2.0. See `LICENSE`.
