# meter

The meter package aggregates counters and computes rates and percentages for running pipelines. A meter is updated by sensors and can emit periodic summaries via Monitor().

## Responsibilities

- Track counters (submitted, processed, transformed, errors).
- Compute rates and percentages from counters.
- Sample runtime stats (CPU, RAM, goroutines) during monitor updates.
- Provide a stable surface for metrics used across components.

## Key types and functions

- Meter[T]: main type.
- Monitor(): periodic progress output loop.
- SetMetricCount / IncrementCount / DecrementCount.
- SetMetricPercentage / SetMetricPeak.
- SetDynamicMetric: register custom metrics at runtime.

## Configuration

Meters are configured with functional options:

- WithTotalItems: expected total items for progress.
- WithIdleTimeout: cancel after inactivity.
- WithUpdateFrequency: monitor tick cadence.
- WithDynamicMetric: register custom metrics.
- WithLogger: emit log messages.

## Behavior

Meters do not collect metrics on their own. Sensors and components must increment counters. Monitor() is purely a reporting loop; it does not control pipeline execution unless you wire cancellation hooks or idle timeouts.

## Usage

```go
meter := builder.NewMeter[Item](
    ctx,
    builder.MeterWithTotalItems[Item](1000),
    builder.MeterWithIdleTimeout[Item](10*time.Second),
)

sensor := builder.NewSensor(builder.SensorWithMeter[Item](meter))

meter.Monitor()
```

## References

- examples: example/meter_example/
- builder: pkg/builder/meter.go
- internal contracts: pkg/internal/types/meter.go
