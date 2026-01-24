package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *Sensor[T]) snapshotMeters() []types.Meter[T] {
	s.metersLock.Lock()
	meters := append([]types.Meter[T](nil), s.meters...)
	s.metersLock.Unlock()
	return meters
}

func (s *Sensor[T]) incrementMeterCounters(metric string) {
	for _, m := range s.snapshotMeters() {
		m.IncrementCount(metric)
	}
}

func (s *Sensor[T]) setMetricTimestampValue(metric string, time int64) {
	for _, m := range s.snapshotMeters() {
		m.SetMetricTimestamp(metric, time)
	}
}

func (s *Sensor[T]) incrementMeterCountersAndReportActivity(metric string) {
	for _, m := range s.snapshotMeters() {
		m.ReportData()
		m.IncrementCount(metric)
	}
}

func (s *Sensor[T]) decrementMeterCounters(metric string) {
	for _, m := range s.snapshotMeters() {
		m.DecrementCount(metric)
	}
}

func (s *Sensor[T]) decorateCallbacks(options ...types.Option[types.Sensor[T]]) []types.Option[types.Sensor[T]] {
	options = append(options, s.decorateCircuitBreakerCallbacks()...)
	options = append(options, s.decorateHttpClientCallbacks()...)
	options = append(options, s.decorateSurgeProtectorCallbacks()...)
	options = append(options, s.decorateWireCallbacks()...)
	options = append(options, s.decorateResisterCallbacks()...)

	options = append(
		options,
		WithOnRestartFunc[T](func(c types.ComponentMetadata) {
			s.incrementMeterCounters(types.MetricComponentRestartCount)
		}),
		WithOnStartFunc[T](func(c types.ComponentMetadata) {
			switch c.Type {
			case "WIRE":
				s.incrementMeterCounters(types.MetricWireRunningCount)
			case "GENERATOR":
				s.incrementMeterCounters(types.MetricGeneratorRunningCount)
			case "FORWARD_RELAY":
				s.incrementMeterCounters(types.MetricForwardRelayRunningCount)
			case "CONDUIT":
				s.incrementMeterCounters(types.MetricConduitRunningCount)
			case "RECEIVING_RELAY":
				s.incrementMeterCounters(types.MetricReceivingRelayRunningCount)
			}
			s.incrementMeterCounters(types.MetricComponentRunningCount)
		}),
		WithOnStopFunc[T](func(c types.ComponentMetadata) {
			switch c.Type {
			case "WIRE":
				s.decrementMeterCounters(types.MetricWireRunningCount)
			case "GENERATOR":
				s.decrementMeterCounters(types.MetricGeneratorRunningCount)
			case "FORWARD_RELAY":
				s.decrementMeterCounters(types.MetricForwardRelayRunningCount)
			case "CONDUIT":
				s.decrementMeterCounters(types.MetricConduitRunningCount)
			case "RECEIVING_RELAY":
				s.decrementMeterCounters(types.MetricReceivingRelayRunningCount)
			}
			s.decrementMeterCounters(types.MetricComponentRunningCount)
		}),
	)

	return options
}
