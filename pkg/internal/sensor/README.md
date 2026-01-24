# Sensor

Sensor provides callback hooks for observing pipeline activity across the core components. It is the
preferred way to emit telemetry without mutating live components at runtime.

## Responsibilities

- Register callbacks for lifecycle, error, and submission events.
- Emit adapter-specific hooks for HTTP, S3, and Kafka behaviors.
- Increment meter counters and timestamps through internal decorators.

## Usage

Sensors are attached to components during construction. Callbacks are executed synchronously on
invocation, so they should be fast and non-blocking.

## Configuration Notes

Sensor callbacks are safe to register before runtime. Avoid long-running work inside callbacks to
keep hot paths responsive.

## Extending

- Contracts: `pkg/internal/types/sensor.go`
- Implementation: `pkg/internal/sensor`
- Builder entry points: `pkg/builder/sensor.go`

## Examples

See `example/sensor_example/`.

## License

Apache 2.0. See `LICENSE`.
