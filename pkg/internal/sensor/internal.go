package sensor

import "github.com/joeydtaylor/electrician/pkg/internal/types"

func (s *Sensor[T]) incrementMeterCounters(metric string) {
	for _, m := range s.meters {
		m.IncrementCount(metric)
	}
}

func (s *Sensor[T]) setMetricTimestampValue(metric string, time int64) {
	for _, m := range s.meters {
		m.SetMetricTimestamp(metric, time)
	}
}

func (s *Sensor[T]) incrementMeterCountersAndReportActivity(metric string) {
	for _, m := range s.meters {
		m.ReportData()
		m.IncrementCount(metric)
	}
}

func (s *Sensor[T]) decrementMeterCounters(metric string) {
	for _, m := range s.meters {
		m.DecrementCount(metric)
	}
}

func (s *Sensor[T]) decorateCallbacks(options ...types.Option[types.Sensor[T]]) []types.Option[types.Sensor[T]] {

	cbCallbacks := s.decorateCircuitBreakerCallbacks()
	options = append(options, cbCallbacks...)
	httpClientCallbacks := s.decorateHttpClientCallbacks()
	options = append(options, httpClientCallbacks...)
	surgeProtectorCallbacks := s.decorateSurgeProtectorCallbacks()
	options = append(options, surgeProtectorCallbacks...)
	wireCallbacks := s.decorateWireCallbacks()
	options = append(options, wireCallbacks...)
	resisterCallbacks := s.decorateResisterCallbacks()
	options = append(options, resisterCallbacks...)

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
