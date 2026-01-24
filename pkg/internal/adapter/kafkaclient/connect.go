package kafkaclient

import "github.com/joeydtaylor/electrician/pkg/internal/types"

// ConnectSensor attaches sensors to observe Kafka events.
func (a *KafkaClient[T]) ConnectSensor(sensors ...types.Sensor[T]) {
	if len(sensors) == 0 {
		return
	}

	n := 0
	for _, s := range sensors {
		if s != nil {
			sensors[n] = s
			n++
		}
	}
	if n == 0 {
		return
	}
	sensors = sensors[:n]

	a.sensorLock.Lock()
	a.sensors = append(a.sensors, sensors...)
	a.sensorLock.Unlock()

	for _, s := range sensors {
		a.NotifyLoggers(types.DebugLevel, "%s => level: DEBUG, event: ConnectSensor, target: %v",
			a.componentMetadata, s.GetComponentMetadata())
	}
}

// ConnectLogger attaches loggers for adapter events.
func (a *KafkaClient[T]) ConnectLogger(loggers ...types.Logger) {
	if len(loggers) == 0 {
		return
	}

	n := 0
	for _, l := range loggers {
		if l != nil {
			loggers[n] = l
			n++
		}
	}
	if n == 0 {
		return
	}
	loggers = loggers[:n]

	a.loggersLock.Lock()
	a.loggers = append(a.loggers, loggers...)
	total := len(a.loggers)
	a.loggersLock.Unlock()

	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConnectLogger, total_loggers: %d",
		a.componentMetadata, total)
}

// ConnectInput registers wire outputs as inputs to the Kafka writer.
func (a *KafkaClient[T]) ConnectInput(ws ...types.Wire[T]) {
	if len(ws) == 0 {
		return
	}

	n := 0
	for _, w := range ws {
		if w != nil {
			ws[n] = w
			n++
		}
	}
	if n == 0 {
		return
	}
	ws = ws[:n]

	a.inputWires = append(a.inputWires, ws...)
	a.NotifyLoggers(types.InfoLevel, "%s => level: INFO, event: ConnectInput, wires_added: %d, total_wires: %d",
		a.componentMetadata, len(ws), len(a.inputWires))
}
