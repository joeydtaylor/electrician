# Resister

Resister is an in-memory queue used to buffer elements when a pipeline cannot accept work immediately.
It is typically paired with a surge protector and drained by a wire loop once capacity returns.

## Responsibilities

- Enqueue and dequeue elements with priority awareness.
- Expose queue size for monitoring and control decisions.
- Emit sensor and logger events for queue activity.

## Behavior

- Higher `QueuePriority` elements are dequeued first.
- Requeueing an existing element updates its position instead of duplicating it.
- `DecayPriorities` can lower priority for elements that have exceeded retry thresholds.

## Integration With Surge Protector

The surge protector uses resister as a temporary buffer when it is tripped or rate-limited.
The wire drain loop replays items directly into the wire so queued elements are not re-enqueued.

## Configuration Notes

Resister is in-memory only. It is not durable storage and makes no ordering guarantees beyond
priority-based dequeueing.

## Extending

- Contracts: `pkg/internal/types/resister.go`
- Implementation: `pkg/internal/resister`
- Builder entry points: `pkg/builder/resister.go`

## Examples

See `example/surge_protector_example/`.

## License

Apache 2.0. See `LICENSE`.
