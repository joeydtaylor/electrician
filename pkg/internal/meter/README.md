# Meter

The meter package provides lightweight metrics aggregation and reporting for running pipelines. A meter stores counters, calculates rates and percentages, and can print periodic summaries when `Monitor()` is running.

Meters do not collect metrics on their own. Components emit events, a sensor consumes those events, and the sensor increments meter counters. The meter is intentionally simple so it stays out of the hot path.

## Responsibilities

- Count items submitted, processed, transformed, and errored.
- Derive per-second rates and percentages from counters.
- Sample runtime stats (CPU, RAM, goroutines) during monitor updates.
- Emit periodic snapshots through `Monitor()` without blocking the pipeline.

## Usage

```go
meter := builder.NewMeter[Item](
    ctx,
    builder.MeterWithTotalItems[Item](1000),
    builder.MeterWithIdleTimeout[Item](10*time.Second),
    builder.MeterWithUpdateFrequency[Item](250*time.Millisecond),
)

sensor := builder.NewSensor(builder.SensorWithMeter[Item](meter))
```

Call `Monitor()` on the meter to print periodic progress updates. This is usually done in the main goroutine after the pipeline starts.

## Dynamic metrics

For custom counters, register a metric and optionally add it to the display:

```go
meter.SetDynamicMetric("custom_count", 0, 0, 0)
info, _ := meter.GetDynamicMetricInfo("custom_count")
meter.AddMetricMonitor(info)
```

## Notes

- The meter is concurrency-safe for updates from multiple goroutines.
- `Monitor()` is responsible for display output only; it does not manage pipeline lifecycle unless you wire it to a cancel hook or idle timeout.
- Avoid expensive work in the hot path. Increment counters quickly and let the monitor compute derived values on a ticker.

## References

- Root architecture: `README.md`
- Internal package overview: `pkg/internal/README.MD`
- Example: `example/meter_example/main.go`
