# resister

The resister package provides a queue-based smoothing layer. It buffers items when downstream systems are overloaded and releases them later according to policy.

## Responsibilities

- Queue items for deferred processing.
- Dequeue items in order when capacity is available.
- Emit telemetry for queue depth and dequeue events.

## Key types and functions

- Resister[T]: main type.
- Queue/Dequeue operations and lifecycle controls.

## Configuration

Resisters are configured with:

- Queue size and behavior
- Optional sensor and logger

Configuration must be finalized before Start().

## Behavior

Resisters are often paired with surge protectors or circuit breakers to avoid dropping data during temporary overloads.

## Observability

Sensors emit queue/dequeue/empty events. Meters can track depth and throughput.

## Usage

```go
resister := builder.NewResister(
    ctx,
    builder.ResisterWithSensor(sensor),
)
```

## References

- builder: pkg/builder/resister.go
- internal contracts: pkg/internal/types/resister.go
