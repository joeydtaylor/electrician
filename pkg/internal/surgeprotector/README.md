# Surge Protector

Surge protector provides admission control for pipelines by rate limiting submissions and optionally
buffering excess work in a resister queue. It is intended to protect the hot path during bursts or
temporary overload.

## Responsibilities

- Rate limit submissions with a token bucket (`TryTake`, `SetRateLimit`).
- Buffer overflow in a resister when connected.
- Route to backup submitters when configured.
- Emit telemetry through loggers and sensors.

## Behavior

- When rate limiting is enabled, `TryTake` controls whether a submission can proceed immediately.
- If backups are configured, submissions are redirected to those systems first.
- If a resister is connected, the protector drains queued work when tokens are available.
- `Trip` and `Reset` control the active state and can trigger restarts of managed components.

## Integration With Wire

Wire delegates burst handling to the surge protector. When rate limited or tripped, the wire hands
elements to the protector, which may enqueue them in a resister. The wire drain loop replays queued
elements directly into the wire to avoid re-queueing.

## Configuration Notes

Configure the surge protector before wiring it into a running pipeline. It does not provide durable
storage and should not be treated as a broker.

## Extending

- Contracts: `pkg/internal/types/surgeprotector.go`
- Implementation: `pkg/internal/surgeprotector`
- Builder entry points: `pkg/builder/surgeprotector.go`

## Examples

See `example/surge_protector_example/`.

## License

Apache 2.0. See `LICENSE`.
