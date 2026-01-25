# sensor

The sensor package provides structured callbacks for pipeline events. Sensors are the primary integration point for observability and instrumentation; they emit events that meters and loggers consume.

## Responsibilities

- Register callbacks for lifecycle, submit, error, and processing events.
- Connect meters and loggers to emit structured telemetry.
- Keep instrumentation separate from hot-path processing logic.

## Key types and functions

- Sensor[T]: main type.
- Register* and Invoke* callback methods.
- WithSensor options for components.

## Configuration

Sensors are configured with:

- A meter (optional)
- A logger (optional)
- Custom callbacks

Sensors should be attached before Start().

## Behavior

Sensors are best-effort. They should not block the pipeline or introduce heavy work in the hot path.

## Usage

```go
meter := builder.NewMeter[Item](ctx)
logger := builder.NewLogger()

sensor := builder.NewSensor(
    builder.SensorWithMeter[Item](meter),
    builder.SensorWithLogger[Item](logger),
)
```

## References

- examples: example/sensor_example/
- builder: pkg/builder/sensor.go
- internal contracts: pkg/internal/types/sensor.go
