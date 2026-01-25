package meter

func (m *Meter[T]) initializeMetrics() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, display := range metricDisplayNames {
		m.registerMetricLocked(name, display, metricThresholdType(name))
	}

	m.metricNames = m.metricNames[:0]
	seen := make(map[string]struct{}, len(defaultMetricNames))
	for _, name := range defaultMetricNames {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		m.registerMetricLocked(name, metricDisplay(name), metricThresholdType(name))
		m.metricNames = append(m.metricNames, name)
	}
}
